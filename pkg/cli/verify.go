package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newVerifyCommand())
}

func newVerifyCommand() *cobra.Command {
	var expected, observed string
	cmd := &cobra.Command{
		Use:           "verify",
		Short:         "Compare an expected and an observed SHA-256 digest",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := validateVerifyDigest("expected", expected); err != nil {
				return err
			}
			if err := validateVerifyDigest("observed", observed); err != nil {
				return err
			}
			if !strings.EqualFold(expected, observed) {
				return fmt.Errorf("digest mismatch: expected %s, observed %s", expected, observed)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&expected, "expected", "", "SHA-256 digest from the build record (required)")
	cmd.Flags().StringVar(&observed, "observed", "", "SHA-256 digest from the runtime (required)")
	return cmd
}

func validateVerifyDigest(flag, digest string) error {
	if digest == "" {
		return fmt.Errorf("--%s digest is required", flag)
	}
	value, hasPrefix := strings.CutPrefix(digest, "sha256:")
	if _, err := hex.DecodeString(value); !hasPrefix || len(value) != sha256.Size*2 || err != nil {
		return fmt.Errorf("--%s digest must be sha256: followed by 64 hex characters", flag)
	}
	return nil
}
