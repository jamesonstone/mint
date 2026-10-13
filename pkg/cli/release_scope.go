package cli

import (
	"github.com/jamesonstone/mint/pkg/release"
	"github.com/spf13/pflag"
)

// scopeFlags holds the component scope shared by release commands.
type scopeFlags struct {
	tagPrefix string
	paths     []string
}

// bindScopeFlags adds --tag-prefix and, when withPaths is set, the repeatable
// --path flag.
func bindScopeFlags(flags *pflag.FlagSet, values *scopeFlags, withPaths bool) {
	flags.StringVar(&values.tagPrefix, "tag-prefix", "", "component tag prefix, for example backend/ for backend/vX.Y.Z tags; empty means unprefixed vX.Y.Z")
	if withPaths {
		flags.StringArrayVar(&values.paths, "path", nil, "repeatable Git pathspec; only commits touching these paths count toward the component's version and notes")
	}
}

func (values scopeFlags) scope() release.Scope {
	return release.Scope{TagPrefix: values.tagPrefix, Paths: values.paths}
}
