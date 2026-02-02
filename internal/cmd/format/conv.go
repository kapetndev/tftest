package format

import (
	"github.com/kapetndev/tftest/internal/check"
)

func ToModuleResult(results []check.Result) ModuleResult {
	if len(results) == 0 {
		return ModuleResult{}
	}

	var checkResults []CheckResult
	for _, res := range results {
		checkResults = append(checkResults, CheckResult{
			Check:   res.Check,
			Status:  string(res.Status),
			Message: res.Message,
			Details: res.Details,
		})
	}

	return ModuleResult{
		Module:  results[0].Module,
		Results: checkResults,
		Passed:  check.AllPassed(results),
	}
}

func ToModulesResult(results []check.Result) ModulesResult {
	moduleResults := make(map[string][]check.Result)
	for _, res := range results {
		moduleResults[res.Module] = append(moduleResults[res.Module], res)
	}

	ret := []ModuleResult{}
	for _, res := range moduleResults {
		ret = append(ret, ToModuleResult(res))
	}

	return ret
}
