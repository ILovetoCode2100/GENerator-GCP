# YAML Format Converter

The YAML Format Converter provides bidirectional conversion between the three Virtuoso test formats:

1. **Compact Format** - AI-optimized, concise syntax
2. **Simplified Format** - Human-readable, clear structure  
3. **Extended Format** - Verbose, detailed metadata

## Features

- **Automatic format detection** with confidence scoring
- **Lossless conversion** where possible
- **Validation** of source and target formats
- **Warnings** for incompatible features
- **Batch conversion** support
- **Unified intermediate representation** for reliable conversions

## Usage

### As a Library

```go
import "github.com/your-org/virtuoso-generator/pkg/api-cli/yaml-layer/converter"

// Create converter
conv := converter.NewFormatConverter()

// Convert YAML content
result, err := conv.Convert(yamlContent, detector.FormatSimplified)
if err != nil {
    log.Fatal(err)
}

// Check warnings
for _, warning := range result.Warnings {
    log.Printf("Warning: %s", warning)
}

// Use converted content
fmt.Println(string(result.Output))
```

### CLI Tool

```go
// Create CLI instance
cli := converter.NewCLI()

// Convert file
err := cli.ConvertFile("input.yaml", "output.yaml", detector.FormatCompact)

// Detect format
detection, err := cli.DetectFormat("test.yaml")
fmt.Printf("Format: %s (confidence: %.2f)\n", detection.Format, detection.Confidence)

// Analyze file
err := cli.AnalyzeFile("test.yaml")

// Show all formats
cli.ShowFormats()
```

## Format Examples

### Compact Format
```yaml
test: Login Test
nav: https://example.com
do:
  - c: "button.login"
  - t: {input#email: "test@example.com"}
  - k: "Enter"
  - ch: "Welcome"
```

### Simplified Format
```yaml
name: Login Test
starting_url: https://example.com
steps:
  - click: "button.login"
  - write:
      selector: "input#email"
      text: "test@example.com"
  - key: "Enter"
  - assert: "Welcome"
```

### Extended Format
```yaml
name: Login Test
infrastructure:
  starting_url: https://example.com
steps:
  - type: click
    target: "button.login"
  - type: write
    selector: "input#email"
    text: "test@example.com"
  - type: key
    value: "Enter"
  - type: assert
    command: exists
    target: "Welcome"
```

## Conversion Matrix

| From Format | To Compact | To Simplified | To Extended |
|-------------|------------|---------------|-------------|
| Compact     | -          | ✓             | ✓           |
| Simplified  | ✓          | -             | ✓           |
| Extended    | ✓          | ✓             | -           |

## Feature Support

### Compact Format
- Concise syntax (c:, t:, ch:, etc.)
- AI-optimized structure
- Full CLI support (`yaml` command)
- Control flow (if/loop)
- Setup/teardown sections
- Data variables

### Simplified Format  
- Readable syntax
- Full CLI support (`run-test` command)
- Infrastructure configuration
- Test variables
- Config options

### Extended Format
- Verbose syntax
- No CLI support
- Full metadata
- Infrastructure configuration
- Detailed step information

## Conversion Warnings

The converter provides warnings when:
- Features are not supported in the target format
- Mixed format indicators are detected
- Low confidence format detection
- Control flow cannot be represented
- Infrastructure config is lost

## Implementation Details

### Unified Intermediate Format

All conversions go through a unified intermediate representation:

```go
type UnifiedTest struct {
    Name           string
    Description    string
    BaseURL        string
    StartURL       string
    Config         map[string]interface{}
    Setup          []UnifiedStep
    Steps          []UnifiedStep
    Teardown       []UnifiedStep
    Data           map[string]interface{}
    Variables      map[string]interface{}
    Infrastructure map[string]interface{}
}

type UnifiedStep struct {
    Type        string
    Target      string
    Value       string
    Variable    string
    Selector    string
    Text        string
    Attribute   string
    Index       int
    Options     map[string]interface{}
    Condition   string
    Actions     []UnifiedStep
    ElseActions []UnifiedStep
    Original    map[string]interface{}
}
```

### Conversion Process

1. **Detection** - Automatically detect source format
2. **Parsing** - Parse to unified intermediate format
3. **Transformation** - Convert unified format to target
4. **Validation** - Verify output matches target format

### Lossless Conversion

The converter preserves original data where possible:
- Original step data is stored in the `Original` field
- Complex structures are maintained through `Options`
- Format-specific features generate warnings when lost

## Testing

Run the comprehensive test suite:

```bash
go test ./pkg/api-cli/yaml-layer/converter/...
```

Tests include:
- Format detection accuracy
- Bidirectional conversions
- Round-trip conversions
- Edge case handling
- Invalid format handling

## Examples

### Convert Compact to Simplified
```go
compact := `test: My Test
do:
  - c: "button"
  - t: "text"`

result, _ := converter.Convert([]byte(compact), detector.FormatSimplified)
// Output:
// name: My Test
// steps:
//   - click: "button"
//   - write: "text"
```

### Detect Format with Confidence
```go
detection, _ := detector.DetectFormat(yamlContent)
fmt.Printf("Format: %s (%.0f%% confidence)\n", 
    detection.Format, detection.Confidence * 100)
```

### Batch Conversion
```go
cli := NewCLI()
files := []string{"test1.yaml", "test2.yaml", "test3.yaml"}
err := cli.ConvertBatch(files, "output/", detector.FormatCompact)
```

## Error Handling

The converter returns errors for:
- Invalid YAML syntax
- Unknown source format
- Unsupported target format
- File I/O errors
- Validation failures

Warnings are provided for:
- Feature incompatibilities
- Mixed format indicators
- Low confidence detection
- Data loss during conversion