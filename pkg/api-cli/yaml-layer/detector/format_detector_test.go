package detector

import (
	"testing"
)

func TestDetectFormat(t *testing.T) {
	detector := NewFormatDetector()
	
	tests := []struct {
		name           string
		yaml           string
		expectedFormat YAMLFormat
		minConfidence  float64
	}{
		{
			name: "Compact format",
			yaml: `test: Login Test
nav: https://example.com
do:
  - c: "button.login"
  - t: "#email"
  - ch: "Welcome"`,
			expectedFormat: FormatCompact,
			minConfidence:  0.8,
		},
		{
			name: "Simplified format",
			yaml: `name: Login Test
steps:
  - navigate: "https://example.com"
  - click: "button.login"
  - write:
      selector: "#email"
      text: "test@example.com"
  - assert: "Welcome"`,
			expectedFormat: FormatSimplified,
			minConfidence:  0.8,
		},
		{
			name: "Extended format",
			yaml: `name: Login Test
infrastructure:
  organization_id: "2242"
  project:
    name: "Test Project"
steps:
  - type: navigate
    target: "https://example.com"
  - type: click
    target: "button.login"`,
			expectedFormat: FormatExtended,
			minConfidence:  0.7,
		},
		{
			name: "Empty YAML",
			yaml: ``,
			expectedFormat: FormatUnknown,
			minConfidence:  0,
		},
		{
			name: "Mixed format indicators",
			yaml: `test: Test Name
name: Also Test Name
do:
  - c: "button"
steps:
  - click: "button"`,
			expectedFormat: FormatUnknown, // Should have low confidence
			minConfidence:  0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := detector.DetectFormat([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("DetectFormat failed: %v", err)
			}
			
			if result.Format != tt.expectedFormat {
				t.Errorf("Expected format %s, got %s", tt.expectedFormat, result.Format)
			}
			
			if result.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %.2f, got %.2f", tt.minConfidence, result.Confidence)
			}
			
			// Check for warnings on mixed format
			if tt.name == "Mixed format indicators" && len(result.Warnings) == 0 {
				t.Error("Expected warnings for mixed format indicators")
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	detector := NewFormatDetector()
	
	compactYAML := []byte(`test: Test
do:
  - c: "button"`)
	
	// Should pass validation for compact format
	err := detector.ValidateFormat(compactYAML, FormatCompact)
	if err != nil {
		t.Errorf("ValidateFormat failed for correct format: %v", err)
	}
	
	// Should fail validation for wrong format
	err = detector.ValidateFormat(compactYAML, FormatSimplified)
	if err == nil {
		t.Error("ValidateFormat should have failed for wrong format")
	}
}

func TestFormatHelpers(t *testing.T) {
	// Test format descriptions
	if desc := GetFormatDescription(FormatCompact); desc == "" {
		t.Error("GetFormatDescription returned empty for compact format")
	}
	
	// Test format examples
	if example := GetFormatExample(FormatCompact); example == "" {
		t.Error("GetFormatExample returned empty for compact format")
	}
	
	// Test format support
	if !IsFormatSupported(FormatCompact) {
		t.Error("IsFormatSupported returned false for compact format")
	}
	
	if IsFormatSupported(FormatExtended) {
		t.Error("IsFormatSupported returned true for extended format")
	}
	
	// Test supported commands
	if cmd := GetSupportedCommand(FormatCompact); cmd != "yaml" {
		t.Errorf("GetSupportedCommand returned %s for compact format, expected yaml", cmd)
	}
}