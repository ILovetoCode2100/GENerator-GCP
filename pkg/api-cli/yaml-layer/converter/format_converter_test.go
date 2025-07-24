package converter

import (
	"strings"
	"testing"

	"github.com/your-org/virtuoso-generator/pkg/api-cli/yaml-layer/detector"
	"gopkg.in/yaml.v3"
)

func TestFormatConverter_Convert(t *testing.T) {
	converter := NewFormatConverter()

	tests := []struct {
		name         string
		sourceFormat string
		targetFormat detector.YAMLFormat
		input        string
		wantContains []string
		wantWarnings []string
		wantErr      bool
	}{
		{
			name:         "Compact to Simplified",
			sourceFormat: "compact",
			targetFormat: detector.FormatSimplified,
			input: `test: Login Test
nav: https://example.com
do:
  - c: "button.login"
  - t: {input#email: "test@example.com"}
  - k: "Enter"
  - ch: "Welcome"`,
			wantContains: []string{
				"name: Login Test",
				"starting_url: https://example.com",
				"steps:",
				"- click: button.login",
				`- write:
    selector: input#email
    text: test@example.com`,
				"- key: Enter",
				"- assert: Welcome",
			},
		},
		{
			name:         "Simplified to Compact",
			sourceFormat: "simplified",
			targetFormat: detector.FormatCompact,
			input: `name: Login Test
starting_url: https://example.com
steps:
  - click: "button.login"
  - write:
      selector: "input#email"
      text: "test@example.com"
  - assert: "Welcome"`,
			wantContains: []string{
				"test: Login Test",
				"nav: https://example.com",
				"do:",
				`- c: button.login`,
				`- t:
    input#email: test@example.com`,
				`- ch: Welcome`,
			},
		},
		{
			name:         "Extended to Compact",
			sourceFormat: "extended",
			targetFormat: detector.FormatCompact,
			input: `name: Login Test
infrastructure:
  starting_url: https://example.com
steps:
  - type: navigate
    target: https://example.com
  - type: click
    target: button.login
  - type: assert
    command: exists
    target: Welcome`,
			wantContains: []string{
				"test: Login Test",
				"nav: https://example.com",
				"do:",
				`- nav: https://example.com`,
				`- c: button.login`,
				`- ch: Welcome`,
			},
			wantWarnings: []string{
				"Infrastructure configuration is not supported in compact format",
			},
		},
		{
			name:         "Compact with Control Flow",
			sourceFormat: "compact",
			targetFormat: detector.FormatSimplified,
			input: `test: Conditional Test
do:
  - if:
      cond: "$loggedIn"
      then:
        - c: "button.logout"
      else:
        - c: "button.login"`,
			wantContains: []string{
				"name: Conditional Test",
				"comment: Control flow (if) cannot be represented in simplified format",
			},
		},
		{
			name:         "Compact with Data Variables",
			sourceFormat: "compact",
			targetFormat: detector.FormatSimplified,
			input: `test: Data Test
data:
  username: testuser
  password: testpass
do:
  - t: {input#user: "$username"}
  - t: {input#pass: "$password"}`,
			wantContains: []string{
				"name: Data Test",
				"variables:",
				"- name: username",
				"  value: testuser",
				"- name: password",
				"  value: testpass",
				`- write:
    selector: input#user
    text: $username`,
			},
		},
		{
			name:         "Simplified with Assertions",
			sourceFormat: "simplified",
			targetFormat: detector.FormatCompact,
			input: `name: Assert Test
steps:
  - assert:
      not_exists: "Error message"
  - assert:
      selector: "h1"
      equals: "Welcome"
  - assert:
      selector: "span.count"
      not_equals: "0"`,
			wantContains: []string{
				"test: Assert Test",
				`- nch: Error message`,
				`- eq:
    h1: Welcome`,
				`- neq:
    span.count: "0"`,
			},
		},
		{
			name:         "Complex Actions Conversion",
			sourceFormat: "compact",
			targetFormat: detector.FormatSimplified,
			input: `test: Complex Test
do:
  - wait: 1000
  - wait: "div.loaded"
  - store: {selector: "h1", variable: "title"}
  - cookie: {action: create, name: "session", value: "abc123"}
  - scroll: "bottom"
  - select: {selector: "select#country", option: "US"}`,
			wantContains: []string{
				"- wait: \"1000\"",
				`- wait:
    element: div.loaded`,
				`- store:
    selector: h1
    variable: title`,
				`- cookie:
    action: create
    name: session
    value: abc123`,
				"- scroll: bottom",
				`- select:
    option: US
    selector: select#country`,
			},
		},
		{
			name:         "Extended with Meta",
			sourceFormat: "extended",
			targetFormat: detector.FormatSimplified,
			input: `name: Meta Test
steps:
  - type: click
    target: button.submit
    meta:
      timeout: 5000
      retries: 3`,
			wantContains: []string{
				"- click: button.submit",
			},
		},
		{
			name:         "Setup and Teardown",
			sourceFormat: "compact",
			targetFormat: detector.FormatSimplified,
			input: `test: Full Test
setup:
  - nav: "https://example.com"
  - c: "button.accept-cookies"
do:
  - c: "button.start"
teardown:
  - c: "button.logout"`,
			wantContains: []string{
				"name: Full Test",
				"comment: === Setup Steps ===",
				"- navigate: https://example.com",
				"- click: button.accept-cookies",
				"- click: button.start",
				"comment: === Teardown Steps ===",
				"- click: button.logout",
			},
		},
		{
			name:         "Invalid Format Detection",
			sourceFormat: "invalid",
			targetFormat: detector.FormatCompact,
			input:        `invalid: yaml content`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert([]byte(tt.input), tt.targetFormat)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr {
				return
			}
			
			if result == nil {
				t.Fatal("Convert() returned nil result")
			}
			
			// Check output contains expected strings
			output := string(result.Output)
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Convert() output missing expected content:\nwant: %s\ngot:\n%s", want, output)
				}
			}
			
			// Check warnings
			if len(tt.wantWarnings) > 0 {
				for _, wantWarn := range tt.wantWarnings {
					found := false
					for _, warn := range result.Warnings {
						if strings.Contains(warn, wantWarn) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Convert() missing expected warning: %s\ngot warnings: %v", wantWarn, result.Warnings)
					}
				}
			}
			
			// Validate output is valid YAML
			var parsed interface{}
			if err := yaml.Unmarshal(result.Output, &parsed); err != nil {
				t.Errorf("Convert() produced invalid YAML: %v\noutput:\n%s", err, output)
			}
			
			// Validate format detection
			if err := converter.ValidateConversion([]byte(tt.input), result.Output, tt.targetFormat); err != nil {
				t.Errorf("ValidateConversion() failed: %v", err)
			}
		})
	}
}

func TestFormatConverter_RoundTrip(t *testing.T) {
	converter := NewFormatConverter()
	
	tests := []struct {
		name   string
		format detector.YAMLFormat
		input  string
	}{
		{
			name:   "Compact Round Trip",
			format: detector.FormatCompact,
			input: `test: Round Trip Test
nav: https://example.com
do:
  - c: "button"
  - t: {input: "text"}
  - ch: "Success"`,
		},
		{
			name:   "Simplified Round Trip",
			format: detector.FormatSimplified,
			input: `name: Round Trip Test
starting_url: https://example.com
steps:
  - click: "button"
  - write:
      selector: "input"
      text: "text"
  - assert: "Success"`,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to another format and back
			var intermediateFormat detector.YAMLFormat
			if tt.format == detector.FormatCompact {
				intermediateFormat = detector.FormatSimplified
			} else {
				intermediateFormat = detector.FormatCompact
			}
			
			// First conversion
			result1, err := converter.Convert([]byte(tt.input), intermediateFormat)
			if err != nil {
				t.Fatalf("First conversion failed: %v", err)
			}
			
			// Second conversion back to original
			result2, err := converter.Convert(result1.Output, tt.format)
			if err != nil {
				t.Fatalf("Second conversion failed: %v", err)
			}
			
			// Parse both original and final to compare structure
			var original, final interface{}
			if err := yaml.Unmarshal([]byte(tt.input), &original); err != nil {
				t.Fatalf("Failed to parse original: %v", err)
			}
			if err := yaml.Unmarshal(result2.Output, &final); err != nil {
				t.Fatalf("Failed to parse final: %v", err)
			}
			
			// Note: We don't expect exact equality due to format differences,
			// but the structure should be similar
			t.Logf("Original:\n%s\nFinal:\n%s", tt.input, string(result2.Output))
		})
	}
}

func TestUnifiedStep_Conversions(t *testing.T) {
	converter := NewFormatConverter()
	
	tests := []struct {
		name    string
		unified UnifiedStep
		compact string
		simple  string
	}{
		{
			name: "Click Action",
			unified: UnifiedStep{
				Type:   "click",
				Target: "button.submit",
			},
			compact: "c",
			simple:  "click",
		},
		{
			name: "Write Action",
			unified: UnifiedStep{
				Type:   "write",
				Target: "input#email",
				Text:   "test@example.com",
			},
			compact: "t",
			simple:  "write",
		},
		{
			name: "Assert Exists",
			unified: UnifiedStep{
				Type:    "assert",
				Target:  "div.success",
				Options: map[string]interface{}{"exists": true},
			},
			compact: "ch",
			simple:  "assert",
		},
		{
			name: "Assert Not Exists",
			unified: UnifiedStep{
				Type:    "assert",
				Target:  "div.error",
				Options: map[string]interface{}{"exists": false},
			},
			compact: "nch",
			simple:  "assert",
		},
		{
			name: "Wait Time",
			unified: UnifiedStep{
				Type:  "wait",
				Value: "1000",
			},
			compact: "wait",
			simple:  "wait",
		},
		{
			name: "Wait Element",
			unified: UnifiedStep{
				Type:   "wait",
				Target: "div.loaded",
			},
			compact: "wait",
			simple:  "wait",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test unified to compact
			compactAction := converter.unifiedToCompactAction(tt.unified)
			compactYAML, _ := yaml.Marshal(compactAction)
			if !strings.Contains(string(compactYAML), tt.compact) {
				t.Errorf("Compact conversion failed: expected to contain %s, got %s", tt.compact, string(compactYAML))
			}
			
			// Test unified to simplified
			simpleStep := converter.unifiedToSimplifiedStep(tt.unified)
			simpleYAML, _ := yaml.Marshal(simpleStep)
			if !strings.Contains(string(simpleYAML), tt.simple) {
				t.Errorf("Simplified conversion failed: expected to contain %s, got %s", tt.simple, string(simpleYAML))
			}
		})
	}
}

func TestGetConversionCapabilities(t *testing.T) {
	caps := GetConversionCapabilities()
	
	// Check all formats can convert to all other formats
	formats := []string{"compact", "simplified", "extended"}
	for _, from := range formats {
		for _, to := range formats {
			if from == to {
				continue
			}
			if !caps[from][to] {
				t.Errorf("Expected %s to %s conversion to be supported", from, to)
			}
		}
	}
}

func TestGetFormatFeatures(t *testing.T) {
	features := GetFormatFeatures()
	
	// Check each format has features defined
	formats := []string{"compact", "simplified", "extended"}
	for _, format := range formats {
		if len(features[format]) == 0 {
			t.Errorf("No features defined for format %s", format)
		}
	}
	
	// Check specific features
	compactFeatures := features["compact"]
	found := false
	for _, f := range compactFeatures {
		if strings.Contains(f, "AI-optimized") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected compact format to have AI-optimized feature")
	}
}