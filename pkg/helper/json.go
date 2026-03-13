package helper

import "encoding/json"

func ParseJSON(raw []byte) map[string]interface{} {
	if raw == nil {
		return nil
	}

	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}

	return m
}

func LookupKey(m map[string]interface{}, key string) interface{} {
	if m == nil {
		return ""
	}

	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}

	return v
}
