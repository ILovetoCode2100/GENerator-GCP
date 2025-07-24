package detector

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// YAMLFormat represents the detected YAML format type
type YAMLFormat string

const (
	// FormatCompact is the AI-optimized compact format used by yaml commands
	FormatCompact YAMLFormat = "compact"
	
	// FormatSimplified is the readable format used by run-test command
	FormatSimplified YAMLFormat = "simplified"
	
	// FormatExtended is the verbose format found in examples (no CLI support)
	FormatExtended YAMLFormat = "extended"
	
	// FormatUnknown indicates format could not be determined
	FormatUnknown YAMLFormat = "unknown"
)

// DetectionResult contains the format detection results
type DetectionResult struct {
	Format     YAMLFormat            `json:"format"`
	Confidence float64               `json:"confidence"`
	Features   map[string]bool       `json:"features"`
	Warnings   []string              `json:"warnings,omitempty"`
}

// FormatDetector handles YAML format detection
type FormatDetector struct {
	// Minimum confidence threshold for detection
	minConfidence float64
}

// NewFormatDetector creates a new format detector
func NewFormatDetector() *FormatDetector {
	return &FormatDetector{
		minConfidence: 0.7,
	}
}

// DetectFormat analyzes YAML content and determines its format
func (d *FormatDetector) DetectFormat(content []byte) (*DetectionResult, error) {
	// Parse YAML into generic map
	var data map[string]interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}
	
	// Initialize result
	result := &DetectionResult{
		Format:   FormatUnknown,
		Features: make(map[string]bool),
		Warnings: []string{},
	}
	
	// Check for empty document
	if len(data) == 0 {
		result.Warnings = append(result.Warnings, "Empty YAML document")
		return result, nil
	}
	
	// Detect features
	d.detectFeatures(data, result)
	
	// Calculate format scores
	compactScore := d.calculateCompactScore(result.Features)
	simplifiedScore := d.calculateSimplifiedScore(result.Features)
	extendedScore := d.calculateExtendedScore(result.Features)
	
	// Determine format based on highest score
	maxScore := 0.0
	if compactScore > maxScore {
		maxScore = compactScore
		result.Format = FormatCompact
		result.Confidence = compactScore
	}
	if simplifiedScore > maxScore {
		maxScore = simplifiedScore
		result.Format = FormatSimplified
		result.Confidence = simplifiedScore
	}
	if extendedScore > maxScore {
		maxScore = extendedScore
		result.Format = FormatExtended
		result.Confidence = extendedScore
	}
	
	// Check for mixed format indicators first
	d.checkMixedFormats(result)
	
	// If mixed formats detected, set to unknown
	if len(result.Warnings) > 0 && 
		(result.Features["has_test_field"] && result.Features["has_name_field"]) ||
		(result.Features["has_do_field"] && result.Features["has_steps_field"]) {
		result.Format = FormatUnknown
		result.Confidence = 0.0
	}
	
	// Check for ambiguous detection
	if result.Confidence < d.minConfidence {
		result.Warnings = append(result.Warnings, 
			fmt.Sprintf("Low confidence detection (%.2f). Format may be ambiguous.", result.Confidence))
	}
	
	return result, nil
}

// detectFeatures analyzes the YAML structure for format-specific features
func (d *FormatDetector) detectFeatures(data map[string]interface{}, result *DetectionResult) {
	// Check top-level fields
	for key := range data {
		result.Features["field_"+key] = true
	}
	
	// Compact format features
	if _, ok := data["test"]; ok {
		result.Features["has_test_field"] = true
	}
	if _, ok := data["do"]; ok {
		result.Features["has_do_field"] = true
		// Check action format
		if actions, ok := data["do"].([]interface{}); ok {
			d.analyzeCompactActions(actions, result)
		}
	}
	if _, ok := data["nav"]; ok {
		result.Features["has_nav_field"] = true
	}
	if _, ok := data["data"]; ok {
		result.Features["has_data_field"] = true
	}
	
	// Simplified format features
	if _, ok := data["name"]; ok {
		result.Features["has_name_field"] = true
	}
	if _, ok := data["steps"]; ok {
		result.Features["has_steps_field"] = true
		// Check step format
		if steps, ok := data["steps"].([]interface{}); ok {
			d.analyzeSimplifiedSteps(steps, result)
		}
	}
	
	// Extended format features
	if _, ok := data["infrastructure"]; ok {
		result.Features["has_infrastructure_field"] = true
	}
	if config, ok := data["config"]; ok {
		result.Features["has_config_field"] = true
		// Check if it's extended config
		if configMap, ok := config.(map[string]interface{}); ok {
			if _, hasTimeout := configMap["timeout"]; hasTimeout {
				result.Features["has_extended_config"] = true
			}
		}
	}
}

// analyzeCompactActions checks if actions follow compact format
func (d *FormatDetector) analyzeCompactActions(actions []interface{}, result *DetectionResult) {
	compactCount := 0
	totalCount := 0
	
	for _, action := range actions {
		totalCount++
		
		switch v := action.(type) {
		case string:
			// Simple string actions like "wait: 1000"
			result.Features["has_string_actions"] = true
			compactCount++
			
		case map[interface{}]interface{}, map[string]interface{}:
			// Check for compact syntax (c:, t:, ch:, etc.)
			actionMap := make(map[string]interface{})
			switch vm := v.(type) {
			case map[interface{}]interface{}:
				for k, v := range vm {
					if ks, ok := k.(string); ok {
						actionMap[ks] = v
					}
				}
			case map[string]interface{}:
				actionMap = vm
			}
			
			// Check for compact keys
			for key := range actionMap {
				if len(key) <= 4 { // c, t, ch, wait, note, etc.
					compactCount++
					result.Features["has_compact_actions"] = true
					break
				}
			}
		}
	}
	
	if totalCount > 0 && float64(compactCount)/float64(totalCount) > 0.8 {
		result.Features["majority_compact_actions"] = true
	}
}

// analyzeSimplifiedSteps checks if steps follow simplified format
func (d *FormatDetector) analyzeSimplifiedSteps(steps []interface{}, result *DetectionResult) {
	simplifiedCount := 0
	totalCount := 0
	
	for _, step := range steps {
		totalCount++
		
		if stepMap, ok := step.(map[interface{}]interface{}); ok {
			// Check for simplified syntax (navigate:, click:, assert:, etc.)
			for key := range stepMap {
				if keyStr, ok := key.(string); ok {
					switch keyStr {
					case "navigate", "click", "write", "assert", "wait":
						simplifiedCount++
						result.Features["has_simplified_steps"] = true
					case "type":
						// Extended format indicator
						result.Features["has_type_field_in_steps"] = true
					}
				}
			}
		} else if stepMap, ok := step.(map[string]interface{}); ok {
			// String key version
			for key := range stepMap {
				switch key {
				case "navigate", "click", "write", "assert", "wait":
					simplifiedCount++
					result.Features["has_simplified_steps"] = true
				case "type":
					result.Features["has_type_field_in_steps"] = true
				}
			}
		}
	}
	
	if totalCount > 0 && float64(simplifiedCount)/float64(totalCount) > 0.8 {
		result.Features["majority_simplified_steps"] = true
	}
}

// calculateCompactScore calculates confidence score for compact format
func (d *FormatDetector) calculateCompactScore(features map[string]bool) float64 {
	score := 0.0
	weights := map[string]float64{
		"has_test_field":          0.3,
		"has_do_field":            0.3,
		"has_compact_actions":     0.2,
		"majority_compact_actions": 0.1,
		"has_nav_field":           0.05,
		"has_data_field":          0.05,
	}
	
	// Negative indicators
	if features["has_name_field"] {
		score -= 0.2
	}
	if features["has_steps_field"] {
		score -= 0.2
	}
	if features["has_infrastructure_field"] {
		score -= 0.3
	}
	
	// Calculate positive score
	for feature, weight := range weights {
		if features[feature] {
			score += weight
		}
	}
	
	// Ensure score is between 0 and 1
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}
	
	return score
}

// calculateSimplifiedScore calculates confidence score for simplified format
func (d *FormatDetector) calculateSimplifiedScore(features map[string]bool) float64 {
	score := 0.0
	weights := map[string]float64{
		"has_name_field":             0.3,
		"has_steps_field":            0.3,
		"has_simplified_steps":       0.2,
		"majority_simplified_steps":  0.2,
	}
	
	// Negative indicators
	if features["has_test_field"] {
		score -= 0.3
	}
	if features["has_do_field"] {
		score -= 0.3
	}
	if features["has_type_field_in_steps"] {
		score -= 0.2
	}
	if features["has_infrastructure_field"] {
		score -= 0.1
	}
	
	// Calculate positive score
	for feature, weight := range weights {
		if features[feature] {
			score += weight
		}
	}
	
	// Ensure score is between 0 and 1
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}
	
	return score
}

// calculateExtendedScore calculates confidence score for extended format
func (d *FormatDetector) calculateExtendedScore(features map[string]bool) float64 {
	score := 0.0
	weights := map[string]float64{
		"has_name_field":           0.2,
		"has_steps_field":          0.2,
		"has_infrastructure_field": 0.3,
		"has_type_field_in_steps":  0.2,
		"has_extended_config":      0.1,
	}
	
	// Negative indicators
	if features["has_test_field"] {
		score -= 0.3
	}
	if features["has_do_field"] {
		score -= 0.3
	}
	if features["has_compact_actions"] {
		score -= 0.2
	}
	if features["has_simplified_steps"] {
		score -= 0.1
	}
	
	// Calculate positive score
	for feature, weight := range weights {
		if features[feature] {
			score += weight
		}
	}
	
	// Ensure score is between 0 and 1
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}
	
	return score
}

// checkMixedFormats warns about mixed format indicators
func (d *FormatDetector) checkMixedFormats(result *DetectionResult) {
	// Check for conflicting indicators
	if result.Features["has_test_field"] && result.Features["has_name_field"] {
		result.Warnings = append(result.Warnings, 
			"Mixed format indicators: both 'test' and 'name' fields present")
	}
	
	if result.Features["has_do_field"] && result.Features["has_steps_field"] {
		result.Warnings = append(result.Warnings, 
			"Mixed format indicators: both 'do' and 'steps' fields present")
	}
	
	if result.Features["has_compact_actions"] && result.Features["has_type_field_in_steps"] {
		result.Warnings = append(result.Warnings, 
			"Mixed action styles: both compact and extended syntax detected")
	}
}

// ValidateFormat validates that content matches the expected format
func (d *FormatDetector) ValidateFormat(content []byte, expectedFormat YAMLFormat) error {
	result, err := d.DetectFormat(content)
	if err != nil {
		return err
	}
	
	if result.Format != expectedFormat {
		return fmt.Errorf("format mismatch: expected %s but detected %s (confidence: %.2f)",
			expectedFormat, result.Format, result.Confidence)
	}
	
	if result.Confidence < d.minConfidence {
		return fmt.Errorf("low confidence format match: %.2f (minimum: %.2f)",
			result.Confidence, d.minConfidence)
	}
	
	return nil
}

// GetFormatDescription returns a human-readable description of the format
func GetFormatDescription(format YAMLFormat) string {
	switch format {
	case FormatCompact:
		return "Compact format (AI-optimized, used by yaml commands)"
	case FormatSimplified:
		return "Simplified format (readable, used by run-test command)"
	case FormatExtended:
		return "Extended format (verbose, no CLI support)"
	case FormatUnknown:
		return "Unknown format"
	default:
		return string(format)
	}
}

// GetFormatExample returns an example of the specified format
func GetFormatExample(format YAMLFormat) string {
	switch format {
	case FormatCompact:
		return `test: Example Test
nav: https://example.com
do:
  - c: "button.start"
  - t: {input#email: "test@example.com"}
  - ch: "Success"`
  
	case FormatSimplified:
		return `name: Example Test
steps:
  - navigate: "https://example.com"
  - click: "button.start"
  - write:
      selector: "input#email"
      text: "test@example.com"
  - assert: "Success"`
  
	case FormatExtended:
		return `name: Example Test
infrastructure:
  organization_id: "2242"
  project:
    name: "Test Project"
steps:
  - type: navigate
    target: "https://example.com"
  - type: click
    target: "button.start"
  - type: assert
    command: exists
    target: "Success"`
    
	default:
		return ""
	}
}

// IsFormatSupported returns whether the format has CLI support
func IsFormatSupported(format YAMLFormat) bool {
	switch format {
	case FormatCompact, FormatSimplified:
		return true
	case FormatExtended:
		return false
	default:
		return false
	}
}

// GetSupportedCommand returns the CLI command that supports the format
func GetSupportedCommand(format YAMLFormat) string {
	switch format {
	case FormatCompact:
		return "yaml"
	case FormatSimplified:
		return "run-test"
	case FormatExtended:
		return ""
	default:
		return ""
	}
}