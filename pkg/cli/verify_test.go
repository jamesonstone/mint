package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVerifyCommand(t *testing.T) {
	expected := "sha256:" + strings.Repeat("ab", 32)
	different := "sha256:" + strings.Repeat("cd", 32)
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{name: "match", args: []string{"--expected", expected, "--observed", expected}},
		{name: "hex case", args: []string{"--expected", expected, "--observed", "sha256:" + strings.Repeat("AB", 32)}},
		{name: "missing expected", args: []string{"--observed", expected}, want: "--expected digest is required"},
		{name: "missing observed", args: []string{"--expected", expected}, want: "--observed digest is required"},
		{name: "empty observed", args: []string{"--expected", expected, "--observed="}, want: "--observed digest is required"},
		{name: "missing prefix", args: []string{"--expected", strings.TrimPrefix(expected, "sha256:"), "--observed", expected}, want: "--expected digest must be sha256:"},
		{name: "wrong algorithm", args: []string{"--expected", "sha512:" + strings.Repeat("ab", 32), "--observed", expected}, want: "--expected digest must be sha256:"},
		{name: "short observed", args: []string{"--expected", expected, "--observed", "sha256:ab"}, want: "--observed digest must be sha256:"},
		{name: "non-hex observed", args: []string{"--expected", expected, "--observed", "sha256:" + strings.Repeat("zz", 32)}, want: "--observed digest must be sha256:"},
		{name: "whitespace", args: []string{"--expected", expected, "--observed", expected + " "}, want: "--observed digest must be sha256:"},
		{name: "mismatch", args: []string{"--expected", expected, "--observed", different}, want: "digest mismatch"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			cmd := newVerifyCommand()
			cmd.SetArgs(test.args)
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			err := cmd.Execute()
			if test.want == "" {
				if err != nil || output.Len() != 0 {
					t.Fatalf("verify result = %v, output = %q", err, output.String())
				}
			} else if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("verify error = %v, want %q", err, test.want)
			}
		})
	}
}
