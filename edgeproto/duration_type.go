package edgeproto

import (
	"encoding/json"
	fmt "fmt"
	reflect "reflect"
	"time"
)

// Duration Type allows protobufs to store/manage duration
// as a raw int64 value, but accept string values via json/yaml
// marshalling for friendly user input.

type Duration int64

func (e *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err == nil {
		dur, err := time.ParseDuration(str)
		if err != nil {
			return err
		}
		*e = Duration(dur)
		return nil
	}
	var val int64
	err = unmarshal(&val)
	if err == nil {
		*e = Duration(val)
		return nil
	}
	return fmt.Errorf("Invalid duration type")
}

func (e Duration) MarshalYAML() (interface{}, error) {
	dur := time.Duration(e)
	return dur.String(), nil
}

func (e *Duration) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		dur, err := time.ParseDuration(str)
		if err != nil {
			return err
		}
		*e = Duration(dur)
		return nil
	}
	var val int64
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = Duration(val)
		return nil
	}
	return fmt.Errorf("Invalid duration type")
}

func (e Duration) MarshalJSON() ([]byte, error) {
	dur := time.Duration(e)
	return json.Marshal(dur.String())
}

func (e Duration) TimeDuration() time.Duration {
	return time.Duration(e)
}

// DecodeHook for use with mapstructure package.
func DecodeHook(from, to reflect.Type, data interface{}) (interface{}, error) {
	if from.Kind() == reflect.String {
		switch to {
		case reflect.TypeOf(Duration(0)):
			dur, err := time.ParseDuration(data.(string))
			if err != nil {
				return data, err
			}
			return Duration(dur), nil
		case reflect.TypeOf(time.Time{}):
			return time.Parse(time.RFC3339, data.(string))
		}
	}

	// decode enums
	return EnumDecodeHook(from, to, data)
}
