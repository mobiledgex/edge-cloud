package util

import (
	"encoding/json"
)

// EmptyStringJsonNumber behaves like JsonNumber for valid numbers,
// but treats the empty string differently. JsonNumber marshals out the empty
// string as 0, but this marshals it out as the empty string. JsonNumber will fail
// to unmarshal the empty string, but this will allow it.
// This allows the field to distinguish between an unspecified value (the empty string)
// and a zero value.
type EmptyStringJsonNumber json.Number

func (s EmptyStringJsonNumber) MarshalJSON() ([]byte, error) {
	if string(s) == "" {
		return json.Marshal("")
	}
	return json.Marshal(s)

}

func (s *EmptyStringJsonNumber) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil && str == "" {
		return nil
	}
	var val json.Number
	err = json.Unmarshal(b, &val)
	if err == nil {
		*s = EmptyStringJsonNumber(val)
		return nil
	}
	return err
}
