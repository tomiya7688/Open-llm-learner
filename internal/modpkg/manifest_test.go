package modpkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifestAndResolveExecutable(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
  "format":"open-llm-learner.mod",
  "format_version":1,
  "id":"test.echo",
  "version":"0.1.0",
  "protocol":"1.0",
  "entry":{"command":"bin/echo"},
  "provides":{"echo":{"capability":"test.echo"}}
}`
	if err := os.WriteFile(filepath.Join(dir, "mod.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Concurrency != "serial" {
		t.Fatalf("expected serial default, got %q", m.Concurrency)
	}
	path, err := m.ExecutablePath(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "bin", "echo")
	if path != want {
		t.Fatalf("got %q want %q", path, want)
	}
}

func TestExecutableCannotEscapePackage(t *testing.T) {
	m := Manifest{Entry: Entry{Command: "../escape"}}
	if _, err := m.ExecutablePath(t.TempDir()); err == nil {
		t.Fatal("expected escaping entry path to fail")
	}
}
