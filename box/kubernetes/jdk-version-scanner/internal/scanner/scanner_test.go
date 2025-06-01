package scanner

import (
	"testing"
	"time"
)

func TestParseJavaVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Java 8",
			input:    `java version "1.8.0_292"`,
			expected: "1.8.0_292",
		},
		{
			name:     "Java 11",
			input:    `openjdk version "11.0.16" 2022-07-19`,
			expected: "11.0.16",
		},
		{
			name:     "Java 17",
			input:    `openjdk version "17.0.2" 2022-01-18`,
			expected: "17.0.2",
		},
		{
			name: "Multi-line output",
			input: `openjdk version "11.0.16" 2022-07-19
OpenJDK Runtime Environment (build 11.0.16+8-post-Ubuntu-0ubuntu120.04)
OpenJDK 64-Bit Server VM (build 11.0.16+8-post-Ubuntu-0ubuntu120.04, mixed mode, sharing)`,
			expected: "11.0.16",
		},
		{
			name:     "No version found",
			input:    "No java command found",
			expected: "Unknown",
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseJavaVersion(tt.input)
			if result != tt.expected {
				t.Errorf("parseJavaVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsDaemonSet(t *testing.T) {
	scanner := &Scanner{
		config: &Config{},
	}

	tests := []struct {
		name     string
		pod      Pod
		expected bool
	}{
		{
			name: "DaemonSet pod",
			pod: Pod{
				OwnerReferences: []OwnerReference{
					{Kind: "DaemonSet"},
				},
			},
			expected: true,
		},
		{
			name: "Deployment pod",
			pod: Pod{
				OwnerReferences: []OwnerReference{
					{Kind: "ReplicaSet"},
				},
			},
			expected: false,
		},
		{
			name: "Multiple owners with DaemonSet",
			pod: Pod{
				OwnerReferences: []OwnerReference{
					{Kind: "ReplicaSet"},
					{Kind: "DaemonSet"},
				},
			},
			expected: true,
		},
		{
			name:     "No owners",
			pod:      Pod{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.isDaemonSet(tt.pod)
			if result != tt.expected {
				t.Errorf("isDaemonSet() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewScanner(t *testing.T) {
	config := &Config{
		Namespaces:    []string{"default", "kube-system"},
		MaxGoroutines: 10,
		Timeout:       30 * time.Second,
		SkipDaemonSet: true,
		Verbose:       false,
	}

	scanner := New(config)

	if scanner == nil {
		t.Fatal("New() returned nil")
	}

	if scanner.config != config {
		t.Error("New() did not set config correctly")
	}
}
