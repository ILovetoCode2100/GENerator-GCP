package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLNormalizer handles YAML type conversions and format normalization
type YAMLNormalizer struct {
	strict bool
	debug  bool
}

// NewYAMLNormalizer creates a new YAML normalizer
func NewYAMLNormalizer(strict, debug bool) *YAMLNormalizer {
	return &YAMLNormalizer{
		strict: strict,
		debug:  debug,
	}
}

// NormalizeYAML converts YAML types to JSON-compatible types
func (n *YAMLNormalizer) NormalizeYAML(data interface{}) interface{} {
	switch v := data.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for key, val := range v {
			keyStr := fmt.Sprintf("%v", key)
			m[keyStr] = n.NormalizeYAML(val)
		}
		return m
	case []interface{}:
		normalized := make([]interface{}, len(v))
		for i, val := range v {
			normalized[i] = n.NormalizeYAML(val)
		}
		return normalized
	case map[string]interface{}:
		for key, val := range v {
			v[key] = n.NormalizeYAML(val)
		}
		return v
	default:
		return v
	}
}

// YAMLParser provides robust YAML parsing with format detection
type YAMLParser struct {
	normalizer *YAMLNormalizer
}

// NewYAMLParser creates a new YAML parser
func NewYAMLParser(strict bool) *YAMLParser {
	return &YAMLParser{
		normalizer: NewYAMLNormalizer(strict, false),
	}
}

// Parse handles various YAML formats
func (p *YAMLParser) Parse(content []byte) (map[string]interface{}, error) {
	var raw interface{}
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Normalize the parsed data
	normalized := p.normalizer.NormalizeYAML(raw)

	// Ensure it's a map
	result, ok := normalized.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("YAML root must be an object, got %T", normalized)
	}

	return result, nil
}

// ValidateAndConvert ensures YAML is in the expected format
func (p *YAMLParser) ValidateAndConvert(content []byte, target interface{}) error {
	data, err := p.Parse(content)
	if err != nil {
		return err
	}

	// Convert to JSON then back to target type
	// This ensures compatibility
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target type: %w", err)
	}

	return nil
}

// YAMLFormat represents different YAML test formats
type YAMLFormat int

const (
	FormatUnknown YAMLFormat = iota
	FormatCompact
	FormatSimplified
	FormatExtended
	FormatHybrid
)

// DetectYAMLFormat identifies which format a YAML file uses
func DetectYAMLFormat(data map[string]interface{}) YAMLFormat {
	// Check for format indicators
	if steps, hasSteps := data["steps"]; hasSteps {
		stepsList, ok := steps.([]interface{})
		if !ok || len(stepsList) == 0 {
			return FormatUnknown
		}

		// Analyze first few steps to determine format
		compactCount := 0
		extendedCount := 0
		simplifiedCount := 0

		for i, step := range stepsList {
			if i >= 3 { // Check first 3 steps
				break
			}

			switch v := step.(type) {
			case string:
				compactCount++
			case map[string]interface{}:
				if _, hasAction := v["action"]; hasAction {
					if _, hasExtension := v["extension"]; hasExtension {
						extendedCount++
					} else {
						simplifiedCount++
					}
				} else {
					// Check for compact map format like {click: "button"}
					if len(v) == 1 {
						compactCount++
					}
				}
			}
		}

		// Determine format based on counts
		if compactCount > 0 && extendedCount == 0 && simplifiedCount == 0 {
			return FormatCompact
		}
		if extendedCount > 0 && compactCount == 0 {
			return FormatExtended
		}
		if simplifiedCount > 0 && compactCount == 0 {
			return FormatSimplified
		}
		if compactCount > 0 && (extendedCount > 0 || simplifiedCount > 0) {
			return FormatHybrid
		}
	}

	return FormatUnknown
}

// ConvertToStandardFormat converts any YAML format to standard format
func ConvertToStandardFormat(data map[string]interface{}, format YAMLFormat) (map[string]interface{}, error) {
	switch format {
	case FormatCompact:
		return convertCompactFormat(data)
	case FormatSimplified:
		return convertSimplifiedFormat(data)
	case FormatExtended:
		return data, nil // Already standard
	case FormatHybrid:
		return convertHybridFormat(data)
	default:
		return nil, fmt.Errorf("unknown YAML format")
	}
}

func convertCompactFormat(data map[string]interface{}) (map[string]interface{}, error) {
	steps, ok := data["steps"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid compact format: missing steps")
	}

	var convertedSteps []interface{}
	for _, step := range steps {
		switch v := step.(type) {
		case string:
			// Parse compact step string like "click: Login button"
			parsed := parseCompactStep(v)
			convertedSteps = append(convertedSteps, parsed)
		case map[string]interface{}:
			// Handle map format like {click: "button"}
			if len(v) == 1 {
				for action, target := range v {
					convertedSteps = append(convertedSteps, map[string]interface{}{
						"action": action,
						"target": target,
					})
				}
			}
		}
	}

	data["steps"] = convertedSteps
	return data, nil
}

func parseCompactStep(step string) map[string]interface{} {
	// Parse compact format like "click: Login button"
	parts := strings.SplitN(step, ":", 2)
	if len(parts) != 2 {
		return map[string]interface{}{
			"action": "unknown",
			"target": step,
		}
	}

	action := strings.TrimSpace(parts[0])
	target := strings.TrimSpace(parts[1])

	// Handle special cases
	switch action {
	case "navigate":
		return map[string]interface{}{
			"action": "navigate",
			"url":    target,
		}
	case "write", "type":
		// Parse "write: 'input#email' 'test@example.com'"
		targetParts := parseQuotedStrings(target)
		if len(targetParts) >= 2 {
			return map[string]interface{}{
				"action": "write",
				"target": targetParts[0],
				"text":   targetParts[1],
			}
		}
	case "wait":
		if strings.HasPrefix(target, "for ") {
			return map[string]interface{}{
				"action":  "wait",
				"element": strings.TrimPrefix(target, "for "),
			}
		}
		if strings.HasSuffix(target, "ms") || strings.HasSuffix(target, "s") {
			return map[string]interface{}{
				"action": "wait",
				"time":   target,
			}
		}
	}

	return map[string]interface{}{
		"action": action,
		"target": target,
	}
}

func convertSimplifiedFormat(data map[string]interface{}) (map[string]interface{}, error) {
	// Simplified format is close to standard, just ensure consistency
	steps, ok := data["steps"].([]interface{})
	if !ok {
		return data, nil
	}

	for i, step := range steps {
		if stepMap, ok := step.(map[string]interface{}); ok {
			// Ensure all fields are properly typed
			steps[i] = normalizeStepMap(stepMap)
		}
	}

	return data, nil
}

func convertHybridFormat(data map[string]interface{}) (map[string]interface{}, error) {
	// Handle mixed format files
	steps, ok := data["steps"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid hybrid format: missing steps")
	}

	var convertedSteps []interface{}
	for _, step := range steps {
		switch v := step.(type) {
		case string:
			parsed := parseCompactStep(v)
			convertedSteps = append(convertedSteps, parsed)
		case map[string]interface{}:
			convertedSteps = append(convertedSteps, normalizeStepMap(v))
		}
	}

	data["steps"] = convertedSteps
	return data, nil
}

func normalizeStepMap(step map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})

	// Copy all fields
	for k, v := range step {
		normalized[k] = v
	}

	// Ensure action is present
	if _, hasAction := normalized["action"]; !hasAction {
		// Try to infer action from other fields
		if _, hasClick := normalized["click"]; hasClick {
			normalized["action"] = "click"
			normalized["target"] = normalized["click"]
			delete(normalized, "click")
		}
	}

	return normalized
}

func parseQuotedStrings(input string) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range input {
		switch {
		case !inQuote && (r == '"' || r == '\''):
			inQuote = true
			quoteChar = r
		case inQuote && r == quoteChar:
			inQuote = false
			result = append(result, current.String())
			current.Reset()
		case inQuote:
			current.WriteRune(r)
		case !inQuote && r == ' ':
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
