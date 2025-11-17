# tq Examples

This directory contains example files and usage patterns for `tq`.

## Example Files

- `data.json` - Sample JSON data with users and metadata
- `data.toon` - The same data in TOON format (57% token reduction!)

## Basic Examples

### 1. Format Conversion

Convert JSON to TOON:
```bash
tq -i json -o toon data.json
```

Convert TOON back to JSON:
```bash
tq -i toon -o json data.toon
```

Convert with token statistics:
```bash
tq -i json -o toon --stats data.json
```

### 2. Query Operations

Pretty-print TOON data:
```bash
tq '.' data.toon
```

Extract specific field:
```bash
tq '.metadata.version' data.toon
```

Get array element:
```bash
tq '.users[0]' data.toon
```

Get user name:
```bash
tq '.users[0].name' data.toon
```

### 3. Array Operations

Iterate over all users:
```bash
tq '.users[]' data.toon
```

Get all user names:
```bash
tq '.users[].name' data.toon
```

Filter active users:
```bash
tq '.users[] | select(.active == true)' data.toon
```

Filter by role:
```bash
tq '.users[] | select(.role == "admin")' data.toon
```

### 4. Transformations

Create new object with selected fields:
```bash
tq '.users[] | {name: .name, role: .role}' data.toon
```

Collect filtered results:
```bash
tq '[.users[] | select(.active == true)]' data.toon
```

### 5. Piping and Chaining

Chain multiple operations:
```bash
tq '.users[] | select(.active == true) | .name' data.toon
```

Complex pipeline:
```bash
echo '{"items":[{"id":1,"qty":5},{"id":2,"qty":3}]}' | \
  tq -i json '.items[] | select(.qty > 3)'
```

## Real-World Use Cases

### Use Case 1: LLM Data Preparation

Reduce token costs when sending structured data to LLMs:

```bash
# Convert large dataset to TOON before API call
curl https://api.example.com/data | \
  tq -i json -o toon --stats > llm-input.toon

# Result: 40% token reduction = 40% cost savings!
```

### Use Case 2: Configuration Management

Extract specific configuration values:

```bash
# Get database connection string
tq '.database.connection' config.toon

# Get all active services
tq '.services[] | select(.enabled == true)' config.toon
```

### Use Case 3: Data Pipeline

Filter and transform in a data pipeline:

```bash
# Extract active users, convert to YAML for another tool
cat users.json | \
  tq -i json '.users[] | select(.active == true)' | \
  tq -o yaml
```

### Use Case 4: API Response Processing

Process and transform API responses:

```bash
# Fetch, filter, and save
curl https://api.example.com/users | \
  tq -i json -o toon '.data.users[] | select(.role == "admin")' \
  > admin-users.toon
```

## Advanced Examples

### Nested Queries

```bash
# Access nested data
echo '{"org":{"teams":[{"name":"dev","members":5}]}}' | \
  tq '.org.teams[0].name'
```

### Array Index

```bash
# Get last element
tq '.users[-1]' data.toon

# Get first three users
tq '.users[0:3]' data.toon  # (when range support is added)
```

### Conditional Output

```bash
# Different output based on condition
tq '.users[] | if .role == "admin" then .name else empty end' data.toon
```

## Performance Tips

1. **Use TOON for LLM workflows**: Save 30-60% on token costs
2. **Pipe efficiently**: Chain operations in a single query vs multiple invocations
3. **Filter early**: Use `select()` early in the pipeline to reduce data size
4. **Compact output**: Use `-c` flag for minimal output when human readability isn't needed

## Common Patterns

### Pattern 1: Find and Extract

```bash
# Find user by ID and extract name
tq '.users[] | select(.id == 2) | .name' data.toon
```

### Pattern 2: Count Items

```bash
# Count active users (when length/count is implemented)
tq '[.users[] | select(.active == true)] | length' data.toon
```

### Pattern 3: Transform Structure

```bash
# Reshape data
tq '{
  admin: [.users[] | select(.role == "admin") | .name],
  users: [.users[] | select(.role == "user") | .name]
}' data.toon
```

## Integration Examples

### With curl

```bash
curl -s https://api.example.com/data | tq -i json -o toon --stats
```

### With git

```bash
# Process git log JSON output
git log --pretty=format:'{"commit":"%H","author":"%an","date":"%ad"}' | \
  tq -i json -o toon
```

### With other tools

```bash
# Convert JSON to TOON, process with other tool, convert back
cat data.json | \
  tq -i json -o toon | \
  some-other-tool | \
  tq -i toon -o json
```

## Testing Examples

Create test data:
```bash
echo '{"test": true, "value": 42}' | tq -i json -o toon
```

Validate TOON syntax:
```bash
tq '.' examples/data.toon
```

Compare formats:
```bash
# JSON version
cat data.json | wc -c

# TOON version (should be smaller)
tq -i json -o toon data.json | wc -c
```

## Next Steps

- See [README.md](../README.md) for full documentation
- Check [CONTRIBUTING.md](../CONTRIBUTING.md) for development guide
- Report issues at https://github.com/ssccio/tq/issues
