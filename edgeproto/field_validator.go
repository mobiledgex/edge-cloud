package edgeproto

import (
	fmt "fmt"
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

func (s *FieldValidator) CheckGT(field string, val, gt int64) {
	if s.err != nil {
		return
	}
	if val <= gt {
		s.err = fmt.Errorf("%s must be greater than %d", s.fieldDesc[field], gt)
	}
}

func (s *FieldValidator) CheckFloatGE(field string, val, gt float64) {
	if s.err != nil {
		return
	}
	if val < gt {
		s.err = fmt.Errorf("%s must be greater than or equal to %f", s.fieldDesc[field], gt)
	}
}

func (s *FieldValidator) CheckLT(field string, val, lt int64) {
	if s.err != nil {
		return
	}
	if val >= lt {
		s.err = fmt.Errorf("%s must be less than %d", s.fieldDesc[field], lt)
	}
}
