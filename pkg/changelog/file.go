package changelog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var releaseHeaderPattern = regexp.MustCompile(`^## \[(\d+\.\d+\.\d+)\]\([^)]+\) - \d{4}-\d{2}-\d{2}$`)

func readExistingChangelog(path string, version string) (string, os.FileMode, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", 0o644, nil
		}
		return "", 0, err
	}
	if info.IsDir() {
		return "", 0, fmt.Errorf("%s is a directory", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", 0, err
	}
	content := string(data)
	if err := validateExistingChangelog(content, version); err != nil {
		return "", 0, err
	}
	return content, info.Mode().Perm(), nil
}

func validateExistingChangelog(content string, version string) error {
	for _, line := range strings.Split(content, "\n") {
		if !strings.HasPrefix(line, "## ") {
			continue
		}
		match := releaseHeaderPattern.FindStringSubmatch(line)
		if match == nil {
			return fmt.Errorf("cannot parse CHANGELOG.md")
		}
		if match[1] == version {
			return fmt.Errorf("version %s already in CHANGELOG", version)
		}
	}
	return nil
}

func prependRelease(path string, block string, existing string, mode os.FileMode) error {
	nextContent := mergeReleaseBlock(block, existing)

	dir := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dir, ".mint-changelog-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.WriteString(nextContent); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Chmod(mode); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	removeTemp = false
	return nil
}

func mergeReleaseBlock(block string, existing string) string {
	releaseBlock := strings.TrimRight(block, "\n") + "\n"
	if existing == "" {
		return "# Changelog\n\n" + releaseBlock
	}

	if strings.HasPrefix(existing, "# ") {
		headingEnd := strings.Index(existing, "\n")
		if headingEnd >= 0 {
			heading := existing[:headingEnd+1]
			rest := strings.TrimLeft(existing[headingEnd+1:], "\n")
			if rest == "" {
				return heading + "\n" + releaseBlock
			}
			return heading + "\n" + releaseBlock + "\n" + rest
		}
		return existing + "\n\n" + releaseBlock
	}

	return releaseBlock + "\n" + existing
}
