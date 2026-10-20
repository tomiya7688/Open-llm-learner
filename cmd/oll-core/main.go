package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tomiya7688/Open-llm-learner/internal/flow"
	"github.com/tomiya7688/Open-llm-learner/internal/modpkg"
	"github.com/tomiya7688/Open-llm-learner/internal/modproc"
	"github.com/tomiya7688/Open-llm-learner/internal/modruntime"
	"github.com/tomiya7688/Open-llm-learner/internal/project"
	"github.com/tomiya7688/Open-llm-learner/internal/protocol"
	"github.com/tomiya7688/Open-llm-learner/internal/resource"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "project":
		err = projectCommand(os.Args[2:])
	case "flow":
		err = flowCommand(os.Args[2:])
	case "mod":
		err = modCommand(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Open LLM Learner Core proof CLI")
	fmt.Fprintln(os.Stderr, "  oll-core project create --name <name> <project-dir>")
	fmt.Fprintln(os.Stderr, "  oll-core project validate <project-dir>")
	fmt.Fprintln(os.Stderr, "  oll-core project permission <grant|revoke> <project-dir> <mod-id> <permission>")
	fmt.Fprintln(os.Stderr, "  oll-core flow run --project <dir> [--entry default] [--input <json>] [--user-mods <dir>] [--official-mods <dir>]")
	fmt.Fprintln(os.Stderr, "  oll-core mod inspect <mod-package-dir>")
	fmt.Fprintln(os.Stderr, "  oll-core mod invoke --package <dir> --action <name> [--input <json>] [--timeout 30s]")
}

func projectCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing project subcommand")
	}
	switch args[0] {
	case "create":
		fs := flag.NewFlagSet("project create", flag.ContinueOnError)
		name := fs.String("name", "", "project display name")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *name == "" || fs.NArg() != 1 {
			return fmt.Errorf("usage: oll-core project create --name <name> <project-dir>")
		}
		loaded, err := project.Create(fs.Arg(0), *name)
		if err != nil {
			return err
		}
		fmt.Printf("created project %s at %s\n", loaded.Project.ID, loaded.Root)
		return nil
	case "validate":
		if len(args) != 2 {
			return fmt.Errorf("usage: oll-core project validate <project-dir>")
		}
		loaded, err := project.Load(args[1])
		if err != nil {
			return err
		}
		if err := loaded.ValidateEntrypoints(); err != nil {
			return err
		}
		fmt.Printf("project %s is valid (%d entrypoints)\n", loaded.Project.ID, len(loaded.Project.Entrypoints))
		return nil
	case "permission":
		if len(args) != 5 || (args[1] != "grant" && args[1] != "revoke") {
			return fmt.Errorf("usage: oll-core project permission <grant|revoke> <project-dir> <mod-id> <permission>")
		}
		loaded, err := project.Load(args[2])
		if err != nil {
			return err
		}
		allowed := args[1] == "grant"
		changed, err := loaded.SetPermission(args[3], args[4], allowed)
		if err != nil {
			return err
		}
		if changed {
			if err := loaded.Save(); err != nil {
				return err
			}
		}
		state := "revoked"
		if allowed {
			state = "granted"
		}
		fmt.Printf("%s %s for %s (changed=%v)\n", state, args[4], args[3], changed)
		return nil
	default:
		return fmt.Errorf("unknown project subcommand %q", args[0])
	}
}

func flowCommand(args []string) error {
	if len(args) == 0 || args[0] != "run" {
		return fmt.Errorf("usage: oll-core flow run --project <dir> [options]")
	}
	fs := flag.NewFlagSet("flow run", flag.ContinueOnError)
	projectDir := fs.String("project", "", "project directory")
	entry := fs.String("entry", "default", "project entrypoint name")
	inputText := fs.String("input", "{}", "flow input JSON object")
	userMods := fs.String("user-mods", "", "user-managed MOD root")
	officialMods := fs.String("official-mods", "", "official bundled MOD root")
	allowUnknownVRAM := fs.Bool("allow-unknown-vram", false, "allow VRAM-estimated work when host VRAM cannot be measured")
	timeout := fs.Duration("timeout", 0, "optional whole-flow timeout")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if *projectDir == "" {
		return fmt.Errorf("--project is required")
	}
	loaded, err := project.Load(*projectDir)
	if err != nil {
		return err
	}
	flowPath, ok := loaded.Project.Entrypoints[*entry]
	if !ok {
		return fmt.Errorf("project entrypoint %q not found", *entry)
	}
	doc, err := loaded.LoadFlow(flowPath)
	if err != nil {
		return err
	}
	var input map[string]any
	if err := json.Unmarshal([]byte(*inputText), &input); err != nil {
		return fmt.Errorf("invalid --input JSON object: %w", err)
	}
	resolver := modpkg.Resolver{Roots:modpkg.DefaultRoots(loaded.Root, *userMods, *officialMods)}
	governor, err := resource.NewGovernor(resource.GovernorConfig{
		Policy: resource.Policy{},
		Sampler: resource.NewHostSampler(),
		AllowUnknownVRAM: *allowUnknownVRAM,
	})
	if err != nil {
		return err
	}
	registry, err := modruntime.New(modruntime.Config{Project:loaded, Resolver:resolver, Stderr:os.Stderr, Governor:governor})
	if err != nil {
		return err
	}
	defer func() {
		if err := modruntime.ShutdownWithDefaultTimeout(registry); err != nil {
			fmt.Fprintln(os.Stderr, "shutdown warning:", err)
		}
	}()

	runCtx := context.Background()
	cancel := func() {}
	if *timeout > 0 {
		runCtx, cancel = context.WithTimeout(runCtx, *timeout)
	}
	defer cancel()

	engine := &flow.Engine{
		Invoker: registry,
		OnEvent: func(event flow.NodeEvent) {
			payload, _ := json.Marshal(event.Event.Data)
			fmt.Fprintf(os.Stderr, "node %s event %s: %s\n", event.NodeID, event.Event.Name, payload)
		},
	}
	result, err := engine.Run(runCtx, doc, input)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func modCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing mod subcommand")
	}
	switch args[0] {
	case "inspect":
		if len(args) != 2 {
			return fmt.Errorf("usage: oll-core mod inspect <mod-package-dir>")
		}
		m, err := modpkg.Load(args[1])
		if err != nil {
			return err
		}
		fmt.Printf("%s %s protocol=%s concurrency=%s actions=%d\n", m.ID, m.Version, m.Protocol, m.Concurrency, len(m.Provides))
		return nil
	case "invoke":
		return invokeCommand(args[1:])
	default:
		return fmt.Errorf("unknown mod subcommand %q", args[0])
	}
}

func invokeCommand(args []string) error {
	fs := flag.NewFlagSet("mod invoke", flag.ContinueOnError)
	packageDir := fs.String("package", "", "MOD package directory")
	action := fs.String("action", "", "action name")
	inputText := fs.String("input", "{}", "JSON input")
	timeout := fs.Duration("timeout", 30*time.Second, "invocation timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *packageDir == "" || *action == "" {
		return fmt.Errorf("--package and --action are required")
	}
	m, err := modpkg.Load(*packageDir)
	if err != nil {
		return err
	}
	executable, err := m.ExecutablePath(*packageDir)
	if err != nil {
		return err
	}
	var input any
	if err := json.Unmarshal([]byte(*inputText), &input); err != nil {
		return fmt.Errorf("invalid --input JSON: %w", err)
	}
	client, err := modproc.Start(context.Background(), modproc.Config{
		Executable:         executable,
		Args:               m.Entry.Args,
		WorkDir:            *packageDir,
		ExpectedModID:      m.ID,
		ExpectedModVersion: m.Version,
		SupportedProtocols: []string{m.Protocol},
		Stderr:             os.Stderr,
	})
	if err != nil {
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = client.Shutdown(ctx)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	out, err := client.Invoke(ctx, *action, input, func(e protocol.Event) {
		fmt.Fprintf(os.Stderr, "event %s: %s\n", e.Name, string(e.Data))
	})
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
