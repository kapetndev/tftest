package check

import "context"

type RunOptions struct {
	Verbose bool
}

// Check represents a test that can be run against a module.
type Check interface {
	// Name returns the name of the check.
	Name() string

	// Run executes the check and returns the result.
	Run(ctx context.Context, modulePath string, opts RunOptions) Result

	// HaltOnFailure indicates whether execution should stop if this check fails.
	HaltOnFailure() bool
}

// Executor manages and runs a set of checks.
type Executor struct {
	checks []Check
}

// NewExecutor creates a new [Checker] with the provided checks.
func NewExecutor(checks []Check) *Executor {
	return &Executor{
		checks: checks,
	}
}

// RunAll executes all checks against the specified module path.
func (e *Executor) RunAll(ctx context.Context, modulePath string, opts RunOptions) []Result {
	results := []Result{}
	for _, check := range e.checks {
		result := check.Run(ctx, modulePath, opts)
		results = append(results, result)

		if result.Status == StatusFailed && check.HaltOnFailure() {
			break
		}
	}
	return results
}

// AllPassed returns true if all results indicate passed checks.
func AllPassed(results []Result) bool {
	for _, r := range results {
		if !r.IsPassed() {
			return false
		}
	}
	return true
}
