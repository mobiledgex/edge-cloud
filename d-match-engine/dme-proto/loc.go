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

package distributed_match_engine

import (
	"encoding/json"
	fmt "fmt"
	strings "strings"
	"time"
)

func TimeToTimestamp(t time.Time) Timestamp {
	ts := Timestamp{}
	ts.Seconds = t.Unix()
	ts.Nanos = int32(t.Nanosecond())
	return ts
}

func TimestampToTime(ts Timestamp) time.Time {
	return time.Unix(ts.Seconds, int64(ts.Nanos))
}

func ParseTimestamp(str string) (Timestamp, error) {
	str = strings.TrimSpace(str)
	if str == "" {
		return Timestamp{}, nil
	}
	t, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		return Timestamp{}, err
	}
	return TimeToTimestamp(t), nil
}

func (s *Timestamp) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err == nil {
		ts, err := ParseTimestamp(str)
		if err != nil {
			return err
		}
		*s = ts
		return nil
	}
	return fmt.Errorf("Invalid RFC3339 timestamp %v", err)
}

func (s Timestamp) MarshalYAML() (interface{}, error) {
	if s.Seconds == 0 && s.Nanos == 0 {
		return "", nil
	}
	t := TimestampToTime(s)
	return t.Format(time.RFC3339Nano), nil
}

// To be able to unmarshal timestamp as written out before
// the custom JSON marshaler, the expanded object json should be
// unmarshaled as if there is no custom marshaler. But we can't
// do that with the new object, so we have to use a copied version
// of the old object to unmarshal into.
type oldUnmarshalTimestamp struct {
	Seconds int64 `json:"seconds,omitempty"`
	Nanos   int32 `json:"nanos,omitempty"`
}

func (s *Timestamp) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		ts, err := ParseTimestamp(str)
		if err != nil {
			return err
		}
		*s = ts
		return nil
	} else {
		// for backwards compatibility, support marshaled object
		ts := oldUnmarshalTimestamp{}
		errObj := json.Unmarshal(b, &ts)
		if errObj == nil {
			s.Seconds = ts.Seconds
			s.Nanos = ts.Nanos
			return nil
		}
	}
	return fmt.Errorf("Invalid RFC3339 timestamp %v", err)
}

func (s Timestamp) MarshalJSON() ([]byte, error) {
	str := ""
	if s.Seconds != 0 || s.Nanos != 0 {
		t := TimestampToTime(s)
		str = t.Format(time.RFC3339Nano)
	}
	return json.Marshal(str)
}
