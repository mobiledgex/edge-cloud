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
