package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ssccio/tq/pkg/converter"
	"github.com/ssccio/tq/pkg/query"
)

// ErrExitWithStatus is returned when exit-status flag is set and result is false/nil
var ErrExitWithStatus = errors.New("exit with status 1")

var (
	inputFormat  string
	outputFormat string
	rawOutput    bool
	compact      bool
	slurp        bool
	nullInput    bool
	exitStatus   bool
	fromFile     string
	indent       int
	useTab       bool
	delimiter    string
	showStats    bool
	showCompare  bool
)

func Execute(version, commit, date string) error {
	rootCmd := &cobra.Command{
		Use:   "tq [flags] [query] [files...]",
		Short: "TOON query tool - jq/yq for TOON format",
		Long: `tq is a command-line tool for querying and transforming structured data
using the TOON (Token-Oriented Object Notation) format.

TOON reduces token usage by 30-60% compared to JSON while maintaining
readability and structure - perfect for LLM workflows.`,
		Example: `  # Query TOON data
  tq '.users[0].name' data.toon

  # Convert JSON to TOON
  tq -i json -o toon data.json

  # Query with filter
  echo '{"users":[{"name":"Alice","age":30}]}' | tq '.users[] | select(.age > 25)'

  # Show token statistics
  tq -i json -o toon --stats data.json`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		RunE:    run,
		SilenceUsage: true,
	}

	// Input/Output flags
	rootCmd.Flags().StringVarP(&inputFormat, "input-format", "i", "auto",
		"Input format: auto, json, yaml, toon")
	rootCmd.Flags().StringVarP(&outputFormat, "output-format", "o", "toon",
		"Output format: toon, json, yaml")

	// Output options
	rootCmd.Flags().BoolVarP(&rawOutput, "raw-output", "r", false,
		"Output raw text, not TOON/JSON strings")
	rootCmd.Flags().BoolVarP(&compact, "compact-output", "c", false,
		"Compact output (no pretty-printing)")

	// Input options
	rootCmd.Flags().BoolVarP(&slurp, "slurp", "s", false,
		"Read entire input into single array")
	rootCmd.Flags().BoolVarP(&nullInput, "null-input", "n", false,
		"Don't read input, use null as input")

	// Query options
	rootCmd.Flags().BoolVarP(&exitStatus, "exit-status", "e", false,
		"Set exit code based on output")
	rootCmd.Flags().StringVarP(&fromFile, "from-file", "f", "",
		"Read query from file")

	// TOON-specific options
	rootCmd.Flags().IntVar(&indent, "indent", 2,
		"Indentation spaces")
	rootCmd.Flags().BoolVar(&useTab, "tab", false,
		"Use tabs for indentation")
	rootCmd.Flags().StringVar(&delimiter, "delimiter", ",",
		"TOON delimiter character")
	rootCmd.Flags().BoolVar(&showStats, "stats", false,
		"Show token usage statistics (JSON vs TOON)")
	rootCmd.Flags().BoolVar(&showCompare, "compare", false,
		"Show format comparison (JSON/YAML/TOON sizes)")

	return rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	// Parse query and input files
	var queryStr string
	var inputFiles []string

	if fromFile != "" {
		// Sanitize file path to prevent path traversal
		cleanPath := filepath.Clean(fromFile)
		if strings.Contains(cleanPath, "..") {
			return fmt.Errorf("invalid query file path: path traversal not allowed")
		}
		data, err := os.ReadFile(cleanPath)
		if err != nil {
			return fmt.Errorf("failed to read query file: %w", err)
		}
		queryStr = string(data)
		inputFiles = args
	} else if len(args) > 0 {
		// If the first arg looks like a file path and exists, treat it as input file with default query
		if len(args) == 1 && !strings.HasPrefix(args[0], ".") && !strings.HasPrefix(args[0], "[") {
			if _, err := os.Stat(args[0]); err == nil {
				queryStr = "."
				inputFiles = args
			} else {
				// Not a file, treat as query
				queryStr = args[0]
				inputFiles = args[1:]
			}
		} else {
			queryStr = args[0]
			inputFiles = args[1:]
		}
	} else {
		queryStr = "."
	}

	// Determine input source
	var input io.Reader
	if nullInput {
		input = nil
	} else if len(inputFiles) == 0 {
		input = os.Stdin
	} else {
		// For now, just use the first file
		// TODO: Support multiple files
		// Sanitize file path to prevent path traversal
		cleanPath := filepath.Clean(inputFiles[0])
		if strings.Contains(cleanPath, "..") {
			return fmt.Errorf("invalid input file path: path traversal not allowed")
		}
		f, err := os.Open(cleanPath)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer f.Close()
		input = f
	}

	// Create converter
	conv := converter.New(converter.Options{
		InputFormat:  inputFormat,
		OutputFormat: outputFormat,
		Indent:       indent,
		UseTab:       useTab,
		Delimiter:    delimiter,
		Compact:      compact,
		RawOutput:    rawOutput,
		ShowStats:    showStats,
		ShowCompare:  showCompare,
		Slurp:        slurp,
		MaxInputSize: 100 * 1024 * 1024, // 100MB default limit
	})

	// Read and parse input
	var data interface{}
	var err error
	if nullInput {
		// null-input mode: use null (nil) as input
		data = nil
	} else {
		// Normal mode: read from input
		data, err = conv.Read(input)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
	}

	// Execute query
	engine := query.New()
	result, err := engine.Execute(queryStr, data)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	// Write output
	if err := conv.Write(os.Stdout, result); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Handle exit status
	if exitStatus {
		if result == nil || result == false {
			return ErrExitWithStatus
		}
	}

	return nil
}
