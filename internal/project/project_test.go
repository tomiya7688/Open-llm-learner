package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndValidateEntrypoints(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "flows"), 0o755); err != nil {
		t.Fatal(err)
	}
	projectJSON := `{
  "format":"open-llm-learner.project",
  "format_version":1,
  "id":"demo",
  "entrypoints":{"default":"flows/main.flow.json"},
  "capabilities":{"test.echo":{"mod":"test.echo","action":"echo"}}
}`
	flowJSON := `{"nodes":[{"id":"start","type":"start"},{"id":"end","type":"end"}],"edges":[]}`
	if err := os.WriteFile(filepath.Join(dir, "project.json"), []byte(projectJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "flows", "main.flow.json"), []byte(flowJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := loaded.ValidateEntrypoints(); err != nil {
		t.Fatal(err)
	}
}

func TestRejectsEscapingFlowPath(t *testing.T) {
	dir := t.TempDir()
	projectJSON := `{"format":"open-llm-learner.project","format_version":1,"id":"demo"}`
	if err := os.WriteFile(filepath.Join(dir, "project.json"), []byte(projectJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := loaded.LoadFlow("../outside.flow.json"); err == nil {
		t.Fatal("expected escaping flow path to fail")
	}
}
