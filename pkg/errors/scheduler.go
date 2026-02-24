package errors

import "errors"

var (
	ErrKeyExists = errors.New("redis: key already exists")
	ErrInvalidValue = errors.New("redis: invalid value type")

	ErrWaitTimeout = errors.New("wait time out")
)