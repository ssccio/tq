# tq Quick Start Guide

Get up and running with `tq` in 5 minutes!

## Installation

```bash
# Clone the repository
git clone https://github.com/ssccio/tq.git
cd tq

# Build
go build -o tq cmd/tq/main.go

# Or use Task
task build
```

## Your First TOON Conversion

### 1. Convert JSON to TOON

```bash
# Try it with the example file
./tq -i json -o toon examples/data.json
```

You'll see output like:
```toon
users[3]{id,name,role,lastLogin,active}:
  1,Alice,admin,"2025-01-15T10:30:00Z",true
  2,Bob,user,"2025-01-14T15:22:00Z",true
  3,Charlie,user,"2025-01-13T09:45:00Z",false
metadata:
  version: 1.0
  generated: "2025-01-15T12:00:00Z"
```

**Notice how much more compact it is!** ✨

### 2. Convert it Back

```bash
./tq -i toon -o json examples/data.toon
```

### 3. Use in a Pipeline

```bash
echo '{"name":"Alice","role":"admin","active":true}' | ./tq -i json -o toon
```

Output:
```toon
name: Alice
role: admin
active: true
```

## Understanding TOON Format

### Simple Objects
```json
{"id": 123, "name": "Ada", "active": true}
```
becomes:
```toon
id: 123
name: Ada
active: true
```

### Arrays of Objects (Tabular)
```json
{
  "users": [
    {"id": 1, "name": "Alice"},
    {"id": 2, "name": "Bob"}
  ]
}
```
becomes:
```toon
users[2]{id,name}:
  1,Alice
  2,Bob
```

**This is where you save tokens!** The field names appear once in the header instead of repeated for each row.

### Primitive Arrays
```json
{"tags": ["admin", "ops", "dev"]}
```
becomes:
```toon
tags[3]: admin,ops,dev
```

## Common Use Cases

### Use Case 1: Reduce LLM Costs

Before sending data to an LLM:
```bash
# Convert to TOON first
curl https://api.example.com/data | tq -i json -o toon > llm-input.toon

# Save 30-60% on tokens!
```

### Use Case 2: Configuration Files

```bash
# Store config in TOON format (more readable, fewer tokens)
cat config.json | tq -i json -o toon > config.toon

# Read it when needed
tq -i toon -o json config.toon
```

### Use Case 3: Data Processing Pipeline

```bash
# JSON → TOON → process → JSON
cat data.json | \
  tq -i json -o toon | \
  process-with-llm | \
  tq -i toon -o json > result.json
```

## Tips

1. **Omit the query for simple conversion**
   ```bash
   tq -i json -o toon file.json    # Just works!
   ```

2. **Pipe from stdin**
   ```bash
   cat file.json | tq -i json -o toon
   ```

3. **Pretty-print TOON**
   ```bash
   tq '.' data.toon
   ```

4. **Custom formatting**
   ```bash
   # Tab-separated
   tq -i json -o toon --delimiter "\t" data.json

   # Different indentation
   tq -i json -o toon --indent 4 data.json
   ```

## Current Limitations

The query engine is in early development. These work:

✅ Format conversion (JSON/YAML ↔ TOON)
✅ Identity query (`.`)
✅ Simple field access (`.field`)

These are being worked on:

⚠️ Array iteration (`.users[]`)
⚠️ Chained operations (`.users[0].name`)
⚠️ Filtering (`select()`)
⚠️ Advanced jq operations

See [STATUS.md](STATUS.md) for full details.

## Next Steps

- Read the [README](README.md) for complete documentation
- Check [EXAMPLES.md](examples/EXAMPLES.md) for more use cases
- See [STATUS.md](STATUS.md) for what's working
- Contribute! See [CONTRIBUTING.md](CONTRIBUTING.md)

## Getting Help

- 📖 [Full Documentation](README.md)
- 🐛 [Report Issues](https://github.com/ssccio/tq/issues)
- 💡 [Feature Requests](https://github.com/ssccio/tq/issues)

---

**Welcome to tq - making structured data more efficient for the AI age!** 🚀
