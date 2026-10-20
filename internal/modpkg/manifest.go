package modpkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ManifestFormat        = "open-llm-learner.mod"
	ManifestFormatVersion = 1
)

type Entry struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

type Action struct {
	Capability   string `json:"capability,omitempty"`
	InputSchema  string `json:"input_schema,omitempty"`
	OutputSchema string `json:"output_schema,omitempty"`
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
}

type Dependency struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type Manifest struct {
	Format        string            `json:"format"`
	FormatVersion int               `json:"format_version"`
	ID            string            `json:"id"`
	Name          string            `json:"name,omitempty"`
	Version       string            `json:"version"`
	Protocol      string            `json:"protocol"`
	Entry         Entry             `json:"entry"`
	Concurrency   string            `json:"concurrency,omitempty"`
	Permissions   []string          `json:"permissions,omitempty"`
	Provides      map[string]Action `json:"provides"`
	Dependencies  []Dependency      `json:"dependencies,omitempty"`
}

func Load(dir string) (*Manifest, error) {
	b, err := os.ReadFile(filepath.Join(dir, "mod.json"))
	if err != nil {
		return nil, fmt.Errorf("read mod.json: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("parse mod.json: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Manifest) Validate() error {
	if m.Format != ManifestFormat {
		return fmt.Errorf("unsupported MOD manifest format %q", m.Format)
	}
	if m.FormatVersion != ManifestFormatVersion {
		return fmt.Errorf("unsupported MOD manifest format_version %d", m.FormatVersion)
	}
	if strings.TrimSpace(m.ID) == "" {
		return errors.New("MOD manifest id is required")
	}
	if strings.TrimSpace(m.Version) == "" {
		return errors.New("MOD manifest version is required")
	}
	if strings.TrimSpace(m.Protocol) == "" {
		return errors.New("MOD manifest protocol is required")
	}
	if strings.TrimSpace(m.Entry.Command) == "" {
		return errors.New("MOD manifest entry.command is required")
	}
	if len(m.Provides) == 0 {
		return errors.New("MOD manifest provides must contain at least one action")
	}
	if m.Concurrency == "" {
		m.Concurrency = "serial"
	}
	if m.Concurrency != "serial" && m.Concurrency != "parallel" {
		return fmt.Errorf("invalid MOD concurrency %q", m.Concurrency)
	}
	return nil
}

func (m *Manifest) ExecutablePath(packageDir string) (string, error) {
	if strings.HasPrefix(m.Entry.Command, "mod://") {
		return "", fmt.Errorf("runner-backed entry %q requires MOD dependency resolution", m.Entry.Command)
	}
	if filepath.IsAbs(m.Entry.Command) {
		return "", errors.New("MOD entry.command must be package-relative")
	}
	root, err := filepath.Abs(packageDir)
	if err != nil {
		return "", err
	}
	candidate, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(m.Entry.Command)))
	if err != nil {
		return "", err
	}
	if candidate != root && !strings.HasPrefix(candidate, root+string(os.PathSeparator)) {
		return "", errors.New("MOD entry.command escapes package directory")
	}
	return candidate, nil
}
