package modpkg

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var modIDPattern = regexp.MustCompile("^[a-z0-9][a-z0-9._-]*$")

type SearchRoot struct {
	Name string
	Path string
}

type Resolver struct {
	Roots []SearchRoot
}

type Package struct {
	Root     SearchRoot
	Dir      string
	Manifest *Manifest
	Digest   string
}

type Resolved struct {
	Package      Package
	Dependencies map[string]Package
}

func DefaultRoots(projectRoot, userRoot, officialRoot string) []SearchRoot {
	roots := []SearchRoot{{Name: "project", Path: filepath.Join(projectRoot, "packages", "mods")}}
	if userRoot != "" {
		roots = append(roots, SearchRoot{Name: "user", Path: userRoot})
	}
	if officialRoot != "" {
		roots = append(roots, SearchRoot{Name: "official", Path: officialRoot})
	}
	return roots
}

func (r Resolver) Resolve(id, versionConstraint, expectedDigest string) (*Resolved, error) {
	pkg, err := r.resolveOne(id, versionConstraint)
	if err != nil {
		return nil, err
	}
	if expectedDigest != "" && pkg.Digest != expectedDigest {
		return nil, fmt.Errorf("MOD %s %s digest mismatch: expected %s got %s", pkg.Manifest.ID, pkg.Manifest.Version, expectedDigest, pkg.Digest)
	}
	deps := make(map[string]Package, len(pkg.Manifest.Dependencies))
	for _, dep := range pkg.Manifest.Dependencies {
		resolvedDep, err := r.resolveOne(dep.ID, dep.Version)
		if err != nil {
			return nil, fmt.Errorf("resolve dependency %s %q for %s %s: %w", dep.ID, dep.Version, pkg.Manifest.ID, pkg.Manifest.Version, err)
		}
		deps[dep.ID] = resolvedDep
	}
	return &Resolved{Package: pkg, Dependencies: deps}, nil
}

func (r Resolver) resolveOne(id, versionConstraint string) (Package, error) {
	if !modIDPattern.MatchString(id) {
		return Package{}, fmt.Errorf("invalid MOD id %q", id)
	}
	if strings.TrimSpace(versionConstraint) == "" {
		versionConstraint = "*"
	}
	var searched []string
	for _, root := range r.Roots {
		if root.Path == "" {
			continue
		}
		searched = append(searched, root.Path)
		pkg, found, err := resolveFromRoot(root, id, versionConstraint)
		if err != nil {
			return Package{}, err
		}
		if found {
			return pkg, nil
		}
	}
	return Package{}, fmt.Errorf("MOD %s version %q not found in search roots: %s", id, versionConstraint, strings.Join(searched, ", "))
}

func resolveFromRoot(root SearchRoot, id, versionConstraint string) (Package, bool, error) {
	idDir := filepath.Join(root.Path, id)
	entries, err := os.ReadDir(idDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Package{}, false, nil
		}
		return Package{}, false, fmt.Errorf("read MOD directory %q: %w", idDir, err)
	}
	type candidate struct {
		name string
		v    Version
	}
	var candidates []candidate
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		v, err := ParseVersion(entry.Name())
		if err != nil {
			continue
		}
		ok, err := MatchVersion(entry.Name(), versionConstraint)
		if err != nil {
			return Package{}, false, fmt.Errorf("invalid version constraint %q: %w", versionConstraint, err)
		}
		if ok {
			candidates = append(candidates, candidate{name: entry.Name(), v: v})
		}
	}
	if len(candidates) == 0 {
		return Package{}, false, nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].v.Compare(candidates[j].v) > 0
	})
	dir := filepath.Join(idDir, candidates[0].name)
	manifest, err := Load(dir)
	if err != nil {
		return Package{}, false, fmt.Errorf("load compatible MOD package %q: %w", dir, err)
	}
	if manifest.ID != id {
		return Package{}, false, fmt.Errorf("MOD package %q manifest id %q does not match directory id %q", dir, manifest.ID, id)
	}
	if manifest.Version != candidates[0].name {
		return Package{}, false, fmt.Errorf("MOD package %q manifest version %q does not match directory version %q", dir, manifest.Version, candidates[0].name)
	}
	digest, err := PackageDigest(dir)
	if err != nil {
		return Package{}, false, fmt.Errorf("digest MOD package %q: %w", dir, err)
	}
	return Package{Root: root, Dir: dir, Manifest: manifest, Digest: digest}, true, nil
}

func PackageDigest(root string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	var paths []string
	err = filepath.WalkDir(rootAbs, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == rootAbs {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink is not allowed in MOD package digest: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("non-regular file is not allowed in MOD package digest: %s", path)
		}
		rel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(paths)

	h := sha256.New()
	var length [8]byte
	for _, rel := range paths {
		pathBytes := []byte(rel)
		binary.BigEndian.PutUint64(length[:], uint64(len(pathBytes)))
		_, _ = h.Write(length[:])
		_, _ = h.Write(pathBytes)

		f, err := os.Open(filepath.Join(rootAbs, filepath.FromSlash(rel)))
		if err != nil {
			return "", err
		}
		info, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return "", err
		}
		binary.BigEndian.PutUint64(length[:], uint64(info.Size()))
		_, _ = h.Write(length[:])
		if _, err := io.Copy(h, f); err != nil {
			_ = f.Close()
			return "", err
		}
		if err := f.Close(); err != nil {
			return "", err
		}
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}
