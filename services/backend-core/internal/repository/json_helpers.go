package repository

import "encoding/json"

func mustMarshalJSON(v interface{}) []byte {
	if v == nil {
		return []byte("{}")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}

func mustUnmarshalJSON(data []byte) map[string]interface{} {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return map[string]interface{}{}
	}
	return v
}
