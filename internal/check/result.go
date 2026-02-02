package check

// Status represents the result status of a check.
type Status string

const (
	StatusPassed  Status = "passed"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

// Result represents the result of a single check.
type Result struct {
	Module  string
	Check   string
	Status  Status
	Message string
	Details string
}

// IsPassed returns true if the check passed or skipped.
func (r Result) IsPassed() bool {
	return r.Status == StatusPassed || r.Status == StatusSkipped
}
