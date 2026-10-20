package project

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCreateSaveLoadAndValidateEntrypoints(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "demo")
	loaded, err := Create(dir, "Demo")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range defaultDirectories {
		if info, err := os.Stat(filepath.Join(dir, name)); err != nil || !info.IsDir() {
			t.Fatalf("default directory %q missing: %v", name, err)
		}
	}

	flowJSON := `{
  "format":"open-llm-learner.flow",
  "format_version":1,
  "id":"main",
  "nodes":[{"id":"start","type":"start"},{"id":"end","type":"end"}],
  "edges":[]
}`
	if err := os.WriteFile(filepath.Join(dir, "flows", "main.flow.json"), []byte(flowJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded.Project.Entrypoints["default"] = "flows/main.flow.json"
	loaded.Project.Capabilities["test.echo"] = CapabilityBinding{Mod:"test.echo", Action:"echo"}
	if err := loaded.Save(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Project.Name != "Demo" {
		t.Fatalf("unexpected name %q", reopened.Project.Name)
	}
	if err := reopened.ValidateEntrypoints(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateDoesNotOverwriteExistingProject(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "demo")
	if _, err := Create(dir, "Demo"); err != nil {
		t.Fatal(err)
	}
	if _, err := Create(dir, "Other"); err == nil {
		t.Fatal("expected existing project to be preserved")
	}
}

func TestRejectsEscapingFlowPath(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "demo")
	loaded, err := Create(dir, "Demo")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := loaded.LoadFlow("../outside.flow.json"); err == nil {
		t.Fatal("expected escaping flow path to fail")
	}
	if _, err := loaded.ResolvePath("file:///tmp/outside"); err == nil {
		t.Fatal("expected URI path to fail")
	}
}

func TestRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation commonly requires additional privileges on Windows")
	}
	base := t.TempDir()
	dir := filepath.Join(base, "demo")
	loaded, err := Create(dir, "Demo")
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(base, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "refs", "escape")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := loaded.ResolvePath("refs/escape/model.ref.json"); err == nil {
		t.Fatal("expected symlink escape to fail")
	}
}
