package modruntime

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tomiya7688/Open-llm-learner/internal/flow"
	"github.com/tomiya7688/Open-llm-learner/internal/modpkg"
	"github.com/tomiya7688/Open-llm-learner/internal/modproc"
	"github.com/tomiya7688/Open-llm-learner/internal/project"
	"github.com/tomiya7688/Open-llm-learner/internal/protocol"
)

func TestRegistryMODToCoreCapabilityCall(t *testing.T) {
	root := t.TempDir()
	modsRoot := filepath.Join(root, "packages", "mods")
	writeHelperPackage(t, modsRoot, "test.a", nil, map[string]modpkg.Action{
		"delegate": {Capability:"test.delegate"},
	})
	writeHelperPackage(t, modsRoot, "test.b", nil, map[string]modpkg.Action{
		"echo": {Capability:"test.echo"},
	})

	loaded := testProject(root, []project.ModSelection{
		{ID:"test.a", Version:"1.0.0"},
		{ID:"test.b", Version:"1.0.0"},
	})
	loaded.Project.Capabilities["test.echo"] = project.CapabilityBinding{Mod:"test.b", Action:"echo"}

	registry, err := New(Config{
		Project:loaded,
		Resolver:modpkg.Resolver{Roots:modpkg.DefaultRoots(root, "", "")},
		Stderr:io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = registry.Close(ctx)
	}()

	out, err := registry.Invoke(context.Background(), flow.InvocationRequest{
		Mod:"test.a",
		Action:"delegate",
		Input:map[string]any{"message":"hello"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	object, ok := out.(map[string]any)
	if !ok || object["message"] != "hello" {
		t.Fatalf("unexpected delegated result: %#v", out)
	}
}

func TestRegistryRejectsMissingPermissionGrant(t *testing.T) {
	root := t.TempDir()
	modsRoot := filepath.Join(root, "packages", "mods")
	writeHelperPackage(t, modsRoot, "test.secure", []string{"host.process"}, map[string]modpkg.Action{
		"run": {Capability:"test.run"},
	})
	loaded := testProject(root, []project.ModSelection{{ID:"test.secure", Version:"1.0.0"}})
	_, err := New(Config{
		Project:loaded,
		Resolver:modpkg.Resolver{Roots:modpkg.DefaultRoots(root, "", "")},
	})
	var remote *modproc.RemoteError
	if !errors.As(err, &remote) || remote.Code != "PERMISSION_DENIED" {
		t.Fatalf("expected PERMISSION_DENIED, got %T %v", err, err)
	}
}

func testProject(root string, mods []project.ModSelection) *project.Loaded {
	return &project.Loaded{
		Root:root,
		Project:project.Project{
			Format:project.Format,
			FormatVersion:project.FormatVersion,
			ID:"test-project",
			Name:"Test Project",
			Entrypoints:map[string]string{},
			Mods:mods,
			Capabilities:map[string]project.CapabilityBinding{},
			Configs:map[string]string{},
			Refs:map[string]string{},
			Schemas:map[string]string{},
			Metadata:map[string]any{},
		},
	}
}

func writeHelperPackage(t *testing.T, root, id string, permissions []string, provides map[string]modpkg.Action) {
	t.Helper()
	dir := filepath.Join(root, id, "1.0.0")
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	src, err := os.Open(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dstPath := filepath.Join(dir, "bin", "mod")
	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		t.Fatal(err)
	}
	if err := dst.Close(); err != nil {
		t.Fatal(err)
	}
	manifest := modpkg.Manifest{
		Format:modpkg.ManifestFormat,
		FormatVersion:modpkg.ManifestFormatVersion,
		ID:id,
		Version:"1.0.0",
		Protocol:protocol.Version,
		Entry:modpkg.Entry{
			Command:"bin/mod",
			Args:[]string{"-test.run=TestRegistryHelperProcess", "--", "--mod-helper="+id},
		},
		Concurrency:"serial",
		Permissions:permissions,
		Provides:provides,
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mod.json"), append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRegistryHelperProcess(t *testing.T) {
	modID := ""
	for _, arg := range os.Args {
		if len(arg) > len("--mod-helper=") && arg[:len("--mod-helper=")] == "--mod-helper=" {
			modID = arg[len("--mod-helper="):]
			break
		}
	}
	if modID == "" {
		return
	}

	dec := protocol.NewDecoder(os.Stdin, protocol.DefaultMaxMessageBytes)
	pending := map[string]string{}
	for {
		var msg protocol.Message
		if err := dec.Decode(&msg); err != nil {
			return
		}
		switch msg.Type {
		case "hello":
			_ = protocol.Encode(os.Stdout, protocol.Message{
				Type:"hello",
				Protocol:protocol.Version,
				Mod:&protocol.ModInfo{ID:modID, Version:"1.0.0"},
			})
		case "invoke":
			if modID == "test.b" && msg.Action == "echo" {
				_ = protocol.Encode(os.Stdout, protocol.Message{Type:"result", ID:msg.ID, Output:msg.Input})
				continue
			}
			if modID == "test.a" && msg.Action == "delegate" {
				callID := "call-" + msg.ID
				pending[callID] = msg.ID
				_ = protocol.Encode(os.Stdout, protocol.Message{
					Type:"call",
					ID:callID,
					Parent:msg.ID,
					Target:&protocol.CallTarget{Capability:"test.echo"},
					Input:msg.Input,
				})
				continue
			}
			_ = protocol.Encode(os.Stdout, protocol.Message{
				Type:"error",
				ID:msg.ID,
				Error:&protocol.ErrorPayload{Code:"ACTION_NOT_FOUND", Message:"unknown action"},
			})
		case "call_result":
			parent := pending[msg.ID]
			delete(pending, msg.ID)
			_ = protocol.Encode(os.Stdout, protocol.Message{Type:"result", ID:parent, Output:msg.Output})
		case "call_error":
			parent := pending[msg.ID]
			delete(pending, msg.ID)
			_ = protocol.Encode(os.Stdout, protocol.Message{Type:"error", ID:parent, Error:msg.Error})
		case "cancel":
			_ = protocol.Encode(os.Stdout, protocol.Message{Type:"cancelled", ID:msg.ID})
		case "shutdown":
			_ = protocol.Encode(os.Stdout, protocol.Message{Type:"shutdown_ack"})
			return
		}
	}
}

func TestParseResourceEstimate(t *testing.T) {
	estimate, err := parseResourceEstimate(map[string]any{
		"estimated_peak_ram_bytes": float64(1024),
		"vram_bytes": float64(2048),
		"disk_cache_bytes": float64(4096),
	})
	if err != nil {
		t.Fatal(err)
	}
	if estimate.RAMBytes != 1024 || estimate.VRAMBytes != 2048 || estimate.DiskCacheBytes != 4096 || !estimate.Heavy {
		t.Fatalf("unexpected estimate: %+v", estimate)
	}
	if _, err := parseResourceEstimate(map[string]any{"ram_bytes": 1.5}); err == nil {
		t.Fatal("expected fractional resource estimate to fail")
	}
}
