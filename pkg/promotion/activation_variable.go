package promotion

import (
	"errors"
	"regexp"
)

// DefaultActivationVariable gates generated controllers when policy names none.
const DefaultActivationVariable = "MINT_RELEASE_ENABLED"

// The MINT_ prefix keeps the switch out of the compatibility adapters' runtime
// configuration identity, which hashes every other repository variable: a
// switch outside the prefix would change that identity each time it flips.
var activationVariablePattern = regexp.MustCompile(`^MINT_[A-Z0-9_]{1,250}$`)

var errActivationVariable = errors.New("activation_variable must be an upper-case repository variable name that starts with MINT_")

// validActivationVariable matches a repository variable name that is safe to
// interpolate into the generated controller's Actions expression.
func validActivationVariable(name string) bool {
	return activationVariablePattern.MatchString(name)
}

// ControllerActivationVariable returns the repository variable that gates the
// generated environment controller.
func (p Policy) ControllerActivationVariable() string {
	if p.ActivationVariable != "" {
		return p.ActivationVariable
	}
	return DefaultActivationVariable
}
