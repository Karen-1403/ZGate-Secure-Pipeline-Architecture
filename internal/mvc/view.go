package mvc

import "time"

// JSONResponseView formats responses into JSON-friendly maps
type JSONResponseView struct{}

func (v *JSONResponseView) RenderSuccess(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status":    "success",
		"data":      data,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
}

func (v *JSONResponseView) RenderError(message string, code string) map[string]interface{} {
	return map[string]interface{}{
		"status":    "error",
		"error":     message,
		"code":      code,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
}
