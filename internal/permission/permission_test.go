package permission

import (
	"errors"
	"testing"

	"github.com/tomiya7688/Open-llm-learner/internal/modpkg"
)

func TestCheckManifestAllowsExactGrants(t *testing.T) {
	m := &modpkg.Manifest{ID:"official.file", Permissions:[]string{"host.filesystem.read","host.filesystem.write"}}
	if err := CheckManifest(m, []string{"host.filesystem.write","host.filesystem.read","host.process"}); err != nil {
		t.Fatal(err)
	}
}

func TestCheckManifestReportsMissingGrants(t *testing.T) {
	m := &modpkg.Manifest{ID:"official.file", Permissions:[]string{"host.filesystem.read","host.process"}}
	err := CheckManifest(m, []string{"host.filesystem.read"})
	var denied *DeniedError
	if !errors.As(err, &denied) {
		t.Fatalf("expected DeniedError, got %T %v", err, err)
	}
	if len(denied.Missing) != 1 || denied.Missing[0] != "host.process" {
		t.Fatalf("unexpected missing permissions: %v", denied.Missing)
	}
}
