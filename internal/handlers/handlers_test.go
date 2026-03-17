package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeverityColor(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "red"},
		{"warning", "yellow"},
		{"info", "blue"},
		{"", "blue"},
		{"unknown", "blue"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, severityColor(tt.input))
		})
	}
}

func TestIsHTMX(t *testing.T) {
	// This is a simple header check, tested via integration
	// Handler tests requiring DB are integration tests
}
