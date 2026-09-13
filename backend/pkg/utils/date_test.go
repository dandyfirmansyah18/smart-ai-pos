package utils_test

import (
	"testing"

	"github.com/pos-backend/pkg/utils"
)

func TestParseFlexibleDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // year
	}{
		{name: "RFC3339", input: "2026-09-09T10:00:00Z", expected: 2026},
		{name: "YYYY-MM-DD", input: "2025-12-25", expected: 2025},
		{name: "DD/MM/YYYY", input: "15/08/2024", expected: 2024},
		{name: "Empty String", input: "", expected: 0}, // returns time.Now()
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.ParseFlexibleDate(tt.input)
			if tt.expected != 0 && got.Year() != tt.expected {
				t.Errorf("expected year %d, got %d", tt.expected, got.Year())
			}
			if got.IsZero() {
				t.Errorf("expected non-zero time")
			}
		})
	}
}
