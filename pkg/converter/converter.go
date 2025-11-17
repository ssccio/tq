package converter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ssccio/tq/pkg/toon"
	"gopkg.in/yaml.v3"
)

// Options for format conversion
type Options struct {
	InputFormat  string
	OutputFormat string
	Indent       int
	UseTab       bool
	Delimiter    string
	Compact      bool
	RawOutput    bool
	ShowStats    bool
	MaxInputSize int64 // Maximum input size in bytes (0 = unlimited)
}

// Converter handles format conversion
type Converter struct {
	opts Options
}

// New creates a new converter
func New(opts Options) *Converter {
	return &Converter{opts: opts}
}

// Read reads and parses input in the specified format
func (c *Converter) Read(r io.Reader) (interface{}, error) {
	if r == nil {
		return nil, nil
	}

	// Apply size limit if configured
	if c.opts.MaxInputSize > 0 {
		r = io.LimitReader(r, c.opts.MaxInputSize)
	}

	// For JSON and YAML, use streaming decoders
	// Peek at first bytes to detect format
	data := make([]byte, 0, 512)
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	data = append(data, buf[:n]...)

	format := c.opts.InputFormat
	if format == "auto" {
		format = detectFormat(data)
	}

	// Create MultiReader with peeked data + remaining
	fullReader := io.MultiReader(strings.NewReader(string(data)), r)

	switch format {
	case "json":
		return c.readJSONStream(fullReader)
	case "yaml":
		return c.readYAMLStream(fullReader)
	case "toon":
		// Use streaming reader for TOON as well
		return toon.DecodeReader(bufio.NewReader(fullReader))
	default:
		return nil, fmt.Errorf("unsupported input format: %s", format)
	}
}

// Write writes data in the specified output format
func (c *Converter) Write(w io.Writer, data interface{}) error {
	switch c.opts.OutputFormat {
	case "json":
		return c.writeJSON(w, data)
	case "yaml":
		return c.writeYAML(w, data)
	case "toon":
		return c.writeTOON(w, data)
	default:
		return fmt.Errorf("unsupported output format: %s", c.opts.OutputFormat)
	}
}

func (c *Converter) readJSON(data []byte) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return result, nil
}

func (c *Converter) readJSONStream(r io.Reader) (interface{}, error) {
	var result interface{}
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return result, nil
}

func (c *Converter) readYAML(data []byte) (interface{}, error) {
	var result interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return result, nil
}

func (c *Converter) readYAMLStream(r io.Reader) (interface{}, error) {
	var result interface{}
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return result, nil
}

func (c *Converter) writeJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	if !c.opts.Compact {
		encoder.SetIndent("", strings.Repeat(" ", c.opts.Indent))
	}
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

func (c *Converter) writeYAML(w io.Writer, data interface{}) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(c.opts.Indent)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}
	return nil
}

func (c *Converter) writeTOON(w io.Writer, data interface{}) error {
	opts := toon.Options{
		Indent:    c.opts.Indent,
		Delimiter: c.opts.Delimiter,
		UseTab:    c.opts.UseTab,
	}

	output, err := toon.Encode(data, opts)
	if err != nil {
		return fmt.Errorf("failed to encode TOON: %w", err)
	}

	if _, err := w.Write([]byte(output)); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if !strings.HasSuffix(output, "\n") {
		if _, err := w.Write([]byte("\n")); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	// Show token statistics if requested
	if c.opts.ShowStats {
		c.showTokenStats(data, output)
	}

	return nil
}

func (c *Converter) showTokenStats(original interface{}, toonOutput string) {
	// Compare JSON vs TOON token usage
	jsonData, err := json.Marshal(original)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n--- Could not generate token statistics: %v ---\n", err)
		return
	}

	jsonTokens := estimateTokens(string(jsonData))
	toonTokens := estimateTokens(toonOutput)

	if jsonTokens == 0 {
		fmt.Fprintf(os.Stderr, "\n--- Token Statistics ---\n")
		fmt.Fprintf(os.Stderr, "JSON tokens: 0\n")
		fmt.Fprintf(os.Stderr, "TOON tokens: %d\n", toonTokens)
		return
	}

	reduction := float64(jsonTokens-toonTokens) / float64(jsonTokens) * 100

	// Write to stderr so it doesn't interfere with stdout output
	fmt.Fprintf(os.Stderr, "\n--- Token Statistics ---\n")
	fmt.Fprintf(os.Stderr, "JSON tokens: %d\n", jsonTokens)
	fmt.Fprintf(os.Stderr, "TOON tokens: %d\n", toonTokens)
	fmt.Fprintf(os.Stderr, "Reduction: %.1f%%\n", reduction)
}

func estimateTokens(s string) int {
	// Rough estimate: ~4 characters per token
	// This is a simplification; real tokenization is more complex
	return len(s) / 4
}

func detectFormat(data []byte) string {
	trimmed := strings.TrimSpace(string(data))

	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return "json"
	}

	if strings.Contains(trimmed, "---") || strings.Contains(trimmed, ":") {
		// Could be YAML or TOON
		// Check for TOON-specific patterns
		if strings.Contains(trimmed, "[") && strings.Contains(trimmed, "]{") {
			return "toon"
		}
		return "yaml"
	}

	return "json" // default
}
