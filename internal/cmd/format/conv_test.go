package format_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/kapetndev/tftest/internal/check"
	"github.com/kapetndev/tftest/internal/cmd/format"
)

func TestToModuleResult(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		results []check.Result
		want    format.ModuleResult
	}{
		"empty results": {
			results: []check.Result{},
			want:    format.ModuleResult{},
		},
		"single result": {
			results: []check.Result{
				{Module: "module1", Check: "check1", Status: check.StatusPassed, Message: "passed"},
			},
			want: format.ModuleResult{
				Module: "module1",
				Results: []format.CheckResult{
					{Check: "check1", Status: "passed", Message: "passed"},
				},
				Passed: true,
			},
		},
		"multiple results": {
			results: []check.Result{
				{Module: "module1", Check: "check1", Status: check.StatusPassed, Message: "passed"},
				{Module: "module1", Check: "check2", Status: check.StatusFailed, Message: "failed"},
			},
			want: format.ModuleResult{
				Module: "module1",
				Results: []format.CheckResult{
					{Check: "check1", Status: "passed", Message: "passed"},
					{Check: "check2", Status: "failed", Message: "failed"},
				},
				Passed: false,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := format.ToModuleResult(tt.results)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestToModulesResult(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		results []check.Result
		want    format.ModulesResult
	}{
		"empty results": {
			results: []check.Result{},
			want:    []format.ModuleResult{},
		},
		"single result": {
			results: []check.Result{
				{Module: "module1", Check: "check1", Status: check.StatusPassed, Message: "passed"},
			},
			want: []format.ModuleResult{
				{
					Module: "module1",
					Results: []format.CheckResult{
						{Check: "check1", Status: "passed", Message: "passed"},
					},
					Passed: true,
				},
			},
		},
		"multiple results": {
			results: []check.Result{
				{Module: "module1", Check: "check1", Status: check.StatusPassed, Message: "passed"},
				{Module: "module1", Check: "check2", Status: check.StatusFailed, Message: "failed"},
				{Module: "module2", Check: "check1", Status: check.StatusPassed, Message: "passed"},
			},
			want: []format.ModuleResult{
				{
					Module: "module1",
					Results: []format.CheckResult{
						{Check: "check1", Status: "passed", Message: "passed"},
						{Check: "check2", Status: "failed", Message: "failed"},
					},
					Passed: false,
				},
				{
					Module: "module2",
					Results: []format.CheckResult{
						{Check: "check1", Status: "passed", Message: "passed"},
					},
					Passed: true,
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := format.ToModulesResult(tt.results)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
