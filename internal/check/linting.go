package check

import (
	"context"
	"os/exec"
	"strings"
)

// LintingCheck checks Terraform code using tflint.
type LintingCheck struct{}

// NewLintingCheck creates a new [LintingCheck].
func NewLintingCheck() *LintingCheck {
	return &LintingCheck{}
}

// Name returns the name of the check.
func (c *LintingCheck) Name() string {
	return "linting"
}

// Run executes the linting check.
func (c *LintingCheck) Run(ctx context.Context, modulePath string, opts RunOptions) Result {
	result := Result{
		Module: modulePath,
		Check:  c.Name(),
	}

	// Check if tflint is available
	if !isTFLintAvailable() {
		result.Status = StatusSkipped
		result.Message = "TFLint not installed"
		return result
	}

	// Run tflint --init
	initCmd := exec.CommandContext(ctx, "tflint", "--init")
	initCmd.Dir = modulePath

	// Suppress output unless verbose.
	if !opts.Verbose {
		initCmd.Stdout = nil
		initCmd.Stderr = nil
	}

	if err := initCmd.Run(); err != nil {
		result.Status = StatusFailed
		result.Message = "TFLint init failed"
		result.Details = err.Error()
		return result
	}

	// Run tflint
	lintCmd := exec.CommandContext(ctx, "tflint")
	lintCmd.Dir = modulePath

	output, err := lintCmd.CombinedOutput()
	if err != nil {
		result.Status = StatusFailed
		result.Message = "TFLint found issues"
		if opts.Verbose {
			result.Details = string(output)
		} else {
			result.Details = "Run with -v for details"
		}
		return result
	}

	result.Status = StatusPassed
	result.Message = "No linting issues found"
	if opts.Verbose && len(output) > 0 {
		result.Details = strings.TrimSpace(string(output))
	}
	return result
}

func isTFLintAvailable() bool {
	_, err := exec.LookPath("tflint")
	return err == nil
}

// HaltOnFailure indicates whether execution should stop if this check fails.
func (c *LintingCheck) HaltOnFailure() bool {
	return false
}
