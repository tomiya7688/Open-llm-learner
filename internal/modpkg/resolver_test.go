package modpkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchVersion(t *testing.T) {
	tests := []struct {
		version, expr string
		want          bool
	}{
		{"1.2.3", "1.2.3", true},
		{"1.2.3", "*", true},
		{"1.2.3", ">=1.0.0 <2.0.0", true},
		{"2.0.0", ">=1.0.0 <2.0.0", false},
		{"1.3.0-beta.1", ">=1.2.0 <2.0.0", true},
	}
	for _, tt := range tests {
		got, err := MatchVersion(tt.version, tt.expr)
		if err != nil {
			t.Fatalf("%s %s: %v", tt.version, tt.expr, err)
		}
		if got != tt.want {
			t.Fatalf("MatchVersion(%q,%q)=%v want %v", tt.version, tt.expr, got, tt.want)
		}
	}
}

func TestResolverPrefersProjectAndHighestCompatibleVersion(t *testing.T) {
	base := t.TempDir()
	projectRoot := filepath.Join(base, "project", "packages", "mods")
	userRoot := filepath.Join(base, "user")
	writeTestPackage(t, projectRoot, "test.echo", "1.1.0", nil)
	writeTestPackage(t, projectRoot, "test.echo", "1.2.0", nil)
	writeTestPackage(t, userRoot, "test.echo", "1.9.0", nil)

	r := Resolver{Roots: []SearchRoot{{Name:"project",Path:projectRoot},{Name:"user",Path:userRoot}}}
	got, err := r.Resolve("test.echo", ">=1.0.0 <2.0.0", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Package.Root.Name != "project" || got.Package.Manifest.Version != "1.2.0" {
		t.Fatalf("unexpected resolution: root=%s version=%s", got.Package.Root.Name, got.Package.Manifest.Version)
	}
}

func TestResolverDoesNotFallBackPastBrokenCompatiblePackage(t *testing.T) {
	base := t.TempDir()
	projectRoot := filepath.Join(base, "project")
	userRoot := filepath.Join(base, "user")
	brokenDir := filepath.Join(projectRoot, "test.echo", "1.2.0")
	if err := os.MkdirAll(brokenDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(brokenDir, "mod.json"), []byte(`{"broken":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	writeTestPackage(t, userRoot, "test.echo", "1.1.0", nil)

	r := Resolver{Roots: []SearchRoot{{Name:"project",Path:projectRoot},{Name:"user",Path:userRoot}}}
	if _, err := r.Resolve("test.echo", ">=1.0.0 <2.0.0", ""); err == nil {
		t.Fatal("expected broken project-local compatible package to stop resolution")
	}
}

func TestResolverChecksDigestAndDirectDependencies(t *testing.T) {
	root := t.TempDir()
	deps := []Dependency{{ID:"test.dep", Version:">=2.0.0 <3.0.0"}}
	mainDir := writeTestPackage(t, root, "test.main", "1.0.0", deps)
	writeTestPackage(t, root, "test.dep", "2.1.0", nil)
	digest, err := PackageDigest(mainDir)
	if err != nil {
		t.Fatal(err)
	}

	r := Resolver{Roots: []SearchRoot{{Name:"project",Path:root}}}
	got, err := r.Resolve("test.main", "1.0.0", digest)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dependencies["test.dep"].Manifest.Version != "2.1.0" {
		t.Fatalf("dependency not resolved: %+v", got.Dependencies)
	}
	if _, err := r.Resolve("test.main", "1.0.0", "sha256:deadbeef"); err == nil {
		t.Fatal("expected digest mismatch")
	}
}

func TestPackageDigestRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "mod.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "target")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := PackageDigest(root); err == nil {
		t.Fatal("expected symlink rejection")
	}
}

func writeTestPackage(t *testing.T, root, id, version string, deps []Dependency) string {
	t.Helper()
	dir := filepath.Join(root, id, version)
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	depJSON := "[]"
	if len(deps) > 0 {
		depJSON = `[{"id":"`+deps[0].ID+`","version":"`+deps[0].Version+`"}]`
	}
	manifest := `{
 "format":"open-llm-learner.mod",
 "format_version":1,
 "id":"`+id+`",
 "version":"`+version+`",
 "protocol":"1.0",
 "entry":{"command":"bin/mod"},
 "provides":{"echo":{"capability":"test.echo"}},
 "dependencies":`+depJSON+`
}`
	if err := os.WriteFile(filepath.Join(dir, "mod.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bin", "mod"), []byte("test"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}
