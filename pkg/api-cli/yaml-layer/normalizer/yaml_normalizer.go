package normalizer

import (
	"fmt"
	"reflect"
)

// YAMLNormalizer handles conversion of map[interface{}]interface{} to map[string]interface{}
type YAMLNormalizer struct {
	// Options for normalization
	PreserveNil bool
	DebugMode   bool
}

// NewYAMLNormalizer creates a new YAML normalizer
func NewYAMLNormalizer() *YAMLNormalizer {
	return &YAMLNormalizer{
		PreserveNil: false,
		DebugMode:   false,
	}
}

// Normalize converts any interface{} to a JSON-compatible structure
func (yn *YAMLNormalizer) Normalize(data interface{}) (interface{}, error) {
	return yn.normalizeValue(data), nil
}

// normalizeValue recursively normalizes a value
func (yn *YAMLNormalizer) normalizeValue(value interface{}) interface{} {
	if value == nil {
		if yn.PreserveNil {
			return nil
		}
		return ""
	}

	switch v := value.(type) {
	case map[interface{}]interface{}:
		return yn.normalizeMap(v)

	case map[string]interface{}:
		// Already normalized, but check nested values
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = yn.normalizeValue(val)
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = yn.normalizeValue(item)
		}
		return result

	case []map[interface{}]interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = yn.normalizeMap(item)
		}
		return result

	case []map[string]interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = yn.normalizeValue(item)
		}
		return result

	default:
		// Use reflection for other slice types
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice {
			result := make([]interface{}, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				result[i] = yn.normalizeValue(rv.Index(i).Interface())
			}
			return result
		}

		// Primitive types pass through
		return value
	}
}

// normalizeMap converts map[interface{}]interface{} to map[string]interface{}
func (yn *YAMLNormalizer) normalizeMap(m map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range m {
		// Convert key to string
		keyStr := yn.keyToString(key)

		// Recursively normalize value
		result[keyStr] = yn.normalizeValue(value)
	}

	return result
}

// keyToString converts an interface{} key to string
func (yn *YAMLNormalizer) keyToString(key interface{}) string {
	switch k := key.(type) {
	case string:
		return k
	case int:
		return fmt.Sprintf("%d", k)
	case int64:
		return fmt.Sprintf("%d", k)
	case float64:
		return fmt.Sprintf("%g", k)
	case bool:
		return fmt.Sprintf("%v", k)
	default:
		// Use fmt.Sprint for other types
		return fmt.Sprint(key)
	}
}

// NormalizeYAMLAction normalizes a single action from YAML
func (yn *YAMLNormalizer) NormalizeYAMLAction(action interface{}) (map[string]interface{}, error) {
	// First normalize the entire structure
	normalized := yn.normalizeValue(action)

	// Ensure it's a map
	switch v := normalized.(type) {
	case map[string]interface{}:
		return v, nil
	default:
		return nil, fmt.Errorf("action is not a map after normalization: %T", normalized)
	}
}

// NormalizeYAMLTest normalizes an entire test structure
func (yn *YAMLNormalizer) NormalizeYAMLTest(test interface{}) (map[string]interface{}, error) {
	normalized := yn.normalizeValue(test)

	switch v := normalized.(type) {
	case map[string]interface{}:
		// Ensure required fields exist
		if _, ok := v["test"]; !ok {
			v["test"] = "Unnamed Test"
		}

		// Normalize the 'do' field
		if doField, ok := v["do"]; ok {
			switch actions := doField.(type) {
			case []interface{}:
				normalizedActions := make([]interface{}, len(actions))
				for i, action := range actions {
					normalizedAction, err := yn.NormalizeYAMLAction(action)
					if err != nil {
						return nil, fmt.Errorf("error normalizing action %d: %w", i, err)
					}
					normalizedActions[i] = normalizedAction
				}
				v["do"] = normalizedActions
			default:
				return nil, fmt.Errorf("'do' field is not an array: %T", actions)
			}
		}

		return v, nil

	default:
		return nil, fmt.Errorf("test is not a map after normalization: %T", normalized)
	}
}

// NormalizeForJSON ensures the data structure is JSON-compatible
func (yn *YAMLNormalizer) NormalizeForJSON(data interface{}) interface{} {
	return yn.normalizeValue(data)
}

// DeepEqual compares two normalized structures
func (yn *YAMLNormalizer) DeepEqual(a, b interface{}) bool {
	// Normalize both values
	normalizedA := yn.normalizeValue(a)
	normalizedB := yn.normalizeValue(b)

	return reflect.DeepEqual(normalizedA, normalizedB)
}

// ExtractField safely extracts a field from a potentially unnormalized map
func (yn *YAMLNormalizer) ExtractField(data interface{}, field string) (interface{}, bool) {
	// First normalize the data
	normalized := yn.normalizeValue(data)

	// Extract from normalized map
	switch v := normalized.(type) {
	case map[string]interface{}:
		val, ok := v[field]
		return val, ok
	default:
		return nil, false
	}
}

// SetField safely sets a field in a potentially unnormalized map
func (yn *YAMLNormalizer) SetField(data interface{}, field string, value interface{}) (map[string]interface{}, error) {
	// Normalize the data
	normalized := yn.normalizeValue(data)

	// Ensure it's a map
	switch v := normalized.(type) {
	case map[string]interface{}:
		v[field] = value
		return v, nil
	default:
		// Create a new map
		return map[string]interface{}{
			field: value,
		}, nil
	}
}
