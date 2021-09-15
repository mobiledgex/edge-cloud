package util

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestObj struct {
	Field1 EmptyStringJsonNumber `json:"field-1"`
}

func TestJson(t *testing.T) {
	obj := TestObj{}
	err := json.Unmarshal([]byte(`{"field-1": "1.1"}`), &obj)
	require.Nil(t, err)
	require.Equal(t, string(obj.Field1), "1.1")

	obj = TestObj{}
	err = json.Unmarshal([]byte(`{"field-1": "1"}`), &obj)
	require.Nil(t, err)
	require.Equal(t, string(obj.Field1), "1")

	obj = TestObj{}
	err = json.Unmarshal([]byte(`{"field-1": ""}`), &obj)
	require.Nil(t, err)
	require.Equal(t, string(obj.Field1), "")

	obj = TestObj{}
	err = json.Unmarshal([]byte(`{"field-1": "abcd"}`), &obj)
	require.NotNil(t, err)
}
