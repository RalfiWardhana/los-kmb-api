package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
)

func Sanitized(c echo.Context) (echo.Context, error) {
	// Read the raw request body
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c, errors.New("failed to read request body")
	}

	// Parse JSON into a generic map
	var jsonData map[string]interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return c, errors.New("invalid JSON format")
	}

	// Initialize Bluemonday sanitizer
	sanitizer := bluemonday.StrictPolicy()

	// Recursively sanitize JSON fields
	sanitizeJSONFields(&jsonData, sanitizer)

	// Re-encode sanitized JSON back into the request body
	sanitizedBody, err := json.Marshal(jsonData)
	if err != nil {
		return c, errors.New("error encoded")
	}

	// Replace the request body with the sanitized JSON for binding
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(sanitizedBody))

	// Bind the sanitized data to the target struct
	return c, nil
}

// Recursive function to sanitize all string fields in JSON
func sanitizeJSONFields(data *map[string]interface{}, sanitizer *bluemonday.Policy) {
	for key, value := range *data {
		switch v := value.(type) {
		case string:
			// Sanitize string fields
			(*data)[key] = sanitizer.Sanitize(v)
		case map[string]interface{}:
			// Recursively sanitize nested JSON objects
			sanitizeJSONFields(&v, sanitizer)
		case []interface{}:
			// Sanitize arrays
			for i, elem := range v {
				if str, ok := elem.(string); ok {
					v[i] = sanitizer.Sanitize(str)
				} else if obj, ok := elem.(map[string]interface{}); ok {
					sanitizeJSONFields(&obj, sanitizer)
				}
			}
		}
	}
}
