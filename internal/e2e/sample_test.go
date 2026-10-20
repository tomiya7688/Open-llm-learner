package e2e_test

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tomiya7688/Open-llm-learner/internal/flow"
	"github.com/tomiya7688/Open-llm-learner/internal/modpkg"
	"github.com/tomiya7688/Open-llm-learner/internal/modproc"
	"github.com/tomiya7688/Open-llm-learner/internal/modruntime"
	"github.com/tomiya7688/Open-llm-learner/internal/project"
)

func TestV01CoreProofSample(t *testing.T) {
	repoRoot := repositoryRoot(t)
	sourceProject := filepath.Join(repoRoot, "examples", "v0.1-core-proof")
	projectRoot := filepath.Join(t.TempDir(), "v0.1-core-proof")
	copyTree(t, sourceProject, projectRoot)

	buildReferenceMOD(t, repoRoot, "./cmd/oll-mod-echo", filepath.Join(projectRoot, "packages", "mods", "official.echo", "0.1.0", "oll-mod-echo"))
	buildReferenceMOD(t, repoRoot, "./cmd/oll-mod-json", filepath.Join(projectRoot, "packages", "mods", "official.json", "0.1.0", "oll-mod-json"))
	buildReferenceMOD(t, repoRoot, "./cmd/oll-mod-text", filepath.Join(projectRoot, "packages", "mods", "official.text", "0.1.0", "oll-mod-text"))

	loaded, err := project.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := loaded.ValidateEntrypoints(); err != nil {
		t.Fatal(err)
	}

	registry := openRegistry(t, loaded)
	var eventCount atomic.Int32
	engine := &flow.Engine{
		Invoker: registry,
		LoadFlow: func(ctx context.Context, ref string) (*flow.Document, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return loaded.LoadFlow(ref)
		},
		OnEvent: func(event flow.NodeEvent) {
			eventCount.Add(1)
		},
	}

	mainFlow, err := loaded.LoadFlow(loaded.Project.Entrypoints["default"])
	if err != nil {
		t.Fatal(err)
	}
	mainResult, err := engine.Run(context.Background(), mainFlow, map[string]any{"message":"hello"})
	if err != nil {
		t.Fatal(err)
	}
	if mainResult.EndOutputs["success"]["message"] != "hello" {
		t.Fatalf("unexpected main output: %#v", mainResult.EndOutputs)
	}
	if len(mainResult.Children["wrap"]) != 1 || mainResult.Children["wrap"][0].ParentRunID != mainResult.RunID {
		t.Fatalf("subflow parent/child relation missing: %#v", mainResult.Children)
	}
	if eventCount.Load() == 0 {
		t.Fatal("expected at least one MOD event from the delegated call")
	}

	timeoutFlow, err := loaded.LoadFlow(loaded.Project.Entrypoints["timeout"])
	if err != nil {
		t.Fatal(err)
	}
	timeoutResult, err := engine.Run(context.Background(), timeoutFlow, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if timeoutResult.NodeStatus["slow"] != flow.StatusTimedOut {
		t.Fatalf("expected slow node timed_out, got %s", timeoutResult.NodeStatus["slow"])
	}
	timeoutError, ok := timeoutResult.EndOutputs["recovered"]["error"].(map[string]any)
	if !ok || timeoutError["code"] != "TIMEOUT" {
		t.Fatalf("timeout error route missing: %#v", timeoutResult.EndOutputs)
	}

	readFlow, err := loaded.LoadFlow(loaded.Project.Entrypoints["read"])
	if err != nil {
		t.Fatal(err)
	}
	readResult, err := engine.Run(context.Background(), readFlow, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	text, _ := readResult.EndOutputs["end"]["text"].(string)
	if !strings.Contains(text, "v0.1 Core Proof sample") {
		t.Fatalf("text reference MOD did not read the sample README: %q", text)
	}

	closeRegistry(t, registry)

	changed, err := loaded.SetPermission("official.text", "host.filesystem.read", false)
	if err != nil || !changed {
		t.Fatalf("revoke permission failed: changed=%v err=%v", changed, err)
	}
	if err := loaded.Save(); err != nil {
		t.Fatal(err)
	}
	reopened, err := project.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	_, err = modruntime.New(modruntime.Config{
		Project: reopened,
		Resolver: modpkg.Resolver{Roots:modpkg.DefaultRoots(projectRoot, "", "")},
		Stderr: io.Discard,
	})
	var remote *modproc.RemoteError
	if !errors.As(err, &remote) || remote.Code != "PERMISSION_DENIED" {
		t.Fatalf("expected persisted permission denial, got %T %v", err, err)
	}

	changed, err = reopened.SetPermission("official.text", "host.filesystem.read", true)
	if err != nil || !changed {
		t.Fatalf("restore permission failed: changed=%v err=%v", changed, err)
	}
	if err := reopened.Save(); err != nil {
		t.Fatal(err)
	}
	finalProject, err := project.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	restored := openRegistry(t, finalProject)
	closeRegistry(t, restored)
}

func openRegistry(t *testing.T, loaded *project.Loaded) *modruntime.Registry {
	t.Helper()
	registry, err := modruntime.New(modruntime.Config{
		Project: loaded,
		Resolver: modpkg.Resolver{Roots:modpkg.DefaultRoots(loaded.Root, "", "")},
		Stderr: io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func closeRegistry(t *testing.T, registry *modruntime.Registry) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := registry.Close(ctx); err != nil {
		t.Fatal(err)
	}
}

func buildReferenceMOD(t *testing.T, repoRoot, pkg, output string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		output += ".exe"
		updateManifestCommandForWindows(t, filepath.Join(filepath.Dir(output), "mod.json"))
	}
	cmd := exec.Command("go", "build", "-o", output, pkg)
	cmd.Dir = repoRoot
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v\n%s", pkg, err, data)
	}
}

func updateManifestCommandForWindows(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	text = strings.Replace(text, "\"command\": \"oll-mod-echo\"", "\"command\": \"oll-mod-echo.exe\"", 1)
	text = strings.Replace(text, "\"command\": \"oll-mod-json\"", "\"command\": \"oll-mod-json.exe\"", 1)
	text = strings.Replace(text, "\"command\": \"oll-mod-text\"", "\"command\": \"oll-mod-text.exe\"", 1)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return err
		}
		return out.Close()
	})
	if err != nil {
		t.Fatal(err)
	}
}
