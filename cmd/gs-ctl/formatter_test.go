package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFormatters(t *testing.T) {
	data := map[string]interface{}{
		"Somatic Zone":   "GREEN",
		"Pressure Score": "0%",
		"Memory Usage":   "1024 bytes",
		"Heap Objects":   10,
	}

	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, output []byte)
	}{
		{
			"JSON",
			"json",
			func(t *testing.T, output []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(output, &result); err != nil {
					t.Fatalf("Failed to unmarshal JSON: %v", err)
				}
				if result["Somatic Zone"] != "GREEN" {
					t.Errorf("Expected Somatic Zone GREEN, got %v", result["Somatic Zone"])
				}
			},
		},
		{
			"YAML",
			"yaml",
			func(t *testing.T, output []byte) {
				var result map[string]interface{}
				if err := yaml.Unmarshal(output, &result); err != nil {
					t.Fatalf("Failed to unmarshal YAML: %v", err)
				}
				if result["Somatic Zone"] != "GREEN" {
					t.Errorf("Expected Somatic Zone GREEN, got %v", result["Somatic Zone"])
				}
			},
		},
		{
			"Table",
			"table",
			func(t *testing.T, output []byte) {
				s := string(output)
				// Check for deterministic order (AC-Review #2)
				expected := []string{"Somatic Zone:", "Pressure Score:", "Memory Usage:", "Heap Objects:"}
				lastIdx := -1
				for _, substr := range expected {
					idx := strings.Index(s, substr)
					if idx == -1 {
						t.Errorf("Table output missing %s", substr)
					}
					if idx < lastIdx {
						t.Errorf("Table output order incorrect: %s appeared before previous key", substr)
					}
					lastIdx = idx
				}
				if !strings.Contains(s, "GREEN") {
					t.Errorf("Table output missing GREEN")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFormatter(tt.format)
			if err != nil {
				t.Fatalf("Failed to create formatter: %v", err)
			}

			var buf bytes.Buffer
			if err := f.Format(&buf, data); err != nil {
				t.Fatalf("Format failed: %v", err)
			}

			tt.check(t, buf.Bytes())
		})
	}
}
