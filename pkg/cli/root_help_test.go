package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestRenderRootHelpIncludesMintStructure(t *testing.T) {
	originalTerminalWriterCheck := terminalWriterCheck
	terminalWriterCheck = func(w io.Writer) bool {
		return false
	}
	defer func() {
		terminalWriterCheck = originalTerminalWriterCheck
	}()

	output := &bytes.Buffer{}
	rootCmd.SetOut(output)
	defer rootCmd.SetOut(nil)

	if err := renderRootHelp(rootCmd); err != nil {
		t.Fatalf("renderRootHelp() error = %v", err)
	}

	help := output.String()
	for _, want := range []string{
		"Mint is a release tooling CLI",
		"Release Intent",
		"Usage",
		"Available Commands",
		"version",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("root help missing %q:\n%s", want, help)
		}
	}
}
