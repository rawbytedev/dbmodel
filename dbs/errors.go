package dbs

import "errors"

var (
	ErrEmptydbKey   = errors.New("key is Empty")
	ErrEmptydbValue = errors.New("value is Empty")
	ErrInvalidDB    = errors.New("invalid DB")
)
