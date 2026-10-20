package modruntime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tomiya7688/Open-llm-learner/internal/flow"
	"github.com/tomiya7688/Open-llm-learner/internal/modpkg"
	"github.com/tomiya7688/Open-llm-learner/internal/modproc"
	"github.com/tomiya7688/Open-llm-learner/internal/permission"
	"github.com/tomiya7688/Open-llm-learner/internal/project"
	"github.com/tomiya7688/Open-llm-learner/internal/protocol"
	"github.com/tomiya7688/Open-llm-learner/internal/resource"
)

type Config struct {
	Project  *project.Loaded
	Resolver modpkg.Resolver
	Stderr   io.Writer
	Governor *resource.Governor
}

type Registry struct {
	project  *project.Loaded
	stderr   io.Writer
	governor *resource.Governor

	mu      sync.Mutex
	entries map[string]*entry
	counter atomic.Uint64
}

type entry struct {
	selection project.ModSelection
	resolved  *modpkg.Resolved

	clientMu sync.Mutex
	client   *modproc.Client

	serialMu sync.Mutex
}

func New(cfg Config) (*Registry, error) {
	if cfg.Project == nil {
		return nil, errors.New("project is required")
	}
	if err := cfg.Project.Project.Validate(); err != nil {
		return nil, err
	}
	r := &Registry{
		project: cfg.Project,
		stderr: cfg.Stderr,
		governor: cfg.Governor,
		entries: make(map[string]*entry, len(cfg.Project.Project.Mods)),
	}
	for _, selection := range cfg.Project.Project.Mods {
		resolved, err := cfg.Resolver.Resolve(selection.ID, selection.Version, selection.Digest)
		if err != nil {
			return nil, fmt.Errorf("resolve project MOD %s: %w", selection.ID, err)
		}
		if err := permission.CheckManifest(resolved.Package.Manifest, selection.Permissions); err != nil {
			var denied *permission.DeniedError
			if errors.As(err, &denied) {
				return nil, &modproc.RemoteError{
					Code: "PERMISSION_DENIED",
					Message: denied.Error(),
					Details: map[string]any{"mod": selection.ID, "missing": denied.Missing},
				}
			}
			return nil, err
		}
		r.entries[selection.ID] = &entry{selection: selection, resolved: resolved}
	}
	for capability, binding := range cfg.Project.Project.Capabilities {
		target := r.entries[binding.Mod]
		if target == nil {
			return nil, fmt.Errorf("capability %q binds unselected MOD %q", capability, binding.Mod)
		}
		if _, ok := target.resolved.Package.Manifest.Provides[binding.Action]; !ok {
			return nil, fmt.Errorf("capability %q binds unknown action %s.%s", capability, binding.Mod, binding.Action)
		}
	}
	return r, nil
}

func (r *Registry) Invoke(ctx context.Context, req flow.InvocationRequest, onEvent func(flow.InvocationEvent)) (any, error) {
	modID, action, err := r.resolveTarget(req.Mod, req.Action, req.Capability)
	if err != nil {
		return nil, err
	}
	return r.invokeTarget(ctx, modID, action, req.Input, onEvent)
}

func (r *Registry) resolveTarget(modID, action, capability string) (string, string, error) {
	if capability != "" {
		binding, ok := r.project.Project.Capabilities[capability]
		if !ok {
			return "", "", &modproc.RemoteError{Code:"CAPABILITY_NOT_FOUND", Message:fmt.Sprintf("capability %q is not bound in the project", capability)}
		}
		return binding.Mod, binding.Action, nil
	}
	if modID == "" || action == "" {
		return "", "", &modproc.RemoteError{Code:"INVALID_REQUEST", Message:"direct MOD invocation requires mod and action"}
	}
	return modID, action, nil
}

func (r *Registry) invokeTarget(ctx context.Context, modID, action string, input any, onEvent func(flow.InvocationEvent)) (any, error) {
	target := r.entries[modID]
	if target == nil {
		return nil, &modproc.RemoteError{Code:"CAPABILITY_NOT_FOUND", Message:fmt.Sprintf("MOD %q is not selected by the project", modID)}
	}
	if _, ok := target.resolved.Package.Manifest.Provides[action]; !ok {
		return nil, &modproc.RemoteError{Code:"ACTION_NOT_FOUND", Message:fmt.Sprintf("MOD %s does not provide action %q", modID, action)}
	}
	client, err := r.clientFor(ctx, modID, target)
	if err != nil {
		return nil, err
	}

	estimate, err := r.estimateForAction(ctx, target, client, action, input)
	if err != nil {
		return nil, err
	}
	var lease *resource.Lease
	if r.governor != nil && estimate != nil {
		lease, err = r.governor.Admit(ctx, resource.Request{
			ID: fmt.Sprintf("%s/%s/%d", modID, action, r.counter.Add(1)),
			Estimate: *estimate,
		})
		if err != nil {
			return nil, err
		}
		defer lease.Release()
	}

	if target.resolved.Package.Manifest.Concurrency == "serial" {
		target.serialMu.Lock()
		defer target.serialMu.Unlock()
	}
	raw, err := client.Invoke(ctx, action, input, func(event protocol.Event) {
		if onEvent == nil {
			return
		}
		var data any
		if len(event.Data) > 0 {
			if err := json.Unmarshal(event.Data, &data); err != nil {
				data = string(event.Data)
			}
		}
		onEvent(flow.InvocationEvent{Name:event.Name, Data:data})
	})
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	var output any
	if err := json.Unmarshal(raw, &output); err != nil {
		return nil, &modproc.RemoteError{Code:"PROTOCOL_ERROR", Message:fmt.Sprintf("MOD %s returned invalid JSON result: %v", modID, err)}
	}
	return output, nil
}

func (r *Registry) estimateForAction(ctx context.Context, target *entry, client *modproc.Client, action string, input any) (*resource.Estimate, error) {
	if r.governor == nil {
		return nil, nil
	}
	estimateAction := ""
	for name, provided := range target.resolved.Package.Manifest.Provides {
		if provided.Capability == "resource.estimate" {
			estimateAction = name
			break
		}
	}
	if estimateAction == "" || action == estimateAction {
		return nil, nil
	}

	invoke := func() (json.RawMessage, error) {
		return client.Invoke(ctx, estimateAction, map[string]any{"action":action, "input":input}, nil)
	}
	var raw json.RawMessage
	var err error
	if target.resolved.Package.Manifest.Concurrency == "serial" {
		target.serialMu.Lock()
		raw, err = invoke()
		target.serialMu.Unlock()
	} else {
		raw, err = invoke()
	}
	if err != nil {
		return nil, fmt.Errorf("resource estimate for %s.%s failed: %w", target.resolved.Package.Manifest.ID, action, err)
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, &modproc.RemoteError{Code:"PROTOCOL_ERROR", Message:"resource.estimate returned invalid JSON"}
	}
	estimate, err := parseResourceEstimate(object)
	if err != nil {
		return nil, &modproc.RemoteError{Code:"PROTOCOL_ERROR", Message:err.Error()}
	}
	return &estimate, nil
}

func parseResourceEstimate(object map[string]any) (resource.Estimate, error) {
	var estimate resource.Estimate
	var err error
	if estimate.RAMBytes, err = uintField(object, "estimated_peak_ram_bytes", "ram_bytes"); err != nil {
		return estimate, err
	}
	if estimate.VRAMBytes, err = uintField(object, "estimated_peak_vram_bytes", "vram_bytes"); err != nil {
		return estimate, err
	}
	if estimate.DiskCacheBytes, err = uintField(object, "estimated_peak_disk_cache_bytes", "disk_cache_bytes"); err != nil {
		return estimate, err
	}
	if heavy, ok := object["heavy"].(bool); ok {
		estimate.Heavy = heavy
	} else {
		estimate.Heavy = estimate.RAMBytes > 0 || estimate.VRAMBytes > 0 || estimate.DiskCacheBytes > 0
	}
	return estimate, nil
}

func uintField(object map[string]any, names ...string) (uint64, error) {
	for _, name := range names {
		value, ok := object[name]
		if !ok || value == nil {
			continue
		}
		number, ok := value.(float64)
		if !ok || number < 0 || number != float64(uint64(number)) {
			return 0, fmt.Errorf("resource estimate field %q must be a non-negative integer", name)
		}
		return uint64(number), nil
	}
	return 0, nil
}

func (r *Registry) clientFor(ctx context.Context, modID string, target *entry) (*modproc.Client, error) {
	target.clientMu.Lock()
	defer target.clientMu.Unlock()
	if target.client != nil {
		return target.client, nil
	}
	manifest := target.resolved.Package.Manifest
	executable, err := manifest.ExecutablePath(target.resolved.Package.Dir)
	if err != nil {
		return nil, err
	}
	client, err := modproc.Start(ctx, modproc.Config{
		Executable: executable,
		Args: manifest.Entry.Args,
		WorkDir: target.resolved.Package.Dir,
		ExpectedModID: manifest.ID,
		ExpectedModVersion: manifest.Version,
		SupportedProtocols: []string{manifest.Protocol},
		Stderr: r.stderr,
		CallHandler: func(callCtx context.Context, call protocol.CallRequest) (any, error) {
			return r.handleCall(callCtx, modID, call)
		},
	})
	if err != nil {
		return nil, err
	}
	target.client = client
	return client, nil
}

func (r *Registry) handleCall(ctx context.Context, callerMod string, call protocol.CallRequest) (any, error) {
	var input any = map[string]any{}
	if len(call.Input) > 0 {
		if err := json.Unmarshal(call.Input, &input); err != nil {
			return nil, &modproc.RemoteError{Code:"INVALID_REQUEST", Message:"MOD-to-Core call input is not valid JSON"}
		}
	}
	modID := ""
	action := ""
	if call.Target.Capability != "" {
		var err error
		modID, action, err = r.resolveTarget("", "", call.Target.Capability)
		if err != nil {
			return nil, err
		}
	} else {
		modID = call.Target.Mod
		action = call.Target.Action
		if modID == "" || action == "" {
			return nil, &modproc.RemoteError{Code:"INVALID_REQUEST", Message:"MOD-to-Core direct call requires mod and action"}
		}
	}
	if modID == callerMod {
		target := r.entries[modID]
		if target != nil && target.resolved.Package.Manifest.Concurrency == "serial" {
			return nil, &modproc.RemoteError{Code:"INVALID_REQUEST", Message:"serial MOD cannot synchronously call itself"}
		}
	}
	return r.invokeTarget(ctx, modID, action, input, nil)
}

func (r *Registry) Close(ctx context.Context) error {
	r.mu.Lock()
	entries := make([]*entry, 0, len(r.entries))
	for _, item := range r.entries {
		entries = append(entries, item)
	}
	r.mu.Unlock()

	var firstErr error
	for _, item := range entries {
		item.clientMu.Lock()
		client := item.client
		item.client = nil
		item.clientMu.Unlock()
		if client == nil {
			continue
		}
		if err := client.Shutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
			_ = client.ForceStop()
		}
	}
	return firstErr
}

func (r *Registry) ForceStop() {
	r.mu.Lock()
	entries := make([]*entry, 0, len(r.entries))
	for _, item := range r.entries {
		entries = append(entries, item)
	}
	r.mu.Unlock()
	for _, item := range entries {
		item.clientMu.Lock()
		client := item.client
		item.clientMu.Unlock()
		if client != nil {
			_ = client.ForceStop()
		}
	}
}

func ShutdownWithDefaultTimeout(r *Registry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.Close(ctx)
}
