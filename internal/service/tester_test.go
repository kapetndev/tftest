package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kapetndev/tftest/internal/check"
	"github.com/kapetndev/tftest/internal/service"
)

// mockFinder implements the Finder interface for testing.
type mockFinder struct {
	modules []string
	err     error
}

func (m *mockFinder) FindModules(ctx context.Context, rootPath string) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return m.modules, m.err
}

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

func TestTestService_RunChecks(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		finder          *mockFinder
		checks          []check.Check
		request         service.TestRequest
		want            int
		wantErr         bool
		wantMsgContains string
	}{
		"single module all checks pass": {
			finder: &mockFinder{
				modules: []string{"app"},
			},
			checks: []check.Check{
				&mockCheck{checkName: "structure", status: check.StatusPassed},
				&mockCheck{checkName: "formatting", status: check.StatusPassed},
			},
			request: service.TestRequest{
				RootPath: ".",
				Verbose:  false,
			},
			want: 2,
		},
		"single module with failures": {
			finder: &mockFinder{
				modules: []string{"app"},
			},
			checks: []check.Check{
				&mockCheck{checkName: "structure", status: check.StatusPassed},
				&mockCheck{checkName: "formatting", status: check.StatusFailed, message: "not formatted"},
			},
			request: service.TestRequest{
				RootPath: ".",
				Verbose:  false,
			},
			want: 2,
		},
		"multiple modules all pass": {
			finder: &mockFinder{
				modules: []string{"app", "db", "network"},
			},
			checks: []check.Check{
				&mockCheck{checkName: "structure", status: check.StatusPassed},
			},
			request: service.TestRequest{
				RootPath: ".",
				Verbose:  false,
			},
			want: 3,
		},
		"multiple modules with mixed results": {
			finder: &mockFinder{
				modules: []string{"app", "db"},
			},
			checks: []check.Check{
				&mockCheck{checkName: "structure", status: check.StatusPassed},
				&mockCheck{checkName: "formatting", status: check.StatusFailed, message: "formatting error"},
			},
			request: service.TestRequest{
				RootPath: ".",
				Verbose:  false,
			},
			want: 4,
		},
		"all modules fail": {
			finder: &mockFinder{
				modules: []string{"app", "db"},
			},
			checks: []check.Check{
				&mockCheck{checkName: "structure", status: check.StatusFailed, message: "missing files"},
			},
			request: service.TestRequest{
				RootPath: ".",
				Verbose:  false,
			},
			want: 2,
		},
		"finder returns error": {
			finder: &mockFinder{
				err: errors.New("no modules found"),
			},
			checks: []check.Check{
				&mockCheck{checkName: "structure", status: check.StatusPassed},
			},
			request: service.TestRequest{
				RootPath: ".",
				Verbose:  false,
			},
			wantErr:         true,
			wantMsgContains: "failed to find modules",
		},
		"no modules found returns empty results": {
			finder: &mockFinder{
				modules: []string{},
			},
			checks: []check.Check{
				&mockCheck{checkName: "structure", status: check.StatusPassed},
			},
			request: service.TestRequest{
				RootPath: ".",
				Verbose:  false,
			},
			want: 0,
		},
		"no checks configured": {
			finder: &mockFinder{
				modules: []string{"app"},
			},
			checks: []check.Check{},
			request: service.TestRequest{
				RootPath: ".",
				Verbose:  false,
			},
			want: 0,
		},
		"aggregates results across multiple modules": {
			finder: &mockFinder{
				modules: []string{"module1", "module2", "module3"},
			},
			checks: []check.Check{
				&mockCheck{checkName: "check1", status: check.StatusPassed},
				&mockCheck{checkName: "check2", status: check.StatusFailed, message: "error"},
			},
			request: service.TestRequest{
				RootPath: ".",
				Verbose:  false,
			},
			want: 6,
		},
		"multiple checks with various statuses": {
			finder: &mockFinder{
				modules: []string{"app"},
			},
			checks: []check.Check{
				&mockCheck{checkName: "structure", status: check.StatusPassed},
				&mockCheck{checkName: "formatting", status: check.StatusSkipped, message: "skipped"},
				&mockCheck{checkName: "validation", status: check.StatusPassed},
			},
			request: service.TestRequest{
				RootPath: ".",
				Verbose:  false,
			},
			want: 3,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			executor := check.NewExecutor(tt.checks)
			svc := service.NewTestService(executor, tt.finder)

			got, err := svc.CollectResults(context.Background(), tt.request)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if len(got) != tt.want {
				t.Errorf("mismatch:\n  got:  %d\n  want: %d", len(got), tt.want)
			}

			if tt.wantErr && tt.wantMsgContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantMsgContains) {
					t.Errorf("error = %v, should contain %q", err, tt.wantMsgContains)
				}
			}

			// Verify all results have the correct module path set.
			if !tt.wantErr && len(tt.finder.modules) > 0 {
				moduleSet := make(map[string]bool)
				for _, mod := range tt.finder.modules {
					moduleSet[mod] = true
				}
				for _, res := range got {
					if !moduleSet[res.Module] {
						t.Errorf("Result has unexpected module path: %q", res.Module)
					}
				}
			}
		})
	}
}

func TestTestService_RunChecks_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	executor := check.NewExecutor([]check.Check{
		&mockCheck{checkName: "structure", status: check.StatusPassed},
	})

	finder := &mockFinder{
		err: context.Canceled,
	}

	req := service.TestRequest{
		RootPath: ".",
		Verbose:  false,
	}

	svc := service.NewTestService(executor, finder)

	_, err := svc.CollectResults(ctx, req)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}
