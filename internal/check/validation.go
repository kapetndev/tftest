package check

import (
	"context"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

// ValidationCheck checks if Terraform configuration is valid.
type ValidationCheck struct {
	tfPath string
}

// NewValidationCheck creates a new [ValidationCheck].
func NewValidationCheck(path string) *ValidationCheck {
	return &ValidationCheck{
		tfPath: path,
	}
}

// Name returns the name of the check.
func (c *ValidationCheck) Name() string {
	return "validation"
}

// Run executes the validation check.
func (c *ValidationCheck) Run(ctx context.Context, modulePath string, opts RunOptions) Result {
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

	err = tf.Init(ctx, tfexec.Backend(false), tfexec.Upgrade(false))
	if err != nil {
		result.Status = StatusFailed
		result.Message = "Terraform init failed"
		result.Details = err.Error()
		return result
	}

	validateResult, err := tf.Validate(ctx)
	if err != nil {
		result.Status = StatusFailed
		result.Message = "Terraform validation failed"
		result.Details = err.Error()
		return result
	}

	if !validateResult.Valid {
		result.Status = StatusFailed
		result.Message = "Terraform configuration invalid"
		if len(validateResult.Diagnostics) > 0 {
			result.Details = formatDiagnostics(validateResult.Diagnostics)
		}
		return result
	}

	result.Status = StatusPassed
	result.Message = "Terraform configuration valid"
	return result
}

func formatDiagnostics(diagnostics []tfjson.Diagnostic) string {
	if len(diagnostics) == 0 {
		return ""
	}

	var output string
	for _, diag := range diagnostics {
		output += diag.Summary + "\n"
		if diag.Detail != "" {
			output += "  " + diag.Detail + "\n"
		}
	}
	return output
}

// HaltOnFailure indicates whether execution should stop if this check fails.
func (c *ValidationCheck) HaltOnFailure() bool {
	return false
}
