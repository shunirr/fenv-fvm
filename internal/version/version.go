package version

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// VersionFileName is the name of the version file
	VersionFileName = ".flutter-version"
)

// FindProjectRoot searches upward from startDir for .flutter-version file
// Returns the directory containing the file or an error if not found
func FindProjectRoot(startDir string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	currentDir := absPath
	for {
		versionFile := filepath.Join(currentDir, VersionFileName)
		if _, err := os.Stat(versionFile); err == nil {
			return currentDir, nil
		}

		// Move to parent directory
		parentDir := filepath.Dir(currentDir)

		// Check if we've reached the root
		if parentDir == currentDir {
			return "", fmt.Errorf("fenv-fvm: no Flutter version configured (.flutter-version not found)")
		}

		currentDir = parentDir
	}
}

// ReadVersion reads the Flutter version from .flutter-version file in projectRoot
// Returns the version string (first line) or an error
func ReadVersion(projectRoot string) (string, string, error) {
	versionFilePath := filepath.Join(projectRoot, VersionFileName)

	file, err := os.Open(versionFilePath)
	if err != nil {
		return "", "", fmt.Errorf("fenv-fvm: failed to read .flutter-version")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if scanner.Err() != nil {
			return "", "", fmt.Errorf("fenv-fvm: failed to read .flutter-version")
		}
		return "", "", fmt.Errorf("fenv-fvm: failed to read .flutter-version")
	}

	version := strings.TrimSpace(scanner.Text())
	if version == "" {
		return "", "", fmt.Errorf("fenv-fvm: failed to read .flutter-version")
	}

	return version, versionFilePath, nil
}

// WriteVersion writes the Flutter version to .flutter-version file in dir
func WriteVersion(dir, version string) error {
	versionFilePath := filepath.Join(dir, VersionFileName)

	file, err := os.Create(versionFilePath)
	if err != nil {
		return fmt.Errorf("fenv-fvm: failed to write .flutter-version: %w", err)
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, version)
	if err != nil {
		return fmt.Errorf("fenv-fvm: failed to write .flutter-version: %w", err)
	}

	return nil
}