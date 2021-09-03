package util

import (
	"encoding/json"
)

// Same as `json.Number`, but ignores empty strings
// Empty strings fail for `json.Number`
type CustomJsonNumber json.Number

func (s CustomJsonNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(s)
}

func (s *CustomJsonNumber) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil && str == "" {
		return nil
	}
	var val json.Number
	err = json.Unmarshal(b, &val)
	if err == nil {
		*s = CustomJsonNumber(val)
		return nil
	}
	return err
}
