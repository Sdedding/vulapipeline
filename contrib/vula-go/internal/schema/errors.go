package schema

import (
	"fmt"
	"reflect"
)

type ErrKey struct {
	Key string
	Err error
}

func (e *ErrKey) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("key '%s' not found", e.Key)
	}
	return fmt.Sprintf("key '%s' not found: %v", e.Key, e.Err)
}

func (e *ErrKey) Unwrap() error {
	return e.Err
}

type ErrType struct {
	Message string
}

func (e *ErrType) Error() string {
	return fmt.Sprintf("type error: %s", e.Message)
}

func errRead(v, t any) *ErrType {
	return &ErrType{fmt.Sprintf("can not read %s into %s", reflect.TypeOf(v), reflect.TypeOf(t))}
}

func errWrite(v, t any) *ErrType {
	return &ErrType{fmt.Sprintf("can not write %s to %s", reflect.TypeOf(v), reflect.TypeOf(t))}
}

func errIndex(v any) *ErrType {
	return &ErrType{fmt.Sprintf("can not index %s", reflect.TypeOf(v))}
}
