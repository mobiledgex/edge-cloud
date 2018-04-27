package util

import (
	"errors"
)

var GrpcFieldsErrUninitialized = errors.New("byte array uninitialized")
var GrpcFieldsErrOutOfRange = errors.New("value out of range for byte array")

var GrpcFieldsSize uint = 128
var GrpcFieldsByte uint = 8

func GrpcFieldsNew() []byte {
	return make([]byte, GrpcFieldsSize/GrpcFieldsByte)
}

func convert(arr []byte, val uint) (uint, uint, error) {
	if arr == nil {
		return 0, 0, GrpcFieldsErrUninitialized
	}
	size := uint(len(arr))
	index := val / GrpcFieldsByte
	if index >= size {
		return 0, 0, GrpcFieldsErrOutOfRange
	}
	offset := val % GrpcFieldsByte
	return index, offset, nil
}

func GrpcFieldsSet(arr []byte, val uint) error {
	index, offset, err := convert(arr, val)
	if err != nil {
		return err
	}
	arr[index] |= (1 << offset)
	return nil
}

func GrpcFieldsClear(arr []byte, val uint) error {
	index, offset, err := convert(arr, val)
	if err != nil {
		return err
	}
	arr[index] &= ^(1 << offset)
	return nil
}

func GrpcFieldsGet(arr []byte, val uint) (bool, error) {
	if arr == nil {
		// fields not set in message, this defaults
		// to setting all fields
		return true, nil
	}
	index, offset, err := convert(arr, val)
	if err != nil {
		return false, err
	}
	if arr[index]&(1<<offset) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func GrpcFieldsSetAll(arr []byte) {
	for ii, _ := range arr {
		arr[ii] = 0xff
	}
}

func GrpcFieldsClearAll(arr []byte) {
	for ii, _ := range arr {
		arr[ii] = 0
	}
}
