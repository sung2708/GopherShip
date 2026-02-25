package main

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Formatter defines the interface for CLI output formatting.
type Formatter interface {
	Format(w io.Writer, v interface{}) error
}

// TableFormatter outputs data in a human-readable table.
type TableFormatter struct{}

func (f *TableFormatter) Format(w io.Writer, v interface{}) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	switch data := v.(type) {
	case map[string]interface{}:
		// Deterministic order (AC-Review #2)
		keys := []string{"Somatic Zone", "Pressure Score", "Memory Usage", "Heap Objects"}
		for _, k := range keys {
			if val, ok := data[k]; ok {
				fmt.Fprintf(tw, "%s:\t%v\n", k, val)
			}
		}
		// Catch-all for any other keys
		for k, val := range data {
			found := false
			for _, known := range keys {
				if k == known {
					found = true
					break
				}
			}
			if !found {
				fmt.Fprintf(tw, "%s:\t%v\n", k, val)
			}
		}
	default:
		fmt.Fprintf(tw, "%v\n", v)
	}

	return tw.Flush()
}

// JSONFormatter outputs data in valid JSON.
type JSONFormatter struct{}

func (f *JSONFormatter) Format(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// YAMLFormatter outputs data in valid YAML.
type YAMLFormatter struct{}

func (f *YAMLFormatter) Format(w io.Writer, v interface{}) error {
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(v)
}

// NewFormatter returns the appropriate formatter for the given format string.
func NewFormatter(format string) (Formatter, error) {
	switch format {
	case "table":
		return &TableFormatter{}, nil
	case "json":
		return &JSONFormatter{}, nil
	case "yaml":
		return &YAMLFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
