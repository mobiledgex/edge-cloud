# Test utilities

This package is used for unit-test support code. It should only be included in unit test code, and never in main or shipping code. It is part of its own package so that it can be referenced from other package's unit test files (_test.go files from other packages cannot be referenced in a unit test file).

Some of these files may be auto-generated from protoc-gen-test.
