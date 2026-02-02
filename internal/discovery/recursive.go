package discovery

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
)

// RecursiveModuleFinder walks the directory tree to find all Terraform modules.
type RecursiveModuleFinder struct {
	fsys fs.FS
	// MaxDepth limits how deep to search. 0 means unlimited.
	MaxDepth int
}

// NewRecursiveModuleFinder creates a new [RecursiveModuleFinder].
func NewRecursiveModuleFinder(fsys fs.FS) *RecursiveModuleFinder {
	return &RecursiveModuleFinder{
		fsys:     fsys,
		MaxDepth: 0,
	}
}

// FindModules walks the directory tree to find all Terraform modules,
// respecting the [RecursiveModuleFinder.MaxDepth] setting, and ingoring
// certain, standard, directories.
func (f *RecursiveModuleFinder) FindModules(ctx context.Context, rootPath string) ([]string, error) {
	visited := make(map[string]bool)
	modules, err := f.walkModules(ctx, rootPath, 0, visited)
	if err != nil {
		return nil, err
	}

	if len(modules) == 0 {
		return nil, fmt.Errorf("no Terraform modules found in '%s'", rootPath)
	}

	return modules, nil
}

func (f *RecursiveModuleFinder) walkModules(ctx context.Context, currentPath string, depth int, visited map[string]bool) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check depth limit.
	if f.MaxDepth > 0 && depth > f.MaxDepth {
		return nil, nil
	}

	// Resolve current path to absolute for deduplication
	absPath, err := filepath.Abs(currentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path '%s': %w", currentPath, err)
	}

	if visited[absPath] {
		return nil, nil
	}
	visited[absPath] = true

	var modules []string

	// Check if this directory is a module.
	if isModule(f.fsys, currentPath) {
		modules = append(modules, currentPath)
		// Don't descend into child directories of a module.
		return modules, nil
	}

	// Scan subdirectories.
	entries, err := fs.ReadDir(f.fsys, currentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory '%s': %w", currentPath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories and common directories to ignore
		name := entry.Name()
		if shouldSkipDirectory(name) {
			continue
		}

		childPath := filepath.Join(currentPath, name)
		childModules, err := f.walkModules(ctx, childPath, depth+1, visited)
		if err != nil {
			return nil, err
		}
		modules = append(modules, childModules...)
	}

	return modules, nil
}

// isModule checks if a directory contains Terraform files.
func isModule(fsys fs.FS, path string) bool {
	entries, err := fs.ReadDir(fsys, path)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".tf" {
			return true
		}
	}

	return false
}

// shouldSkipDirectory determines if a directory should be skipped during
// traversal.
func shouldSkipDirectory(name string) bool {
	skipDirs := map[string]bool{
		".git":              true,
		".terraform":        true,
		"node_modules":      true,
		".terragrunt-cache": true,
		"vendor":            true,
	}

	// Skip hidden directories
	if len(name) > 0 && name[0] == '.' {
		return true
	}

	return skipDirs[name]
}
