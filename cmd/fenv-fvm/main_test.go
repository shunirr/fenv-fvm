package main

import (
	"testing"
)

func TestDetermineProgramMode(t *testing.T) {
	tests := []struct {
		name        string
		programName string
		expected    string
	}{
		{
			name:        "CLI mode with fenv-fvm",
			programName: "fenv-fvm",
			expected:    "cli",
		},
		{
			name:        "Shim mode with flutter",
			programName: "flutter",
			expected:    "shim",
		},
		{
			name:        "Shim mode with dart",
			programName: "dart",
			expected:    "shim",
		},
		{
			name:        "Unknown program name",
			programName: "unknown-program",
			expected:    "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineProgramMode(tt.programName)
			if result != tt.expected {
				t.Errorf("determineProgramMode(%q) = %q, want %q", tt.programName, result, tt.expected)
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "Valid command",
			args:     []string{"fenv-fvm", "init"},
			expected: "init",
		},
		{
			name:     "Command with arguments",
			args:     []string{"fenv-fvm", "local", "3.13.9"},
			expected: "local",
		},
		{
			name:     "No command provided",
			args:     []string{"fenv-fvm"},
			expected: "",
		},
		{
			name:     "Empty args",
			args:     []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommand(tt.args)
			if result != tt.expected {
				t.Errorf("parseCommand(%v) = %q, want %q", tt.args, result, tt.expected)
			}
		})
	}
}

func TestValidateInstallArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "Valid version argument",
			args:      []string{"3.13.9"},
			expectErr: false,
		},
		{
			name:      "Valid stable argument",
			args:      []string{"stable"},
			expectErr: false,
		},
		{
			name:      "No arguments",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "Multiple arguments",
			args:      []string{"3.13.9", "extra"},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInstallArgs(tt.args)
			if tt.expectErr && err == nil {
				t.Errorf("validateInstallArgs(%v) expected error, got nil", tt.args)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("validateInstallArgs(%v) expected no error, got %v", tt.args, err)
			}
		})
	}
}