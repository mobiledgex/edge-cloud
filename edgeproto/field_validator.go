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

package edgeproto

import (
	fmt "fmt"
	reflect "reflect"
)

type CompareType int

const (
	CompareGT CompareType = iota
	CompareGTE
	CompareLT
)

type FieldValidator struct {
	err       error
	fieldDesc map[string]string
}

func NewFieldValidator(allFieldsStringMap map[string]string) *FieldValidator {
	v := FieldValidator{}
	v.fieldDesc = allFieldsStringMap
	return &v
}

func (s *FieldValidator) CheckGT(field string, val, gt interface{}) {
	s.Check(field, val, gt, CompareGT)
}

func (s *FieldValidator) CheckGTE(field string, val, gte interface{}) {
	s.Check(field, val, gte, CompareGTE)
}

func (s *FieldValidator) CheckLT(field string, val, lt interface{}) {
	s.Check(field, val, lt, CompareLT)
}

func (s *FieldValidator) Check(field string, valI, gtI interface{}, ct CompareType) {
	if s.err != nil {
		return
	}
	if valI == nil || gtI == nil {
		return
	}
	desc := s.fieldDesc[field]
	valType := reflect.TypeOf(valI)
	gtType := reflect.TypeOf(gtI)
	if valType != gtType {
		s.err = fmt.Errorf("Validator: cannot compare %s to %s for %s", valType.String(), gtType.String(), desc)
		return
	}
	failVal := ""

	switch valType {
	case reflect.TypeOf(Duration(0)):
		val := valI.(Duration)
		gt := gtI.(Duration)
		if ct == CompareGT && val <= gt ||
			ct == CompareGTE && val < gt ||
			ct == CompareLT && val >= gt {
			failVal = gt.TimeDuration().String()
		}
	case reflect.TypeOf(int64(0)):
		val := valI.(int64)
		gt := gtI.(int64)
		if ct == CompareGT && val <= gt ||
			ct == CompareGTE && val < gt ||
			ct == CompareLT && val >= gt {
			failVal = fmt.Sprintf("%d", gt)
		}
	case reflect.TypeOf(int32(0)):
		val := valI.(int32)
		gt := gtI.(int32)
		if ct == CompareGT && val <= gt ||
			ct == CompareGTE && val < gt ||
			ct == CompareLT && val >= gt {
			failVal = fmt.Sprintf("%d", gt)
		}
	case reflect.TypeOf(uint32(0)):
		val := valI.(uint32)
		gt := gtI.(uint32)
		if ct == CompareGT && val <= gt ||
			ct == CompareGTE && val < gt ||
			ct == CompareLT && val >= gt {
			failVal = fmt.Sprintf("%d", gt)
		}
	case reflect.TypeOf(float32(0)):
		val := valI.(float32)
		gt := gtI.(float32)
		if ct == CompareGT && val <= gt ||
			ct == CompareGTE && val < gt ||
			ct == CompareLT && val >= gt {
			failVal = fmt.Sprintf("%g", gt)
		}
	case reflect.TypeOf(float64(0)):
		val := valI.(float64)
		gt := gtI.(float64)
		if ct == CompareGT && val <= gt ||
			ct == CompareGTE && val < gt ||
			ct == CompareLT && val >= gt {
			failVal = fmt.Sprintf("%g", gt)
		}
	default:
		s.err = fmt.Errorf("Unhandled type %s", valType.String())
	}

	if failVal != "" {
		switch ct {
		case CompareGT:
			s.err = fmt.Errorf("%s must be greater than %s", desc, failVal)
		case CompareGTE:
			s.err = fmt.Errorf("%s must be greater than or equal to %s", desc, failVal)
		case CompareLT:
			s.err = fmt.Errorf("%s must be less than %s", desc, failVal)
		}
	}
}
