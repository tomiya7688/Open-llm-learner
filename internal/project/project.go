package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	Format        = "open-llm-learner.project"
	FormatVersion = 1
)

type ModSelection struct {
	ID          string   `json:"id"`
	Version     string   `json:"version"`
	Digest      string   `json:"digest,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type CapabilityBinding struct {
	Mod    string `json:"mod"`
	Action string `json:"action"`
}

type Project struct {
	Format        string                       `json:"format"`
	FormatVersion int                          `json:"format_version"`
	ID            string                       `json:"id"`
	Name          string                       `json:"name,omitempty"`
	Entrypoints   map[string]string            `json:"entrypoints,omitempty"`
	Mods          []ModSelection               `json:"mods,omitempty"`
	Capabilities  map[string]CapabilityBinding `json:"capabilities,omitempty"`
	Configs       map[string]string            `json:"configs,omitempty"`
	Refs          map[string]string            `json:"refs,omitempty"`
	Schemas       map[string]string            `json:"schemas,omitempty"`
	Metadata      map[string]any                `json:"metadata,omitempty"`
}

type FlowDocument struct {
	Format        string           `json:"format,omitempty"`
	FormatVersion int              `json:"format_version,omitempty"`
	ID            string           `json:"id,omitempty"`
	Name          string           `json:"name,omitempty"`
	Inputs        map[string]any   `json:"inputs,omitempty"`
	Outputs       map[string]any   `json:"outputs,omitempty"`
	Nodes         []map[string]any `json:"nodes,omitempty"`
	Edges         []map[string]any `json:"edges,omitempty"`
	Metadata      map[string]any   `json:"metadata,omitempty"`
}

type Loaded struct {
	Root    string
	Project Project
}

func Load(root string) (*Loaded, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(absRoot, "project.json"))
	if err != nil {
		return nil, fmt.Errorf("read project.json: %w", err)
	}
	var p Project
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("parse project.json: %w", err)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &Loaded{Root: absRoot, Project: p}, nil
}

func (p *Project) Validate() error {
	if p.Format != Format {
		return fmt.Errorf("unsupported project format %q", p.Format)
	}
	if p.FormatVersion != FormatVersion {
		return fmt.Errorf("unsupported project format_version %d", p.FormatVersion)
	}
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("project id is required")
	}
	for name, rel := range p.Entrypoints {
		if !strings.HasSuffix(rel, ".flow.json") {
			return fmt.Errorf("entrypoint %q must reference a .flow.json file", name)
		}
	}
	for capability, binding := range p.Capabilities {
		if strings.TrimSpace(binding.Mod) == "" || strings.TrimSpace(binding.Action) == "" {
			return fmt.Errorf("capability %q must bind both mod and action", capability)
		}
	}
	return nil
}

func (l *Loaded) LoadFlow(rel string) (*FlowDocument, error) {
	if !strings.HasSuffix(rel, ".flow.json") {
		return nil, errors.New("flow path must end in .flow.json")
	}
	path, err := safeProjectPath(l.Root, rel)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read flow %q: %w", rel, err)
	}
	var f FlowDocument
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("parse flow %q: %w", rel, err)
	}
	return &f, nil
}

func (l *Loaded) ValidateEntrypoints() error {
	for name, rel := range l.Project.Entrypoints {
		if _, err := l.LoadFlow(rel); err != nil {
			return fmt.Errorf("entrypoint %q: %w", name, err)
		}
	}
	return nil
}

func safeProjectPath(root, rel string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", errors.New("project reference must be relative")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	candidate, err := filepath.Abs(filepath.Join(rootAbs, filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	if candidate != rootAbs && !strings.HasPrefix(candidate, rootAbs+string(os.PathSeparator)) {
		return "", errors.New("project reference escapes project directory")
	}
	return candidate, nil
}
