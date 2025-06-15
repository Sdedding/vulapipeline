package core

import "fmt"

type ErrParse struct {
	Err error
}

func (e *ErrParse) Error() string {
	return fmt.Sprintf("parse error: %v", e.Err)
}
