package repository

import "encoding/json"

func marshalJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func unmarshalJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, v)
}
