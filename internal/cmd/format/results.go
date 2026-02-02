package format

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
	colorCyan   = "\033[0;36m"
)

// CheckResult represents the result of a single check. Used for output
// formatting.
type CheckResult struct {
	Check   string `json:"check"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"` // Verbose output
}

// ModuleResult represents the results of all checks. Used for output
// formatting.
type ModuleResult struct {
	Module  string        `json:"module"`
	Results []CheckResult `json:"results"`
	Passed  bool          `json:"passed"`
}

func (r ModuleResult) FormatText() ([]byte, error) {
	if len(r.Results) == 0 {
		return []byte("No results to display.\n"), nil
	}

	var buffer bytes.Buffer
	moduleName := r.Module

	// Print header.
	r.printHeader(&buffer, moduleName)

	for _, result := range r.Results {
		// Skip structure check in output (it's implicit)
		if result.Check == "structure" {
			if result.Status == "failed" {
				r.printCheckResult(&buffer, result)
			}
			continue
		}

		r.printCheckResult(&buffer, result)
	}

	// Print footer.
	status := "passed"
	if !r.Passed {
		status = "failed"
	}
	r.printFooter(&buffer, moduleName, status)

	return buffer.Bytes(), nil
}

func (r ModuleResult) printHeader(b *bytes.Buffer, moduleName string) {
	border := strings.Repeat("═", 78)

	fmt.Fprintf(b, "%s╔%s╗%s\n", colorCyan, border, colorReset)
	fmt.Fprintf(b, "%s║%s %sTesting Module:%s %-60s %s║%s\n", colorCyan, colorReset, colorBlue, colorReset, moduleName, colorCyan, colorReset)
	fmt.Fprintf(b, "%s╚%s╗%s\n", colorCyan, border, colorReset)
}

func (r ModuleResult) printCheckResult(b *bytes.Buffer, result CheckResult) {
	var statusColor, statusText string

	switch result.Status {
	case "passed":
		statusColor = colorGreen
		statusText = "[PASS]"
	case "failed":
		statusColor = colorRed
		statusText = "[FAIL]"
	case "skipped":
		statusColor = colorYellow
		statusText = "[SKIP]"
	}

	fmt.Fprintf(b, " %s%s%s %s\n", statusColor, statusText, colorReset, result.Message)

	if result.Details != "" {
		details := strings.TrimSpace(result.Details)
		for line := range strings.SplitSeq(details, "\n") {
			fmt.Fprintf(b, "       %s\n", line)
		}
	}
}

func (r ModuleResult) printFooter(b *bytes.Buffer, moduleName, status string) {
	border := strings.Repeat("═", 78)

	fmt.Fprintf(b, "%s╚%s╝%s\n", colorCyan, border, colorReset)

	if status == "passed" {
		fmt.Fprintf(b, "%s✓%s %s - %sALL CHECKS PASSED%s\n\n", colorGreen, colorReset, moduleName, colorGreen, colorReset)
	} else {
		fmt.Fprintf(b, "%s✗%s %s - %sCHECKS FAILED%s\n\n", colorRed, colorReset, moduleName, colorRed, colorReset)
	}
}

type ModulesResult []ModuleResult

func (r ModulesResult) FormatText() ([]byte, error) {
	var buffer bytes.Buffer

	for _, moduleResults := range r {
		moduleText, err := moduleResults.FormatText()
		if err != nil {
			return nil, err
		}
		buffer.Write(moduleText)
	}

	return buffer.Bytes(), nil
}
