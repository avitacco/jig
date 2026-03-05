package module

import "regexp"

type Severity int

const (
	Info Severity = iota
	Warning
	Error
)

func (s Severity) String() string {
	switch s {
	case Info:
		return "info"
	case Warning:
		return "warning"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

type ValidationResult struct {
	Level   Severity
	Field   string
	Message string
}

func (m Metadata) Validate() []ValidationResult {
	var results []ValidationResult

	//
	// Name validation
	//
	if m.Name == "" {
		results = append(results, ValidationResult{
			Level:   Error,
			Field:   "name",
			Message: "name is required",
		})
	}

	validNameRe := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	if !validNameRe.MatchString(m.Name) {
		results = append(results, ValidationResult{
			Level:   Warning,
			Field:   "name",
			Message: "name must start with a lowercase letter and contain only lowercase letters, numbers, and underscores",
		})
	}

	if m.Version == "" {
		results = append(results, ValidationResult{
			Level:   Error,
			Field:   "version",
			Message: "version is required",
		})
	}

	if m.Author == "" {
		results = append(results, ValidationResult{
			Level:   Error,
			Field:   "author",
			Message: "author is required",
		})
	}

	if m.License == "" {
		results = append(results, ValidationResult{
			Level:   Error,
			Field:   "license",
			Message: "license is required",
		})
	}

	if m.Summary == "" {
		results = append(results, ValidationResult{
			Level:   Error,
			Field:   "summary",
			Message: "summary is required",
		})
	}

	if m.Source == "" {
		results = append(results, ValidationResult{
			Level:   Error,
			Field:   "source",
			Message: "source is required",
		})
	}

	return results
}
