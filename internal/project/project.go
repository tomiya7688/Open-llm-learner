package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tomiya7688/Open-llm-learner/internal/flow"
)

const (
	Format        = "open-llm-learner.project"
	FormatVersion = 1
)

var defaultDirectories = []string{"flows", "configs", "refs", "schemas", "scripts", "packages"}

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
	Name          string                       `json:"name"`
	Entrypoints   map[string]string            `json:"entrypoints"`
	Mods          []ModSelection               `json:"mods"`
	Capabilities  map[string]CapabilityBinding `json:"capabilities,omitempty"`
	Configs       map[string]string            `json:"configs,omitempty"`
	Refs          map[string]string            `json:"refs,omitempty"`
	Schemas       map[string]string            `json:"schemas,omitempty"`
	Metadata      map[string]any                `json:"metadata,omitempty"`
}

type Loaded struct {
	Root    string
	Project Project
}

func Create(root, name string) (*Loaded, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("project name is required")
	}
	if _, err := os.Stat(filepath.Join(absRoot, "project.json")); err == nil {
		return nil, fmt.Errorf("project already exists at %q", absRoot)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err := os.MkdirAll(absRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create project directory: %w", err)
	}
	for _, dir := range defaultDirectories {
		if err := os.MkdirAll(filepath.Join(absRoot, dir), 0o755); err != nil {
			return nil, fmt.Errorf("create project directory %q: %w", dir, err)
		}
	}
	p := Project{
		Format:        Format,
		FormatVersion: FormatVersion,
		ID:            filepath.Base(absRoot),
		Name:          name,
		Entrypoints:   map[string]string{},
		Mods:          []ModSelection{},
		Capabilities:  map[string]CapabilityBinding{},
		Configs:       map[string]string{},
		Refs:          map[string]string{},
		Schemas:       map[string]string{},
		Metadata:      map[string]any{},
	}
	loaded := &Loaded{Root: absRoot, Project: p}
	if err := loaded.Save(); err != nil {
		return nil, err
	}
	return loaded, nil
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

func (l *Loaded) Save() error {
	if l == nil {
		return errors.New("project is nil")
	}
	if err := l.Project.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(l.Root, 0o755); err != nil {
		return fmt.Errorf("create project root: %w", err)
	}
	tmp, err := os.CreateTemp(l.Root, ".project.json.*.tmp")
	if err != nil {
		return fmt.Errorf("create project temp file: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}
	if err := tmp.Chmod(0o644); err != nil {
		cleanup()
		return err
	}
	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&l.Project); err != nil {
		cleanup()
		return fmt.Errorf("encode project.json: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync project.json temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close project.json temp file: %w", err)
	}
	if err := os.Rename(tmpName, filepath.Join(l.Root, "project.json")); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("replace project.json: %w", err)
	}
	return nil
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
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("project name is required")
	}
	if p.Entrypoints == nil {
		return errors.New("project entrypoints is required")
	}
	if p.Mods == nil {
		return errors.New("project mods is required")
	}
	for name, rel := range p.Entrypoints {
		if strings.TrimSpace(name) == "" {
			return errors.New("entrypoint name cannot be empty")
		}
		if !strings.HasSuffix(rel, ".flow.json") {
			return fmt.Errorf("entrypoint %q must reference a .flow.json file", name)
		}
	}
	seenMods := map[string]struct{}{}
	for _, mod := range p.Mods {
		if strings.TrimSpace(mod.ID) == "" || strings.TrimSpace(mod.Version) == "" {
			return errors.New("project MOD selection requires id and version")
		}
		if _, ok := seenMods[mod.ID]; ok {
			return fmt.Errorf("duplicate project MOD selection %q", mod.ID)
		}
		seenMods[mod.ID] = struct{}{}
		seenPermissions := map[string]struct{}{}
		for _, permission := range mod.Permissions {
			if strings.TrimSpace(permission) == "" {
				return fmt.Errorf("MOD %q has an empty permission grant", mod.ID)
			}
			if _, ok := seenPermissions[permission]; ok {
				return fmt.Errorf("MOD %q has duplicate permission grant %q", mod.ID, permission)
			}
			seenPermissions[permission] = struct{}{}
		}
	}
	for capability, binding := range p.Capabilities {
		if strings.TrimSpace(binding.Mod) == "" || strings.TrimSpace(binding.Action) == "" {
			return fmt.Errorf("capability %q must bind both mod and action", capability)
		}
	}
	return nil
}

func (l *Loaded) LoadFlow(rel string) (*flow.Document, error) {
	if !strings.HasSuffix(rel, ".flow.json") {
		return nil, errors.New("flow path must end in .flow.json")
	}
	path, err := l.ResolvePath(rel)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read flow %q: %w", rel, err)
	}
	f, err := flow.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("load flow %q: %w", rel, err)
	}
	return f, nil
}

func (l *Loaded) ValidateEntrypoints() error {
	for name, rel := range l.Project.Entrypoints {
		if _, err := l.LoadFlow(rel); err != nil {
			return fmt.Errorf("entrypoint %q: %w", name, err)
		}
	}
	return nil
}

func (l *Loaded) SetPermission(modID, permission string, allowed bool) (bool, error) {
	if strings.TrimSpace(modID) == "" || strings.TrimSpace(permission) == "" {
		return false, errors.New("mod id and permission are required")
	}
	for i := range l.Project.Mods {
		if l.Project.Mods[i].ID != modID {
			continue
		}
		grants := l.Project.Mods[i].Permissions
		index := -1
		for j, grant := range grants {
			if grant == permission {
				index = j
				break
			}
		}
		if allowed {
			if index >= 0 {
				return false, nil
			}
			grants = append(grants, permission)
			sort.Strings(grants)
			l.Project.Mods[i].Permissions = grants
			return true, nil
		}
		if index < 0 {
			return false, nil
		}
		l.Project.Mods[i].Permissions = append(grants[:index], grants[index+1:]...)
		return true, nil
	}
	return false, fmt.Errorf("project does not select MOD %q", modID)
}

func (l *Loaded) ResolvePath(rel string) (string, error) {
	return safeProjectPath(l.Root, rel)
}

func safeProjectPath(root, rel string) (string, error) {
	if strings.Contains(rel, "://") {
		return "", errors.New("project path must not be a URI")
	}
	if filepath.IsAbs(rel) {
		return "", errors.New("project reference must be relative")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	rootCanonical, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	candidateLexical, err := filepath.Abs(filepath.Join(rootAbs, filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	if !within(rootAbs, candidateLexical) {
		return "", errors.New("project reference escapes project directory")
	}
	candidateCanonical, err := resolveExistingPrefix(candidateLexical)
	if err != nil {
		return "", err
	}
	if !within(rootCanonical, candidateCanonical) {
		return "", errors.New("project reference resolves outside project directory")
	}
	return candidateCanonical, nil
}

func resolveExistingPrefix(path string) (string, error) {
	current := path
	var suffix []string
	for {
		_, err := os.Lstat(current)
		if err == nil {
			resolved, err := filepath.EvalSymlinks(current)
			if err != nil {
				return "", err
			}
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return filepath.Clean(resolved), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("no existing ancestor for %q", path)
		}
		suffix = append(suffix, filepath.Base(current))
		current = parent
	}
}

func within(root, candidate string) bool {
	if candidate == root {
		return true
	}
	return strings.HasPrefix(candidate, root+string(os.PathSeparator))
}
