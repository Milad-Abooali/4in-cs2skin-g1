package utils

// SafeExtractErrorStatus safely extracts error code and status from a map
func SafeExtractErrorStatus(resp map[string]interface{}) (errorCode int, status int, errType string) {
	// Extract error code
	if v, ok := resp["error"].(float64); ok {
		errorCode = int(v)
	} else if v, ok := resp["error"].(int); ok {
		errorCode = v
	} else {
		errorCode = 0
	}

	// Extract status
	if v, ok := resp["status"].(float64); ok {
		status = int(v)
	} else if v, ok := resp["status"].(int); ok {
		status = v
	} else {
		status = 0
	}

	// Extract type
	if v, ok := resp["type"].(string); ok {
		errType = v
	} else {
		errType = ""
	}

	return
}
