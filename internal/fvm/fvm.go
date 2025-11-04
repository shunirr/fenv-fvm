package fvm

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetCacheDir returns the fvm cache directory in priority order:
// 1. $FVM_CACHE_PATH environment variable
// 2. $HOME/fvm/versions
// 3. $HOME/Library/Application Support/fvm/versions
func GetCacheDir() (string, error) {
	// Check FVM_CACHE_PATH environment variable
	if cachePath := os.Getenv("FVM_CACHE_PATH"); cachePath != "" {
		if info, err := os.Stat(cachePath); err == nil && info.IsDir() {
			return cachePath, nil
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("fenv-fvm: failed to get home directory")
	}

	// Try $HOME/fvm/versions (common default)
	commonPath := filepath.Join(home, "fvm", "versions")
	if info, err := os.Stat(commonPath); err == nil && info.IsDir() {
		return commonPath, nil
	}

	// Try $HOME/Library/Application Support/fvm/versions (official default for macOS)
	officialPath := filepath.Join(home, "Library", "Application Support", "fvm", "versions")
	if info, err := os.Stat(officialPath); err == nil && info.IsDir() {
		return officialPath, nil
	}

	return "", fmt.Errorf("fenv-fvm: fvm cache directory not found")
}

// GetSDKPath returns the path to the Flutter SDK for the given version
func GetSDKPath(version string) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}

	sdkPath := filepath.Join(cacheDir, version)

	// Check if SDK directory exists
	if info, err := os.Stat(sdkPath); err != nil || !info.IsDir() {
		return "", fmt.Errorf("fenv-fvm: Flutter SDK '%s' is not installed\nPlease run: fvm install %s", version, version)
	}

	return sdkPath, nil
}

// ResolveBinary resolves the path to flutter or dart binary
// binaryName should be "flutter" or "dart"
func ResolveBinary(version, binaryName string) (string, error) {
	sdkPath, err := GetSDKPath(version)
	if err != nil {
		return "", err
	}

	binaryPath := filepath.Join(sdkPath, "bin", binaryName)

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

// VerifySDKExists checks if the SDK for the given version exists
func VerifySDKExists(version string) error {
	_, err := GetSDKPath(version)
	return err
}
