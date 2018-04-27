package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrpcFields(t *testing.T) {
	arr := GrpcFieldsNew()

	err := GrpcFieldsSet(arr, 0)
	assert.Nil(t, err, "Set")

	err = GrpcFieldsSet(arr, 7)
	assert.Nil(t, err, "Set")

	err = GrpcFieldsSet(arr, 127)
	assert.Nil(t, err, "Set")

	assert.Equal(t, "\x81\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x80", string(arr), "Some fields")

	val, err := GrpcFieldsGet(arr, 126)
	assert.Nil(t, err, "Get")
	assert.False(t, val, "Get")

	val, err = GrpcFieldsGet(arr, 1)
	assert.Nil(t, err, "Get")
	assert.False(t, val, "Get")

	val, err = GrpcFieldsGet(arr, 7)
	assert.Nil(t, err, "Get")
	assert.True(t, val, "Get")

	val, err = GrpcFieldsGet(arr, 127)
	assert.Nil(t, err, "Get")
	assert.True(t, val, "Get")

	err = GrpcFieldsClear(arr, 7)
	assert.Nil(t, err, "Clear")

	assert.Equal(t, "\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x80", string(arr), "Some fields")

	val, err = GrpcFieldsGet(arr, 7)
	assert.Nil(t, err, "Get")
	assert.False(t, val, "Get")

	err = GrpcFieldsSet(arr, 128)
	assert.Equal(t, GrpcFieldsErrOutOfRange, err, "Set invalid")

	val, err = GrpcFieldsGet(arr, 1128)
	assert.Equal(t, GrpcFieldsErrOutOfRange, err, "Get invalid")
	assert.False(t, val, "Get")

	err = GrpcFieldsClear(arr, 129)
	assert.Equal(t, GrpcFieldsErrOutOfRange, err, "Clear invalid")

	var arrbad []byte
	err = GrpcFieldsSet(arrbad, 0)
	assert.Equal(t, GrpcFieldsErrUninitialized, err, "Uninitialized")

	GrpcFieldsSetAll(arr)
	assert.Equal(t, "\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff", string(arr), "All fields")

	GrpcFieldsClearAll(arr)
	assert.Equal(t, "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00", string(arr), "No fields")
}
