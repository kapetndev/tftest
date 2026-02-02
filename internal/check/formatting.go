package check

import (
	"context"

	"github.com/hashicorp/terraform-exec/tfexec"
)

// FormattingCheck checks if Terraform files are properly formatted.
type FormattingCheck struct {
	tfPath string
}

// NewFormattingCheck creates a new [FormattingCheck].
func NewFormattingCheck(path string) *FormattingCheck {
	return &FormattingCheck{
		tfPath: path,
	}
}

// Name returns the name of the check.
func (c *FormattingCheck) Name() string {
	return "formatting"
}

// Run executes the formatting check.
func (c *FormattingCheck) Run(ctx context.Context, modulePath string, opts RunOptions) Result {
	result := Result{
		Module: modulePath,
		Check:  c.Name(),
	}

	tf, err := tfexec.NewTerraform(modulePath, c.tfPath)
	if err != nil {
		result.Status = StatusFailed
		result.Message = "Failed to initialize terraform"
		result.Details = err.Error()
		return result
	}

	// Suppress terraform output unless verbose.
	if !opts.Verbose {
		tf.SetStdout(nil)
		tf.SetStderr(nil)
	}

	ok, _, err := tf.FormatCheck(ctx)
	if err != nil {
		result.Status = StatusFailed
		result.Message = "Formatting check failed (run 'terraform fmt -recursive' to fix)"
		result.Details = err.Error()
		return result
	}

	if !ok {
		result.Status = StatusFailed
		result.Message = "Files need formatting"
		result.Details = "Run 'terraform fmt -recursive' to fix"
		return result
	}

	result.Status = StatusPassed
	result.Message = "All files properly formatted"
	return result
}

// HaltOnFailure indicates whether execution should stop if this check fails.
func (c *FormattingCheck) HaltOnFailure() bool {
	return false
}
