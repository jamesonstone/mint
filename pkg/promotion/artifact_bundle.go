package promotion

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
)

// BundleDigest binds the complete named artifact set. encoding/json sorts map
// keys; changing any component or its build configuration changes the identity.
func BundleDigest(members map[string]ArtifactObject) (string, error) {
	if len(members) == 0 {
		return "", fmt.Errorf("artifact bundle is empty")
	}
	data, err := json.Marshal(members)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data)), nil
}

// ValidateArtifact accepts an exact single object or a complete bundle digest.
func ValidateArtifact(a Artifact) error {
	if a.Reference == "" || !digestPattern.MatchString(a.Digest) || !digestPattern.MatchString(a.Configuration) {
		return fmt.Errorf("artifact lacks immutable reference, digest or build configuration")
	}
	if a.Members == nil {
		return nil
	}
	if len(a.Members) == 0 {
		return fmt.Errorf("explicit artifact bundle must not be empty")
	}
	for name, member := range a.Members {
		if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`).MatchString(name) || member.Reference == "" || !digestPattern.MatchString(member.Digest) || !digestPattern.MatchString(member.Configuration) {
			return fmt.Errorf("invalid artifact bundle member %s", name)
		}
	}
	digest, err := BundleDigest(a.Members)
	if err != nil || a.Digest != digest {
		return fmt.Errorf("artifact bundle digest does not bind all members")
	}
	return nil
}

// AcceptsEnvironment preserves legacy scope and explicitly permits authenticated
// reusable build candidates only in schema 2 journals.
func (s State) AcceptsEnvironment(environment string) bool {
	return environment == s.Environment || (s.Schema == 2 && environment == "build")
}

// RuntimeConfiguration keeps legacy build/runtime identity compatibility.
func (b Baseline) RuntimeConfiguration() string {
	if b.Configuration != "" {
		return b.Configuration
	}
	return b.Candidate.Artifact.Configuration
}
