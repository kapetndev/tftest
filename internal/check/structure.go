package check

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type StructureCheck struct {
}

// NewStructureCheck creates a new [StructureCheck].
func NewStructureCheck() *StructureCheck {
	return &StructureCheck{}
}

// Name returns the name of the check.
func (c *StructureCheck) Name() string {
	return "structure"
}

// Run executes the structure check.
func (c *StructureCheck) Run(ctx context.Context, modulePath string, opts RunOptions) Result {
	result := Result{
		Module: modulePath,
		Check:  c.Name(),
	}

	info, err := os.Stat(modulePath)
	if err != nil {
		result.Status = StatusFailed
		result.Details = err.Error()
		if os.IsNotExist(err) {
			result.Message = fmt.Sprintf("directory '%s' does not exist", modulePath)
			return result
		}
		result.Message = "Failed to access directory"
		return result
	}

	if !info.IsDir() {
		result.Status = StatusFailed
		result.Message = fmt.Sprintf("'%s' is not a directory", modulePath)
		return result
	}

	// Check for .tf files
	ok, err := c.hasTerraformFiles(modulePath)
	if err != nil {
		result.Status = StatusFailed
		result.Message = "Failed to scan directory"
		result.Details = err.Error()
		return result
	}
	if !ok {
		result.Status = StatusFailed
		result.Message = fmt.Sprintf("no .tf files found in '%s'", modulePath)
		return result
	}

	result.Status = StatusPassed
	result.Message = "Module structure is valid"
	return result
}

func (c *StructureCheck) hasTerraformFiles(modulePath string) (bool, error) {
	hasTerraformFiles := false
	err := filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only check top level
		if path != modulePath && info.IsDir() {
			return filepath.SkipDir
		}

		if filepath.Ext(path) == ".tf" {
			hasTerraformFiles = true
			return filepath.SkipAll
		}

		return nil
	})
	if err != nil {
		return false, err
	}
	return hasTerraformFiles, nil
}

// HaltOnFailure indicates whether execution should stop if this check fails.
func (c *StructureCheck) HaltOnFailure() bool {
	return true
}
