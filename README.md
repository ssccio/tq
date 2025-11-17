# tq - TOON Query Tool

A command-line tool for querying and transforming structured data using the TOON (Token-Oriented Object Notation) format, with feature parity to `jq` and `yq`.

## What is TOON?

TOON (Token-Oriented Object Notation) is a compact, human-readable serialization format designed specifically for Large Language Models (LLMs). It reduces token usage by 30-60% compared to JSON while maintaining readability and structure.

### Example TOON Format

```toon
users[3]{id,name,role,lastLogin}:
  1,Alice,admin,2025-01-15T10:30:00Z
  2,Bob,user,2025-01-14T15:22:00Z
  3,Charlie,user,2025-01-13T09:45:00Z
```

The same data in JSON would be significantly more verbose and consume more tokens.

## Why tq?

- **Token Efficiency**: 30-60% reduction in token usage for LLM interactions
- **Format Conversion**: Seamlessly convert between JSON, YAML, and TOON
- **Query Power**: Full jq-like query syntax for data manipulation
- **LLM-Optimized**: Perfect for AI workflows and structured data passing

## Installation

```bash
go install github.com/ssccio/tq/cmd/tq@latest
```

Or build from source:

```bash
git clone https://github.com/ssccio/tq.git
cd tq
go build -o tq cmd/tq/main.go
```

## Quick Start

### Basic Usage

```bash
# Query TOON data
tq '.users[0].name' data.toon

# Convert JSON to TOON
tq -i json -o toon data.json

# Convert YAML to TOON
tq -i yaml -o toon data.yaml

# Query and transform
echo '{"users":[{"name":"Alice","age":30}]}' | tq -i json '.users[] | select(.age > 25)'

# Pretty-print TOON
tq '.' data.toon

# Read multiple JSON objects into array (slurp mode)
echo -e '{"id":1}\n{"id":2}\n{"id":3}' | tq --slurp '.'

# Generate data without input (null-input mode)
tq --null-input 'range(10)'

# Compare format sizes (show token savings)
tq --compare -i json -o toon data.json
```

### Format Conversion

```bash
# JSON → TOON (reduce token usage)
tq -i json -o toon input.json > output.toon

# TOON → JSON (for traditional tools)
tq -i toon -o json input.toon > output.json

# YAML → TOON
tq -i yaml -o toon config.yaml > config.toon

# Chain conversions with queries
tq -i json -o toon '.data.items[] | select(.active == true)' input.json
```

## Query Syntax

`tq` supports jq-compatible query syntax:

### Basic Selectors

```bash
# Identity (pretty-print)
tq '.'

# Field access
tq '.field'
tq '.nested.field'

# Array indexing
tq '.[0]'
tq '.items[2]'
tq '.[-1]'  # last element
```

### Array Operations

```bash
# Iterate array
tq '.[]'
tq '.items[]'

# Map over array
tq '.items[] | .name'

# Filter
tq '.items[] | select(.price > 100)'

# Collect results
tq '[.items[] | .name]'
```

### Transformations

```bash
# Create new objects (basic support)
tq '{name: .user, total: .amount}'

# Multiple outputs
tq '.users[] | {id, name}'
```

### Built-in Functions

```bash
# Array/Object functions
tq '. | length()'              # Get length
tq '. | keys()'                # Get object keys or array indices
tq '. | values()'              # Get object values
tq '. | type()'                # Get type (array, object, string, number, boolean, null)

# Array operations
tq '. | sort()'                # Sort array
tq '. | reverse()'             # Reverse array
tq '. | unique()'              # Get unique elements
tq '. | flatten()'             # Flatten array (depth 1)
tq '. | flatten(2)'            # Flatten array to depth 2
tq '. | first()'               # Get first element
tq '. | first(3)'              # Get first 3 elements
tq '. | last()'                # Get last element
tq '. | last(2)'               # Get last 2 elements
tq '. | map(.name)'            # Map expression over array
tq '. | sort_by(.age)'         # Sort array by field
tq '. | group_by(.category)'   # Group array by field

# Math functions
tq '. | add()'                 # Sum all numbers in array
tq '. | min()'                 # Find minimum value
tq '. | max()'                 # Find maximum value
tq '. | floor()'               # Round down
tq '. | ceil()'                # Round up
tq '. | round()'               # Round to nearest integer

# Range generation
tq 'range(5)'                  # Generate [0,1,2,3,4]
tq 'range(2;5)'                # Generate [2,3,4]
tq 'range(0;10;2)'             # Generate [0,2,4,6,8]

# String functions
tq '. | split(",")'            # Split string by delimiter
tq '. | join(" ")'             # Join array with delimiter
tq '. | startswith("prefix")'  # Check if starts with string
tq '. | endswith("suffix")'    # Check if ends with string
tq '. | contains("substring")' # Check if contains string

# Object operations
tq '. | has("field")'          # Check if object has key
```

## Command-Line Options

```
Usage: tq [options] [query] [files...]

Options:
  -i, --input-format FORMAT     Input format: auto, json, yaml, toon (default: auto)
  -o, --output-format FORMAT    Output format: toon, json, yaml (default: toon)
  -r, --raw-output              Output raw text, not TOON/JSON strings
  -c, --compact-output          Compact output (no pretty-printing)
  -s, --slurp                   Read entire input into single array
  -n, --null-input              Don't read input, use null as input
  -e, --exit-status             Set exit code based on output
  -f, --from-file FILE          Read query from file
      --indent N                Indentation spaces (default: 2)
      --tab                     Use tabs for indentation
      --delimiter CHAR          TOON delimiter character (default: ,)
      --stats                   Show token usage statistics (JSON vs TOON)
      --compare                 Show format comparison (JSON/YAML/TOON sizes)
  -h, --help                    Show this help message
  -v, --version                 Show version information

Query Syntax:
  .                   Identity (pass-through)
  .field              Access field
  .field.nested       Access nested field
  .[0]                Array index
  .[]                 Array/object iterator
  |                   Pipe (chain operations)
  select(expr)        Filter by condition
  map(expr)           Transform array elements
  {key: value}        Construct object
  [expr]              Construct array
```

## Use Cases

### 1. LLM Data Passing

Reduce token costs when passing structured data to LLMs:

```bash
# Convert large JSON dataset to TOON and see the savings
tq -i json -o toon --compare large-dataset.json

# Example output:
# --- Format Comparison ---
# JSON:  338 bytes (84 tokens estimated)
# YAML:  360 bytes (90 tokens estimated)
# TOON:  234 bytes (58 tokens estimated)
#
# Input:  JSON (84 tokens)
# Output: TOON (58 tokens)
# Token savings: 31.0%
```

### 2. Configuration Management

```bash
# Extract specific config values
tq '.database.connections[] | select(.primary == true)' config.toon

# Convert between formats for different tools
tq -i yaml -o toon k8s-config.yaml
```

### 3. Data Pipeline

```bash
# Filter and transform in pipeline
cat data.json | tq -i json '.items[] | select(.status == "active")' | tq -o yaml
```

### 4. API Response Processing

```bash
# Fetch API data and convert to TOON
curl https://api.example.com/users | tq -i json -o toon '.data.users'
```

## TOON Format Specification

TOON uses a compact representation optimized for token efficiency:

### Objects

```toon
id: 123
name: Ada
active: true
```

### Nested Objects

```toon
user:
  id: 123
  name: Ada
```

### Primitive Arrays

```toon
tags[3]: admin,ops,dev
```

### Tabular Arrays (uniform objects)

```toon
items[2]{sku,qty,price}:
  A1,2,9.99
  B2,1,14.5
```

### Mixed Arrays

```toon
items[3]:
  - 42
  - text
  - key: value
```

## Development

### Project Structure

```
tq/
├── cmd/tq/              # CLI entry point
├── pkg/
│   ├── toon/            # TOON format handling
│   ├── query/           # Query engine
│   └── converter/       # Format converters
├── internal/
│   ├── parser/          # TOON parser
│   └── serializer/      # TOON serializer
├── examples/            # Example files
└── tests/               # Test files
```

### Building

```bash
# Build
go build -o tq cmd/tq/main.go

# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Build for multiple platforms
task build  # requires Taskfile.yml
```

### Testing

```bash
# Unit tests
go test ./pkg/...

# Integration tests
go test ./tests/...

# Benchmark
go test -bench=. ./pkg/toon
```

## Comparison with jq and yq

| Feature | jq | yq | tq |
|---------|----|----|-----|
| JSON support | ✅ | ✅ | ✅ |
| YAML support | ❌ | ✅ | ✅ |
| TOON support | ❌ | ❌ | ✅ |
| Query syntax | ✅ | ✅ (jq-like) | ✅ (jq-compatible) |
| Token efficiency | Standard | Standard | 30-60% reduction |
| LLM-optimized | ❌ | ❌ | ✅ |

## Token Usage Examples

### Before (JSON - 156 tokens)

```json
{
  "users": [
    {"id": 1, "name": "Alice", "role": "admin", "lastLogin": "2025-01-15T10:30:00Z"},
    {"id": 2, "name": "Bob", "role": "user", "lastLogin": "2025-01-14T15:22:00Z"},
    {"id": 3, "name": "Charlie", "role": "user", "lastLogin": "2025-01-13T09:45:00Z"}
  ]
}
```

### After (TOON - 67 tokens)

```toon
users[3]{id,name,role,lastLogin}:
  1,Alice,admin,2025-01-15T10:30:00Z
  2,Bob,user,2025-01-14T15:22:00Z
  3,Charlie,user,2025-01-13T09:45:00Z
```

**Token savings: 57%**

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

- [x] Core TOON parser and serializer
- [x] Basic query engine
- [x] Format conversion (JSON/YAML/TOON)
- [ ] Full jq syntax compatibility
- [ ] Performance optimizations
- [ ] Streaming support for large files
- [ ] Custom TOON encoding options (delimiters, indentation)
- [ ] Interactive mode (like ijq)
- [ ] Shell completions
- [ ] Syntax highlighting
- [ ] Plugin system

## License

MIT License - see [LICENSE](LICENSE) file for details

## Acknowledgments

- [TOON Format Specification](https://github.com/toon-format/toon)
- [jq](https://github.com/jqlang/jq) - Inspiration for query syntax
- [yq](https://github.com/mikefarah/yq) - YAML processing patterns

## Links

- **Documentation**: [https://github.com/ssccio/tq](https://github.com/ssccio/tq)
- **Issues**: [https://github.com/ssccio/tq/issues](https://github.com/ssccio/tq/issues)
- **TOON Format**: [https://github.com/toon-format/toon](https://github.com/toon-format/toon)

---

**tq** - Making structured data more efficient for the AI age.
