package check_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/kapetndev/tftest/internal/check"
)

// mockCheck implements the Check interface for testing.
type mockCheck struct {
	checkName  string
	status     check.Status
	message    string
	details    string
	haltOnFail bool
}

func (m *mockCheck) Name() string {
	return m.checkName
}

func (m *mockCheck) Run(ctx context.Context, modulePath string, opts check.RunOptions) check.Result {
	if ctx.Err() != nil {
		return check.Result{
			Module:  modulePath,
			Check:   m.checkName,
			Status:  check.StatusFailed,
			Message: "canceled",
		}
	}

	return check.Result{
		Module:  modulePath,
		Check:   m.checkName,
		Status:  m.status,
		Message: m.message,
		Details: m.details,
	}
}

func (m *mockCheck) HaltOnFailure() bool {
	return m.haltOnFail
}

func TestExecutor_RunAll(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		checks []check.Check
		want   []check.Result
	}{
		"no checks": {
			checks: []check.Check{},
			want:   []check.Result{},
		},
		"single passing check": {
			checks: []check.Check{
				&mockCheck{
					checkName: "test-check",
					status:    check.StatusPassed,
					message:   "check passed",
				},
			},
			want: []check.Result{
				{
					Module:  "module/path",
					Check:   "test-check",
					Status:  check.StatusPassed,
					Message: "check passed",
				},
			},
		},
		"multiple passing checks": {
			checks: []check.Check{
				&mockCheck{
					checkName: "check-1",
					status:    check.StatusPassed,
					message:   "first passed",
				},
				&mockCheck{
					checkName: "check-2",
					status:    check.StatusPassed,
					message:   "second passed",
				},
			},
			want: []check.Result{
				{
					Module:  "module/path",
					Check:   "check-1",
					Status:  check.StatusPassed,
					Message: "first passed",
				},
				{
					Module:  "module/path",
					Check:   "check-2",
					Status:  check.StatusPassed,
					Message: "second passed",
				},
			},
		},
		"failing check without halt": {
			checks: []check.Check{
				&mockCheck{
					checkName:  "check-1",
					status:     check.StatusFailed,
					message:    "failed",
					haltOnFail: false,
				},
				&mockCheck{
					checkName: "check-2",
					status:    check.StatusPassed,
					message:   "passed",
				},
			},
			want: []check.Result{
				{
					Module:  "module/path",
					Check:   "check-1",
					Status:  check.StatusFailed,
					Message: "failed",
				},
				{
					Module:  "module/path",
					Check:   "check-2",
					Status:  check.StatusPassed,
					Message: "passed",
				},
			},
		},
		"failing check with halt stops execution": {
			checks: []check.Check{
				&mockCheck{
					checkName:  "check-1",
					status:     check.StatusFailed,
					message:    "failed",
					details:    "detailed failure info",
					haltOnFail: true,
				},
				&mockCheck{
					checkName: "check-2",
					status:    check.StatusPassed,
					message:   "should not run",
				},
			},
			want: []check.Result{
				{
					Module:  "module/path",
					Check:   "check-1",
					Status:  check.StatusFailed,
					Message: "failed",
					Details: "detailed failure info",
				},
			},
		},
		"skipped check": {
			checks: []check.Check{
				&mockCheck{
					checkName: "skipped-check",
					status:    check.StatusSkipped,
					message:   "condition not met",
				},
			},
			want: []check.Result{
				{
					Module:  "module/path",
					Check:   "skipped-check",
					Status:  check.StatusSkipped,
					Message: "condition not met",
				},
			},
		},
		"passed check followed by failed check with halt": {
			checks: []check.Check{
				&mockCheck{
					checkName: "check-1",
					status:    check.StatusPassed,
					message:   "passed",
				},
				&mockCheck{
					checkName:  "check-2",
					status:     check.StatusFailed,
					message:    "failed",
					haltOnFail: true,
				},
				&mockCheck{
					checkName: "check-3",
					status:    check.StatusPassed,
					message:   "should not run",
				},
			},
			want: []check.Result{
				{
					Module:  "module/path",
					Check:   "check-1",
					Status:  check.StatusPassed,
					Message: "passed",
				},
				{
					Module:  "module/path",
					Check:   "check-2",
					Status:  check.StatusFailed,
					Message: "failed",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			executor := check.NewExecutor(tt.checks)

			got := executor.RunAll(context.Background(), "module/path", check.RunOptions{})
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestExecutor_RunAll_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	// Since the context is canceled, we expect the check to fail with a
	// cancellation message.
	want := []check.Result{
		{
			Module:  "module/path",
			Check:   "structure",
			Status:  check.StatusFailed,
			Message: "canceled",
		},
	}

	executor := check.NewExecutor([]check.Check{
		&mockCheck{checkName: "structure", status: check.StatusPassed},
	})

	got := executor.RunAll(ctx, "module/path", check.RunOptions{})
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func TestAllPassed(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		results []check.Result
		want    bool
	}{
		"all passed": {
			results: []check.Result{
				{Status: check.StatusPassed},
				{Status: check.StatusPassed},
			},
			want: true,
		},
		"some failed": {
			results: []check.Result{
				{Status: check.StatusPassed},
				{Status: check.StatusFailed},
			},
			want: false,
		},
		"all skipped": {
			results: []check.Result{
				{Status: check.StatusSkipped},
				{Status: check.StatusSkipped},
			},
			want: true,
		},
		"mixed statuses": {
			results: []check.Result{
				{Status: check.StatusPassed},
				{Status: check.StatusSkipped},
				{Status: check.StatusFailed},
			},
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := check.AllPassed(tt.results)
			if got != tt.want {
				t.Errorf("mismatch:\n  got:  %t\n  want: %t", got, tt.want)
			}
		})
	}
}
