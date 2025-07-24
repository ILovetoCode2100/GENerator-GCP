package converter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/yaml-layer/core"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/yaml-layer/detector"
	"gopkg.in/yaml.v3"
)

// FormatConverter handles conversion between different YAML formats
type FormatConverter struct {
	detector *detector.FormatDetector
}

// NewFormatConverter creates a new format converter
func NewFormatConverter() *FormatConverter {
	return &FormatConverter{
		detector: detector.NewFormatDetector(),
	}
}

// ConversionResult contains the result of a format conversion
type ConversionResult struct {
	Success       bool                   `json:"success"`
	SourceFormat  detector.YAMLFormat    `json:"source_format"`
	TargetFormat  detector.YAMLFormat    `json:"target_format"`
	Output        []byte                 `json:"output,omitempty"`
	Warnings      []string               `json:"warnings,omitempty"`
	OriginalData  map[string]interface{} `json:"original_data,omitempty"`
}

// UnifiedTest represents a test in a unified intermediate format
type UnifiedTest struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	StartingURL    string                 `json:"starting_url,omitempty"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
	Infrastructure *Infrastructure        `json:"infrastructure,omitempty"`
	Config         *TestConfig            `json:"config,omitempty"`
	Setup          []TestStep             `json:"setup,omitempty"`
	Steps          []TestStep             `json:"steps"`
	Teardown       []TestStep             `json:"teardown,omitempty"`
	OriginalData   map[string]interface{} `json:"original_data,omitempty"`
}

// TestStep represents a single test step in unified format
type TestStep struct {
	Action     string                 `json:"action"`
	Target     string                 `json:"target,omitempty"`
	Value      string                 `json:"value,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
	Original   map[string]interface{} `json:"original,omitempty"`
}

// Infrastructure represents test infrastructure configuration
type Infrastructure struct {
	OrganizationID string                 `json:"organization_id,omitempty"`
	Project        map[string]interface{} `json:"project,omitempty"`
	Goal           map[string]interface{} `json:"goal,omitempty"`
	Journey        map[string]interface{} `json:"journey,omitempty"`
}

// TestConfig represents test configuration
type TestConfig struct {
	ContinueOnError     bool   `json:"continue_on_error,omitempty"`
	Timeout             int    `json:"timeout,omitempty"`
	ScreenshotOnFailure bool   `json:"screenshot_on_failure,omitempty"`
	OutputFormat        string `json:"output_format,omitempty"`
}

// Convert converts YAML content from detected format to target format
func (c *FormatConverter) Convert(content []byte, targetFormat detector.YAMLFormat) (*ConversionResult, error) {
	// Detect source format
	detection, err := c.detector.DetectFormat(content)
	if err != nil {
		return nil, fmt.Errorf("failed to detect format: %w", err)
	}

	result := &ConversionResult{
		SourceFormat: detection.Format,
		TargetFormat: targetFormat,
		Warnings:     detection.Warnings,
	}

	// If already in target format, return as-is
	if detection.Format == targetFormat {
		result.Success = true
		result.Output = content
		return result, nil
	}

	// Parse to unified format
	unified, err := c.parseToUnified(content, detection.Format)
	if err != nil {
		return result, fmt.Errorf("failed to parse source format: %w", err)
	}

	// Store original data
	result.OriginalData = unified.OriginalData

	// Convert from unified to target format
	output, warnings, err := c.convertFromUnified(unified, targetFormat)
	if err != nil {
		return result, fmt.Errorf("failed to convert to target format: %w", err)
	}

	result.Success = true
	result.Output = output
	result.Warnings = append(result.Warnings, warnings...)

	return result, nil
}

// parseToUnified parses any format into the unified intermediate format
func (c *FormatConverter) parseToUnified(content []byte, format detector.YAMLFormat) (*UnifiedTest, error) {
	var data map[string]interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	switch format {
	case detector.FormatCompact:
		return c.parseCompactToUnified(data)
	case detector.FormatSimplified:
		return c.parseSimplifiedToUnified(data)
	case detector.FormatExtended:
		return c.parseExtendedToUnified(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// parseCompactToUnified parses compact format to unified format
func (c *FormatConverter) parseCompactToUnified(data map[string]interface{}) (*UnifiedTest, error) {
	unified := &UnifiedTest{
		Variables:    make(map[string]interface{}),
		OriginalData: data,
	}

	// Parse basic fields
	if name, ok := data["test"].(string); ok {
		unified.Name = name
	}

	if url, ok := data["nav"].(string); ok {
		unified.StartingURL = url
	}

	// Parse variables
	if vars, ok := data["data"].(map[string]interface{}); ok {
		unified.Variables = vars
	}

	// Parse steps from do section
	if doSection, ok := data["do"].([]interface{}); ok {
		steps, err := c.parseCompactSteps(doSection)
		if err != nil {
			return nil, err
		}
		unified.Steps = steps
	}

	// Parse setup
	if setup, ok := data["setup"].([]interface{}); ok {
		steps, err := c.parseCompactSteps(setup)
		if err != nil {
			return nil, err
		}
		unified.Setup = steps
	}

	// Parse teardown
	if teardown, ok := data["teardown"].([]interface{}); ok {
		steps, err := c.parseCompactSteps(teardown)
		if err != nil {
			return nil, err
		}
		unified.Teardown = steps
	}

	return unified, nil
}

// parseCompactSteps parses compact format steps
func (c *FormatConverter) parseCompactSteps(steps []interface{}) ([]TestStep, error) {
	var result []TestStep

	for _, step := range steps {
		switch v := step.(type) {
		case string:
			// Simple wait format
			result = append(result, TestStep{
				Action: "wait",
				Value:  v,
				Original: map[string]interface{}{"wait": v},
			})

		case map[interface{}]interface{}:
			// Convert to string keys
			stepMap := make(map[string]interface{})
			for k, val := range v {
				if key, ok := k.(string); ok {
					stepMap[key] = val
				}
			}
			parsed := c.parseCompactStepMap(stepMap)
			if parsed != nil {
				result = append(result, *parsed)
			}

		case map[string]interface{}:
			parsed := c.parseCompactStepMap(v)
			if parsed != nil {
				result = append(result, *parsed)
			}
		}
	}

	return result, nil
}

// parseCompactStepMap parses a single compact step map
func (c *FormatConverter) parseCompactStepMap(step map[string]interface{}) *TestStep {
	// Try each compact action type
	if val, ok := step["c"]; ok {
		return &TestStep{
			Action:   "click",
			Target:   c.getStringValue(val),
			Original: step,
		}
	}

	if val, ok := step["t"]; ok {
		// Type can be string or map
		switch v := val.(type) {
		case string:
			return &TestStep{
				Action:   "write",
				Value:    v,
				Original: step,
			}
		case map[string]interface{}:
			// Format: t: {selector: text}
			for selector, text := range v {
				return &TestStep{
					Action:   "write",
					Target:   selector,
					Value:    c.getStringValue(text),
					Original: step,
				}
			}
		}
	}

	if val, ok := step["ch"]; ok {
		return &TestStep{
			Action:   "assert",
			Target:   c.getStringValue(val),
			Original: step,
		}
	}

	if val, ok := step["wait"]; ok {
		return &TestStep{
			Action:   "wait",
			Target:   c.getStringValue(val),
			Original: step,
		}
	}

	if val, ok := step["note"]; ok {
		return &TestStep{
			Action:   "comment",
			Value:    c.getStringValue(val),
			Original: step,
		}
	}

	if val, ok := step["nav"]; ok {
		return &TestStep{
			Action:   "navigate",
			Target:   c.getStringValue(val),
			Original: step,
		}
	}

	if val, ok := step["store"]; ok {
		if storeMap, ok := val.(map[string]interface{}); ok {
			return &TestStep{
				Action:   "store",
				Options:  storeMap,
				Original: step,
			}
		}
	}

	// Return with original data preserved
	return &TestStep{
		Original: step,
	}
}

// parseSimplifiedToUnified parses simplified format to unified format
func (c *FormatConverter) parseSimplifiedToUnified(data map[string]interface{}) (*UnifiedTest, error) {
	unified := &UnifiedTest{
		Variables:    make(map[string]interface{}),
		OriginalData: data,
	}

	// Parse basic fields
	if name, ok := data["name"].(string); ok {
		unified.Name = name
	}

	if desc, ok := data["description"].(string); ok {
		unified.Description = desc
	}

	if url, ok := data["starting_url"].(string); ok {
		unified.StartingURL = url
	}

	// Parse infrastructure
	if infra, ok := data["infrastructure"].(map[string]interface{}); ok {
		unified.Infrastructure = c.parseInfrastructure(infra)
	}

	// Parse config
	if cfg, ok := data["config"].(map[string]interface{}); ok {
		unified.Config = c.parseConfig(cfg)
	}

	// Parse steps
	if steps, ok := data["steps"].([]interface{}); ok {
		parsedSteps, err := c.parseSimplifiedSteps(steps)
		if err != nil {
			return nil, err
		}
		unified.Steps = parsedSteps
	}

	return unified, nil
}

// parseSimplifiedSteps parses simplified format steps
func (c *FormatConverter) parseSimplifiedSteps(steps []interface{}) ([]TestStep, error) {
	var result []TestStep

	for _, step := range steps {
		if stepMap, ok := step.(map[string]interface{}); ok {
			parsed := c.parseSimplifiedStepMap(stepMap)
			if parsed != nil {
				result = append(result, *parsed)
			}
		} else if stepMap, ok := step.(map[interface{}]interface{}); ok {
			// Convert to string keys
			stringMap := make(map[string]interface{})
			for k, v := range stepMap {
				if key, ok := k.(string); ok {
					stringMap[key] = v
				}
			}
			parsed := c.parseSimplifiedStepMap(stringMap)
			if parsed != nil {
				result = append(result, *parsed)
			}
		}
	}

	return result, nil
}

// parseSimplifiedStepMap parses a single simplified step map
func (c *FormatConverter) parseSimplifiedStepMap(step map[string]interface{}) *TestStep {
	// Check each possible action
	actions := []string{"navigate", "click", "write", "assert", "wait", "comment", "store", "select", "hover", "key"}
	
	for _, action := range actions {
		if val, ok := step[action]; ok {
			testStep := &TestStep{
				Action:   action,
				Original: step,
			}

			switch v := val.(type) {
			case string:
				if action == "write" || action == "comment" {
					testStep.Value = v
				} else {
					testStep.Target = v
				}
			case map[string]interface{}:
				// Complex format with options
				if selector, ok := v["selector"].(string); ok {
					testStep.Target = selector
				}
				if text, ok := v["text"].(string); ok {
					testStep.Value = text
				}
				if value, ok := v["value"].(string); ok {
					testStep.Value = value
				}
				testStep.Options = v
			}

			return testStep
		}
	}

	// Preserve unrecognized steps
	return &TestStep{
		Original: step,
	}
}

// parseExtendedToUnified parses extended format to unified format
func (c *FormatConverter) parseExtendedToUnified(data map[string]interface{}) (*UnifiedTest, error) {
	// Extended format is similar to simplified but with type/command structure
	unified, err := c.parseSimplifiedToUnified(data)
	if err != nil {
		return nil, err
	}

	// Re-parse steps with extended format
	if steps, ok := data["steps"].([]interface{}); ok {
		parsedSteps, err := c.parseExtendedSteps(steps)
		if err != nil {
			return nil, err
		}
		unified.Steps = parsedSteps
	}

	return unified, nil
}

// parseExtendedSteps parses extended format steps
func (c *FormatConverter) parseExtendedSteps(steps []interface{}) ([]TestStep, error) {
	var result []TestStep

	for _, step := range steps {
		if stepMap, ok := step.(map[string]interface{}); ok {
			parsed := c.parseExtendedStepMap(stepMap)
			if parsed != nil {
				result = append(result, *parsed)
			}
		}
	}

	return result, nil
}

// parseExtendedStepMap parses a single extended step map
func (c *FormatConverter) parseExtendedStepMap(step map[string]interface{}) *TestStep {
	testStep := &TestStep{
		Original: step,
	}

	// Get type and command
	stepType := c.getStringValue(step["type"])
	command := c.getStringValue(step["command"])
	
	// Map extended format to unified action
	switch stepType {
	case "navigate":
		testStep.Action = "navigate"
		testStep.Target = c.getStringValue(step["target"])
		
	case "interact":
		switch command {
		case "click":
			testStep.Action = "click"
		case "write", "type":
			testStep.Action = "write"
		case "hover":
			testStep.Action = "hover"
		default:
			testStep.Action = command
		}
		testStep.Target = c.getStringValue(step["target"])
		testStep.Value = c.getStringValue(step["value"])
		
	case "assert":
		testStep.Action = "assert"
		testStep.Target = c.getStringValue(step["target"])
		if command != "" && command != "exists" {
			if testStep.Options == nil {
				testStep.Options = make(map[string]interface{})
			}
			testStep.Options["type"] = command
		}
		
	case "wait":
		testStep.Action = "wait"
		if command == "time" {
			testStep.Value = c.getStringValue(step["value"])
		} else {
			testStep.Target = c.getStringValue(step["target"])
		}
		
	case "data":
		testStep.Action = "store"
		testStep.Options = step
		
	default:
		// Preserve original for unrecognized types
		testStep.Action = stepType
		if command != "" {
			if testStep.Options == nil {
				testStep.Options = make(map[string]interface{})
			}
			testStep.Options["command"] = command
		}
	}

	return testStep
}

// convertFromUnified converts from unified format to target format
func (c *FormatConverter) convertFromUnified(unified *UnifiedTest, targetFormat detector.YAMLFormat) ([]byte, []string, error) {
	var warnings []string
	var result map[string]interface{}

	switch targetFormat {
	case detector.FormatCompact:
		result, warnings = c.convertToCompact(unified)
	case detector.FormatSimplified:
		result, warnings = c.convertToSimplified(unified)
	case detector.FormatExtended:
		result, warnings = c.convertToExtended(unified)
	default:
		return nil, nil, fmt.Errorf("unsupported target format: %s", targetFormat)
	}

	// Marshal to YAML
	output, err := yaml.Marshal(result)
	if err != nil {
		return nil, warnings, err
	}

	return output, warnings, nil
}

// convertToCompact converts unified format to compact format
func (c *FormatConverter) convertToCompact(unified *UnifiedTest) (map[string]interface{}, []string) {
	var warnings []string
	result := make(map[string]interface{})

	// Convert basic fields
	result["test"] = unified.Name
	
	if unified.StartingURL != "" {
		result["nav"] = unified.StartingURL
	}

	// Convert variables
	if len(unified.Variables) > 0 {
		result["data"] = unified.Variables
	}

	// Convert setup
	if len(unified.Setup) > 0 {
		setup, warns := c.convertStepsToCompact(unified.Setup)
		warnings = append(warnings, warns...)
		if len(setup) > 0 {
			result["setup"] = setup
		}
	}

	// Convert main steps
	if len(unified.Steps) > 0 {
		steps, warns := c.convertStepsToCompact(unified.Steps)
		warnings = append(warnings, warns...)
		result["do"] = steps
	}

	// Convert teardown
	if len(unified.Teardown) > 0 {
		teardown, warns := c.convertStepsToCompact(unified.Teardown)
		warnings = append(warnings, warns...)
		if len(teardown) > 0 {
			result["teardown"] = teardown
		}
	}

	// Warn about unsupported features
	if unified.Infrastructure != nil {
		warnings = append(warnings, "Infrastructure configuration not supported in compact format")
	}
	if unified.Config != nil {
		warnings = append(warnings, "Config section not supported in compact format")
	}
	if unified.Description != "" {
		warnings = append(warnings, "Description field not supported in compact format")
	}

	return result, warnings
}

// convertStepsToCompact converts unified steps to compact format
func (c *FormatConverter) convertStepsToCompact(steps []TestStep) ([]interface{}, []string) {
	var result []interface{}
	var warnings []string

	for _, step := range steps {
		compact := c.convertStepToCompact(step)
		if compact != nil {
			result = append(result, compact)
		} else if step.Original != nil {
			// Preserve original if conversion fails
			result = append(result, step.Original)
			warnings = append(warnings, fmt.Sprintf("Could not convert step: %s", step.Action))
		}
	}

	return result, warnings
}

// convertStepToCompact converts a single unified step to compact format
func (c *FormatConverter) convertStepToCompact(step TestStep) interface{} {
	switch step.Action {
	case "click":
		return map[string]interface{}{"c": step.Target}
		
	case "write":
		if step.Target != "" && step.Target != "[focused]" {
			// Format: t: {selector: text}
			return map[string]interface{}{
				"t": map[string]interface{}{
					step.Target: step.Value,
				},
			}
		}
		// Simple format: t: text
		return map[string]interface{}{"t": step.Value}
		
	case "assert":
		return map[string]interface{}{"ch": step.Target}
		
	case "wait":
		if step.Value != "" {
			// Time wait
			return map[string]interface{}{"wait": step.Value}
		}
		// Element wait
		return map[string]interface{}{"wait": step.Target}
		
	case "comment":
		return map[string]interface{}{"note": step.Value}
		
	case "navigate":
		return map[string]interface{}{"nav": step.Target}
		
	case "store":
		if step.Options != nil {
			return map[string]interface{}{"store": step.Options}
		}
		
	case "hover":
		return map[string]interface{}{"h": step.Target}
		
	case "key":
		return map[string]interface{}{"k": step.Value}
		
	case "select":
		if step.Options != nil {
			return map[string]interface{}{"select": step.Options}
		}
	}

	// Return original if available
	return step.Original
}

// convertToSimplified converts unified format to simplified format
func (c *FormatConverter) convertToSimplified(unified *UnifiedTest) (map[string]interface{}, []string) {
	var warnings []string
	result := make(map[string]interface{})

	// Convert basic fields
	result["name"] = unified.Name
	
	if unified.Description != "" {
		result["description"] = unified.Description
	}

	if unified.StartingURL != "" {
		result["starting_url"] = unified.StartingURL
	}

	// Convert infrastructure
	if unified.Infrastructure != nil {
		result["infrastructure"] = c.convertInfrastructure(unified.Infrastructure)
	}

	// Convert config
	if unified.Config != nil {
		result["config"] = c.convertConfig(unified.Config)
	}

	// Convert steps
	if len(unified.Steps) > 0 {
		steps, warns := c.convertStepsToSimplified(unified.Steps)
		warnings = append(warnings, warns...)
		result["steps"] = steps
	}

	// Warn about unsupported features
	if len(unified.Variables) > 0 {
		warnings = append(warnings, "Variables (data section) should be embedded in steps for simplified format")
	}
	if len(unified.Setup) > 0 {
		warnings = append(warnings, "Setup section not directly supported in simplified format - consider adding to main steps")
	}
	if len(unified.Teardown) > 0 {
		warnings = append(warnings, "Teardown section not directly supported in simplified format - consider adding to main steps")
	}

	return result, warnings
}

// convertStepsToSimplified converts unified steps to simplified format
func (c *FormatConverter) convertStepsToSimplified(steps []TestStep) ([]interface{}, []string) {
	var result []interface{}
	var warnings []string

	for _, step := range steps {
		simplified := c.convertStepToSimplified(step)
		if simplified != nil {
			result = append(result, simplified)
		} else if step.Original != nil {
			// Preserve original if conversion fails
			result = append(result, step.Original)
			warnings = append(warnings, fmt.Sprintf("Could not convert step: %s", step.Action))
		}
	}

	return result, warnings
}

// convertStepToSimplified converts a single unified step to simplified format
func (c *FormatConverter) convertStepToSimplified(step TestStep) map[string]interface{} {
	result := make(map[string]interface{})

	switch step.Action {
	case "click", "navigate", "hover", "key":
		// Simple target-based actions
		result[step.Action] = step.Target
		if step.Target == "" && step.Value != "" {
			result[step.Action] = step.Value
		}
		
	case "write":
		if step.Target != "" && step.Target != "[focused]" {
			// Object format
			result["write"] = map[string]interface{}{
				"selector": step.Target,
				"text":     step.Value,
			}
		} else {
			// Simple format
			result["write"] = step.Value
		}
		
	case "assert":
		result["assert"] = step.Target
		
	case "wait":
		if step.Value != "" {
			// Time wait - convert to number if possible
			if ms, err := c.parseMilliseconds(step.Value); err == nil {
				result["wait"] = ms
			} else {
				result["wait"] = step.Value
			}
		} else {
			// Element wait
			result["wait"] = step.Target
		}
		
	case "comment":
		result["comment"] = step.Value
		
	case "store":
		if step.Options != nil {
			result["store"] = step.Options
		}
		
	case "select":
		if step.Options != nil {
			result["select"] = step.Options
		}
		
	default:
		// Preserve original or create generic structure
		if step.Original != nil {
			return step.Original
		}
		result[step.Action] = step.Target
	}

	return result
}

// convertToExtended converts unified format to extended format
func (c *FormatConverter) convertToExtended(unified *UnifiedTest) (map[string]interface{}, []string) {
	// Extended format is similar to simplified but with type/command structure
	result, warnings := c.convertToSimplified(unified)

	// Re-convert steps to extended format
	if _, ok := result["steps"].([]interface{}); ok {
		extendedSteps, warns := c.convertStepsToExtended(unified.Steps)
		warnings = append(warnings, warns...)
		result["steps"] = extendedSteps
	}

	return result, warnings
}

// convertStepsToExtended converts unified steps to extended format
func (c *FormatConverter) convertStepsToExtended(steps []TestStep) ([]interface{}, []string) {
	var result []interface{}
	var warnings []string

	for _, step := range steps {
		extended := c.convertStepToExtended(step)
		if extended != nil {
			result = append(result, extended)
		} else if step.Original != nil {
			// Preserve original if conversion fails
			result = append(result, step.Original)
			warnings = append(warnings, fmt.Sprintf("Could not convert step: %s", step.Action))
		}
	}

	return result, warnings
}

// convertStepToExtended converts a single unified step to extended format
func (c *FormatConverter) convertStepToExtended(step TestStep) map[string]interface{} {
	result := make(map[string]interface{})

	switch step.Action {
	case "navigate":
		result["type"] = "navigate"
		result["target"] = step.Target
		
	case "click", "write", "hover":
		result["type"] = "interact"
		result["command"] = step.Action
		result["target"] = step.Target
		if step.Value != "" {
			result["value"] = step.Value
		}
		
	case "assert":
		result["type"] = "assert"
		result["command"] = "exists"
		result["target"] = step.Target
		if step.Options != nil {
			if cmdType, ok := step.Options["type"].(string); ok {
				result["command"] = cmdType
			}
		}
		
	case "wait":
		result["type"] = "wait"
		if step.Value != "" {
			result["command"] = "time"
			result["value"] = step.Value
		} else {
			result["command"] = "element"
			result["target"] = step.Target
		}
		
	case "comment":
		result["type"] = "misc"
		result["command"] = "comment"
		result["value"] = step.Value
		
	case "store":
		result["type"] = "data"
		result["command"] = "store"
		if step.Options != nil {
			for k, v := range step.Options {
				if k != "type" && k != "command" {
					result[k] = v
				}
			}
		}
		
	case "key":
		result["type"] = "interact"
		result["command"] = "key"
		result["value"] = step.Value
		
	case "select":
		result["type"] = "interact"
		result["command"] = "select"
		if step.Options != nil {
			for k, v := range step.Options {
				if k != "type" && k != "command" {
					result[k] = v
				}
			}
		}
		
	default:
		// Generic extended format
		result["type"] = step.Action
		if step.Target != "" {
			result["target"] = step.Target
		}
		if step.Value != "" {
			result["value"] = step.Value
		}
	}

	return result
}

// Helper methods

func (c *FormatConverter) parseInfrastructure(data map[string]interface{}) *Infrastructure {
	infra := &Infrastructure{}
	
	if orgID, ok := data["organization_id"].(string); ok {
		infra.OrganizationID = orgID
	}
	if project, ok := data["project"].(map[string]interface{}); ok {
		infra.Project = project
	}
	if goal, ok := data["goal"].(map[string]interface{}); ok {
		infra.Goal = goal
	}
	if journey, ok := data["journey"].(map[string]interface{}); ok {
		infra.Journey = journey
	}
	
	return infra
}

func (c *FormatConverter) parseConfig(data map[string]interface{}) *TestConfig {
	cfg := &TestConfig{}
	
	if continueOnError, ok := data["continue_on_error"].(bool); ok {
		cfg.ContinueOnError = continueOnError
	}
	if timeout, ok := data["timeout"].(int); ok {
		cfg.Timeout = timeout
	}
	if screenshot, ok := data["screenshot_on_failure"].(bool); ok {
		cfg.ScreenshotOnFailure = screenshot
	}
	if format, ok := data["output_format"].(string); ok {
		cfg.OutputFormat = format
	}
	
	return cfg
}

func (c *FormatConverter) convertInfrastructure(infra *Infrastructure) map[string]interface{} {
	result := make(map[string]interface{})
	
	if infra.OrganizationID != "" {
		result["organization_id"] = infra.OrganizationID
	}
	if infra.Project != nil {
		result["project"] = infra.Project
	}
	if infra.Goal != nil {
		result["goal"] = infra.Goal
	}
	if infra.Journey != nil {
		result["journey"] = infra.Journey
	}
	
	return result
}

func (c *FormatConverter) convertConfig(cfg *TestConfig) map[string]interface{} {
	result := make(map[string]interface{})
	
	if cfg.ContinueOnError {
		result["continue_on_error"] = cfg.ContinueOnError
	}
	if cfg.Timeout > 0 {
		result["timeout"] = cfg.Timeout
	}
	if cfg.ScreenshotOnFailure {
		result["screenshot_on_failure"] = cfg.ScreenshotOnFailure
	}
	if cfg.OutputFormat != "" {
		result["output_format"] = cfg.OutputFormat
	}
	
	return result
}

func (c *FormatConverter) getStringValue(val interface{}) string {
	if s, ok := val.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", val)
}

func (c *FormatConverter) parseMilliseconds(val string) (int, error) {
	// Try to parse as number
	val = strings.TrimSpace(val)
	val = strings.TrimSuffix(val, "ms")
	return strconv.Atoi(val)
}

// DetectAndConvert detects the format and converts if needed
func (c *FormatConverter) DetectAndConvert(content []byte, preferredFormat detector.YAMLFormat) ([]byte, detector.YAMLFormat, error) {
	// Detect current format
	detection, err := c.detector.DetectFormat(content)
	if err != nil {
		return nil, detector.FormatUnknown, err
	}

	// If already in preferred format or no preference, return as-is
	if detection.Format == preferredFormat || preferredFormat == detector.FormatUnknown {
		return content, detection.Format, nil
	}

	// Convert to preferred format
	result, err := c.Convert(content, preferredFormat)
	if err != nil {
		return nil, detection.Format, err
	}

	if !result.Success {
		return nil, detection.Format, fmt.Errorf("conversion failed")
	}

	return result.Output, preferredFormat, nil
}

// ConvertToYAMLTest converts any format to core.YAMLTest for compilation
func (c *FormatConverter) ConvertToYAMLTest(content []byte) (*core.YAMLTest, error) {
	// Detect source format
	detection, err := c.detector.DetectFormat(content)
	if err != nil {
		return nil, fmt.Errorf("failed to detect format: %w", err)
	}

	// If already in compact format, parse directly
	if detection.Format == detector.FormatCompact {
		var yamlTest core.YAMLTest
		if err := yaml.Unmarshal(content, &yamlTest); err != nil {
			return nil, fmt.Errorf("failed to parse compact YAML: %w", err)
		}
		return &yamlTest, nil
	}

	// Parse to unified format first
	unified, err := c.parseToUnified(content, detection.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source format: %w", err)
	}

	// Convert unified format directly to core.YAMLTest
	yamlTest := &core.YAMLTest{
		Test: unified.Name,
		Nav:  unified.StartingURL,
		Data: unified.Variables,
	}

	// Convert steps
	if len(unified.Setup) > 0 {
		yamlTest.Setup = c.convertUnifiedStepsToActions(unified.Setup)
	}
	if len(unified.Steps) > 0 {
		yamlTest.Do = c.convertUnifiedStepsToActions(unified.Steps)
	}
	if len(unified.Teardown) > 0 {
		yamlTest.Teardown = c.convertUnifiedStepsToActions(unified.Teardown)
	}

	return yamlTest, nil
}

// convertUnifiedStepsToActions converts unified steps directly to core.Action
func (c *FormatConverter) convertUnifiedStepsToActions(steps []TestStep) []core.Action {
	var actions []core.Action
	
	for _, step := range steps {
		action := core.Action{}
		
		switch step.Action {
		case "click":
			action.C = step.Target
			
		case "write":
			if step.Target != "" && step.Target != "[focused]" {
				// Create map with interface{} values to match expected type
				m := make(map[interface{}]interface{})
				m[step.Target] = step.Value
				action.T = m
			} else {
				action.T = step.Value
			}
			
		case "assert":
			action.Ch = step.Target
			
		case "wait":
			if step.Value != "" {
				// Time wait - convert to int if possible
				if ms, err := c.parseMilliseconds(step.Value); err == nil {
					action.Wait = ms
				} else {
					action.Wait = step.Value
				}
			} else {
				// Element wait
				action.Wait = step.Target
			}
			
		case "comment":
			action.Note = step.Value
			
		case "navigate":
			action.Nav = step.Target
			
		case "store":
			action.Store = step.Options
			
		case "hover":
			action.H = step.Target
			
		case "key":
			action.K = step.Value
			
		case "select":
			action.Select = step.Options
		}
		
		actions = append(actions, action)
	}
	
	return actions
}