package validate

import (
	"encoding/json"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"strconv"
	"strings"
)

func RequireString(data map[string]any, field string, allowEmpty bool) (string, models.HandlerError, bool) {
	var appError models.HandlerError

	// 1) must exist
	raw, ok := data[field]
	if !ok {
		appError.Type = "REQUIRED_FIELD_MISSING"
		appError.Code = 5001
		appError.Data = map[string]any{
			"fieldName": field,
			"fieldType": "string",
		}
		return "", appError, false
	}

	// 2) must be a string or json.Number
	var s string
	switch v := raw.(type) {
	case string:
		s = strings.TrimSpace(v)
	case json.Number:
		s = strings.TrimSpace(v.String())
	default:
		appError.Type = "INVALID_TYPE_OR_FORMAT"
		appError.Code = 5003
		appError.Data = map[string]any{
			"fieldName": field,
			"fieldType": "string",
		}
		return "", appError, false
	}

	// 3) check empty rule
	if !allowEmpty && s == "" {
		appError.Type = "FIELD_EMPTY"
		appError.Code = 5002
		appError.Data = map[string]any{
			"fieldName": field,
			"fieldType": "string",
		}
		return "", appError, false
	}

	return s, models.HandlerError{}, true
}

func RequireStringIn(data map[string]any, field string, allowed []string) (string, models.HandlerError, bool) {
	var appError models.HandlerError

	// 1) must exist
	raw, ok := data[field]
	if !ok {
		appError.Type = "REQUIRED_FIELD_MISSING"
		appError.Code = 5001
		appError.Data = map[string]any{
			"fieldName": field,
			"fieldType": "string",
		}
		return "", appError, false
	}

	// 2) must be non-empty (after trimming)
	var s string
	switch v := raw.(type) {
	case string:
		s = strings.TrimSpace(v)
	case json.Number:
		s = strings.TrimSpace(v.String())
	default:
		appError.Type = "INVALID_TYPE_OR_FORMAT"
		appError.Code = 5003
		appError.Data = map[string]any{
			"fieldName": field,
			"fieldType": "string",
		}
		return "", appError, false
	}

	if s == "" {
		appError.Type = "FIELD_EMPTY"
		appError.Code = 5002
		appError.Data = map[string]any{
			"fieldName": field,
			"fieldType": "string",
		}
		return "", appError, false
	}

	// 3) if allowed list provided, enforce membership (exact match)
	if len(allowed) > 0 {
		found := false
		for _, a := range allowed {
			if s == a { // exact match; use strings.EqualFold for case-insensitive
				found = true
				break
			}
		}
		if !found {
			appError.Type = "INVALID_TYPE_OR_FORMAT"
			appError.Code = 5003
			appError.Data = map[string]any{
				"fieldName": field,
				"fieldType": "string",
			}
			return "", appError, false
		}
	}

	return s, models.HandlerError{}, true
}

func RequireInt(data map[string]any, field string) (int64, models.HandlerError, bool) {
	var appError models.HandlerError

	// 1) must exist
	raw, ok := data[field]
	if !ok {
		appError.Type = "REQUIRED_FIELD_MISSING"
		appError.Code = 5001
		appError.Data = map[string]interface{}{
			"fieldName": field,
			"fieldType": "int",
		}
		return 0, appError, false
	}

	// 2) must be "non-empty"
	// For int, we interpret "empty" as nil or "", NOT zero. Zero is considered a valid value.
	if raw == nil {
		appError.Type = "FIELD_EMPTY"
		appError.Code = 5002
		appError.Data = map[string]interface{}{
			"fieldName": field,
			"fieldType": "int",
		}
		return 0, appError, false
	}

	// 3) cast to int64 (handle common JSON cases)
	switch v := raw.(type) {
	case float64: // default for numbers in map[string]any after json.Unmarshal
		return int64(v), appError, true
	case int:
		return int64(v), appError, true
	case int64:
		return v, appError, true
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			appError.Type = "INVALID_TYPE_OR_FORMAT"
			appError.Code = 5003
			appError.Data = map[string]interface{}{
				"fieldName": field,
				"fieldType": "int",
			}
			return 0, appError, false
		}
		return n, appError, true
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			appError.Type = "FIELD_EMPTY"
			appError.Code = 5002
			appError.Data = map[string]interface{}{
				"fieldName": field,
				"fieldType": "int",
			}
			return 0, appError, false
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			appError.Type = "INVALID_TYPE_OR_FORMAT"
			appError.Code = 5003
			appError.Data = map[string]interface{}{
				"fieldName": field,
				"fieldType": "int",
			}
			return 0, appError, false
		}
		return n, appError, true
	default:
		appError.Type = "INVALID_TYPE_OR_FORMAT"
		appError.Code = 5003
		appError.Data = map[string]interface{}{
			"fieldName": field,
			"fieldType": "int",
		}
		return 0, appError, false
	}
}

func RequireFloat(data map[string]any, field string) (float64, models.HandlerError, bool) {
	var appError models.HandlerError

	// 1) must exist
	raw, ok := data[field]
	if !ok {
		appError.Type = "REQUIRED_FIELD_MISSING"
		appError.Code = 5001
		appError.Data = map[string]interface{}{
			"fieldName": field,
			"fieldType": "float",
		}
		return 0, appError, false
	}

	// 2) must be "non-empty"
	if raw == nil {
		appError.Type = "FIELD_EMPTY"
		appError.Code = 5002
		appError.Data = map[string]interface{}{
			"fieldName": field,
			"fieldType": "float",
		}
		return 0, appError, false
	}

	// 3) cast to float64 (handle common JSON cases)
	switch v := raw.(type) {
	case float64: // default for numbers in map[string]any after json.Unmarshal
		return v, appError, true
	case float32:
		return float64(v), appError, true
	case int:
		return float64(v), appError, true
	case int64:
		return float64(v), appError, true
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			appError.Type = "INVALID_TYPE_OR_FORMAT"
			appError.Code = 5003
			appError.Data = map[string]interface{}{
				"fieldName": field,
				"fieldType": "float",
			}
			return 0, appError, false
		}
		return f, appError, true
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			appError.Type = "FIELD_EMPTY"
			appError.Code = 5002
			appError.Data = map[string]interface{}{
				"fieldName": field,
				"fieldType": "float",
			}
			return 0, appError, false
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			appError.Type = "INVALID_TYPE_OR_FORMAT"
			appError.Code = 5003
			appError.Data = map[string]interface{}{
				"fieldName": field,
				"fieldType": "float",
			}
			return 0, appError, false
		}
		return f, appError, true
	default:
		appError.Type = "INVALID_TYPE_OR_FORMAT"
		appError.Code = 5003
		appError.Data = map[string]interface{}{
			"fieldName": field,
			"fieldType": "float",
		}
		return 0, appError, false
	}
}

func RequireBool(data map[string]any, field string) (bool, models.HandlerError, bool) {
	var appError models.HandlerError

	raw, ok := data[field]
	if !ok {
		appError.Type = "REQUIRED_FIELD_MISSING"
		appError.Code = 5001
		appError.Data = map[string]interface{}{
			"fieldName": field,
			"fieldType": "bool",
		}
		return false, appError, false
	}
	switch v := raw.(type) {
	case bool:
		return v, appError, true
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			appError.Type = "FIELD_EMPTY"
			appError.Code = 5002
			appError.Data = map[string]interface{}{
				"fieldName": field,
				"fieldType": "bool",
			}
			return false, appError, false
		}
		b, err := strconv.ParseBool(s)
		if err != nil {
			appError.Type = "INVALID_TYPE_OR_FORMAT"
			appError.Code = 5003
			appError.Data = map[string]interface{}{
				"fieldName": field,
				"fieldType": "bool",
			}
			return false, appError, false
		}
		return b, appError, true
	default:
		appError.Type = "INVALID_TYPE_OR_FORMAT"
		appError.Code = 5003
		appError.Data = map[string]interface{}{
			"fieldName": field,
			"fieldType": "bool",
		}
		return false, appError, false
	}
}
