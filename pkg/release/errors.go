package release

import "fmt"

type validationIssue struct {
	Field   string
	Message string
}

func (e *validationIssue) Error() string {
	if e == nil {
		return ""
	}
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func validationError(field string, format string, args ...any) error {
	return &validationIssue{
		Field:   field,
		Message: fmt.Sprintf(format, args...),
	}
}
