package shim

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetShimRoot returns the shim root directory
// Defaults to ~/.fenv-fvm if FENV_FVM_ROOT is not set
func GetShimRoot() (string, error) {
	if root := os.Getenv("FENV_FVM_ROOT"); root != "" {
		return root, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".fenv-fvm"), nil
}

// InitializeShims creates the shims directory and symlinks for flutter and dart
func InitializeShims(mainBinaryPath string) error {
	shimRoot, err := GetShimRoot()
	if err != nil {
		return fmt.Errorf("fenv-fvm: failed to initialize shims directory: %w", err)
	}

	shimsDir := filepath.Join(shimRoot, "shims")

	// Create shims directory
	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return fmt.Errorf("fenv-fvm: failed to initialize shims directory: %w", err)
	}

	// Get absolute path to main binary
	absMainBinary, err := filepath.Abs(mainBinaryPath)
	if err != nil {
		return fmt.Errorf("fenv-fvm: failed to initialize shims directory: %w", err)
	}

	// Create symlinks for flutter and dart
	shimNames := []string{"flutter", "dart"}
	for _, shimName := range shimNames {
		shimPath := filepath.Join(shimsDir, shimName)

		// Remove existing shim if it exists
		if _, err := os.Lstat(shimPath); err == nil {
			if err := os.Remove(shimPath); err != nil {
				return fmt.Errorf("fenv-fvm: failed to initialize shims directory: %w", err)
			}
		}

		// Create symlink
		if err := os.Symlink(absMainBinary, shimPath); err != nil {
			return fmt.Errorf("fenv-fvm: failed to initialize shims directory: %w", err)
		}
	}

	return nil
}

// PrintPathSetup prints the PATH setup snippet to stdout
func PrintPathSetup() error {
	shimRoot, err := GetShimRoot()
	if err != nil {
		return err
	}

	fmt.Printf("export FENV_FVM_ROOT=\"%s\"\n", shimRoot)
	fmt.Printf("export PATH=\"%s/shims:$PATH\"\n", shimRoot)

	return nil
}