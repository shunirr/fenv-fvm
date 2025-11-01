package fvm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CheckFvmAvailable checks if fvm is available in PATH
func CheckFvmAvailable() error {
	_, err := exec.LookPath("fvm")
	if err != nil {
		return fmt.Errorf("fenv-fvm: fvm not found in PATH")
	}
	return nil
}

// removeShimsFromPath removes the fenv-fvm shims directory from PATH
// to prevent infinite loops when fvm calls flutter/dart
func removeShimsFromPath() string {
	currentPath := os.Getenv("PATH")
	if currentPath == "" {
		return ""
	}

	// Get shim root (default: ~/.fenv-fvm)
	shimRoot := os.Getenv("FENV_FVM_ROOT")
	if shimRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return currentPath // If we can't get home, return original PATH
		}
		shimRoot = filepath.Join(home, ".fenv-fvm")
	}
	shimsDir := filepath.Join(shimRoot, "shims")

	// Split PATH and filter out shims directory
	pathDirs := strings.Split(currentPath, string(os.PathListSeparator))
	var filteredDirs []string
	for _, dir := range pathDirs {
		// Skip the shims directory
		if dir != shimsDir {
			filteredDirs = append(filteredDirs, dir)
		}
	}

	return strings.Join(filteredDirs, string(os.PathListSeparator))
}

// Install runs fvm install <version>
func Install(version string) error {
	cmd := exec.Command("fvm", "install", version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Remove shims from PATH to prevent infinite loop
	// when fvm tries to execute flutter/dart commands
	cmd.Env = append(os.Environ(), "PATH="+removeShimsFromPath())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fenv-fvm: failed to install Flutter '%s' via fvm", version)
	}

	return nil
}

// Use runs fvm use <version> in the project root directory
func Use(version, projectRoot string) error {
	cmd := exec.Command("fvm", "use", version)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Remove shims from PATH to prevent infinite loop
	// when fvm tries to execute flutter/dart commands
	cmd.Env = append(os.Environ(), "PATH="+removeShimsFromPath())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fenv-fvm: failed to prepare Flutter '%s' via fvm", version)
	}

	return nil
}

// Prepare runs both Install and Use for the given version
func Prepare(version, projectRoot string) error {
	if err := Install(version); err != nil {
		return err
	}

	if err := Use(version, projectRoot); err != nil {
		return err
	}

	return nil
}

// ResolveBinary resolves the path to flutter or dart binary
// binaryName should be "flutter" or "dart"
func ResolveBinary(projectRoot, binaryName string) (string, error) {
	binaryPath := filepath.Join(projectRoot, ".fvm", "flutter_sdk", "bin", binaryName)

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("fenv-fvm: resolved Flutter SDK is incomplete (missing bin/%s)", binaryName)
		}
		return "", err
	}

	// Return absolute path
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return "", err
	}

	return absPath, nil
}