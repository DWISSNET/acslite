package repository

import (
	"encoding/json"
	"fmt"
)

func marshalJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshalJSON: %w", err)
	}
	return string(b), nil
}

func unmarshalJSON(s string, v interface{}) error {
	if s == "" || s == "null" {
		return nil
	}
	return json.Unmarshal([]byte(s), v)
}

func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
