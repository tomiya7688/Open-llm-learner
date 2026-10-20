package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tomiya7688/Open-llm-learner/internal/modpkg"
	"github.com/tomiya7688/Open-llm-learner/internal/modproc"
	"github.com/tomiya7688/Open-llm-learner/internal/project"
	"github.com/tomiya7688/Open-llm-learner/internal/protocol"
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
	fmt.Fprintln(os.Stderr, "  oll-core project validate <project-dir>")
	fmt.Fprintln(os.Stderr, "  oll-core mod inspect <mod-package-dir>")
	fmt.Fprintln(os.Stderr, "  oll-core mod invoke --package <dir> --action <name> [--input <json>] [--timeout 30s]")
}

func projectCommand(args []string) error {
	if len(args) != 2 || args[0] != "validate" {
		return fmt.Errorf("usage: oll-core project validate <project-dir>")
	}
	loaded, err := project.Load(args[1])
	if err != nil {
		return err
	}
	if err := loaded.ValidateEntrypoints(); err != nil {
		return err
	}
	fmt.Printf("project %s is valid (%d entrypoints)
", loaded.Project.ID, len(loaded.Project.Entrypoints))
	return nil
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
		fmt.Printf("%s %s protocol=%s concurrency=%s actions=%d
", m.ID, m.Version, m.Protocol, m.Concurrency, len(m.Provides))
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
	inputText := fs.String("input", `{}`, "JSON input")
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
		fmt.Fprintf(os.Stderr, "event %s: %s
", e.Name, string(e.Data))
	})
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
