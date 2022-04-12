// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
