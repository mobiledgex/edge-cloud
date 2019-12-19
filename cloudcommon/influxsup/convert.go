package influxsup

import (
	"encoding/json"
	"fmt"
	"time"
)

// Convert read measurement interfaces to expected types

func ConvString(val interface{}) (string, error) {
	if str, ok := val.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("value (%T)%v not a string", val, val)
}

func ConvInt(val interface{}) (int64, error) {
	if v, ok := val.(json.Number); ok {
		i, err := v.Int64()
		if err != nil {
			return 0, err
		}
		return i, nil
	}
	return 0, fmt.Errorf("value (%T)%v not a json number", val, val)
}

func ConvFloat(val interface{}) (float64, error) {
	if v, ok := val.(json.Number); ok {
		f, err := v.Float64()
		if err != nil {
			return 0, err
		}
		return f, nil
	}
	return 0, fmt.Errorf("value (%T)%v not a json number", val, val)
}

func ConvTime(val interface{}) (time.Time, error) {
	var ts0 time.Time
	if v, ok := val.(string); ok {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return ts0, err
		}
		return t, nil
	}
	return ts0, fmt.Errorf("time value (%T)%v not a string", val, val)
}
