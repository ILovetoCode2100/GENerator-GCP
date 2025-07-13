package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

// Manager handles request body templates
type Manager struct {
	templates map[string]*template.Template
}

// NewManager creates a new template manager
func NewManager() *Manager {
	return &Manager{
		templates: make(map[string]*template.Template),
	}
}

// RegisterTemplate registers a new template with validation
func (m *Manager) RegisterTemplate(name, tmplStr string) error {
	// Parse and validate template
	tmpl, err := template.New(name).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("invalid template %s: %w", name, err)
	}

	m.templates[name] = tmpl
	return nil
}

// Execute fills a template with the provided data
func (m *Manager) Execute(name string, data interface{}) (string, error) {
	tmpl, exists := m.templates[name]
	if !exists {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	// Validate JSON output if applicable
	result := buf.String()
	if isJSON(result) {
		var js json.RawMessage
		if err := json.Unmarshal([]byte(result), &js); err != nil {
			return "", fmt.Errorf("template produced invalid JSON: %w", err)
		}
	}

	return result, nil
}

// LoadDefaults loads default templates for common operations
func (m *Manager) LoadDefaults() error {
	// Example templates - these would be generated from OpenAPI
	defaults := map[string]string{
		"create_user": `{
			"name": "{{.Name}}",
			"email": "{{.Email}}",
			"role": "{{.Role | default "user"}}"
		}`,

		"update_item": `{
			"id": {{.ID}},
			"status": "{{.Status}}",
			"updated_at": "{{.Timestamp}}"
		}`,

		"search_query": `{
			"query": "{{.Query}}",
			"filters": {
				{{if .Category}}"category": "{{.Category}}",{{end}}
				{{if .MinPrice}}"min_price": {{.MinPrice}},{{end}}
				{{if .MaxPrice}}"max_price": {{.MaxPrice}},{{end}}
				"limit": {{.Limit | default 10}}
			}
		}`,
	}

	for name, tmpl := range defaults {
		if err := m.RegisterTemplate(name, tmpl); err != nil {
			return err
		}
	}

	return nil
}

// isJSON checks if a string is valid JSON
func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// TemplateData provides a safe container for template parameters
type TemplateData struct {
	data map[string]interface{}
}

// NewTemplateData creates a new template data container
func NewTemplateData() *TemplateData {
	return &TemplateData{
		data: make(map[string]interface{}),
	}
}

// Set adds a validated parameter to the template data
func (td *TemplateData) Set(key string, value interface{}) error {
	// Validate key to prevent injection
	if !isValidKey(key) {
		return fmt.Errorf("invalid parameter key: %s", key)
	}

	// Validate value types
	switch v := value.(type) {
	case string, int, int64, float64, bool:
		td.data[key] = v
	case []string, []int:
		td.data[key] = v
	default:
		return fmt.Errorf("unsupported value type for %s", key)
	}

	return nil
}

// Get returns the template data map
func (td *TemplateData) Get() map[string]interface{} {
	return td.data
}

// isValidKey checks if a parameter key is safe
func isValidKey(key string) bool {
	// Only allow alphanumeric and underscore
	for _, r := range key {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_') {
			return false
		}
	}
	return len(key) > 0 && len(key) < 100
}
