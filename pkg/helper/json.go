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

func ExtractNestedField(raw []byte, keys ...string) interface{} {
	if len(keys) == 0 {
		return ""
	}

	current := ParseJSON(raw)
	if current == nil {
		return ""
	}

	for _, k := range keys[:len(keys)-1] {
		v, ok := current[k]
		if !ok || v == nil {
			return ""
		}
		next, ok := v.(map[string]interface{})
		if !ok {
			return ""
		}
		current = next
	}

	return LookupKey(current, keys[len(keys)-1])
}
