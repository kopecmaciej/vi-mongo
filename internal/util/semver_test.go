package util

import "testing"

func TestSemverGreater(t *testing.T) {
	tests := []struct {
		a, b     string
		expected bool
	}{
		{"1.0.1", "1.0.0", true},
		{"1.1.0", "1.0.9", true},
		{"2.0.0", "1.9.9", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.0.1", false},
		{"0.9.9", "1.0.0", false},
		{"v1.2.0", "v1.1.0", true},
		{"v1.1.0", "v1.2.0", false},
		// pre-release metadata stripped
		{"0.1.35-3-g70ed907-dirty", "0.1.34", true},
		{"0.1.35-3-g70ed907-dirty", "0.1.35", false},
		{"0.0.1", "0.0.0", true},
		{"0.0.0", "0.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := SemverGreater(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("SemverGreater(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}
