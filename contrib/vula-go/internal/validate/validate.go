package validate

import "fmt"

type V struct {
	Errors []KeyError
}

func (v *V) Err() error {
	if len(v.Errors) > 0 {
		return &Err{v.Errors}
	}
	return nil
}

type KeyError struct {
	Key     string
	Value   any
	Message string
	Err     error
}

func (e *KeyError) Error() string {
	return fmt.Sprintf("validate %s = '%v': %v", e.Key, e.Value, e.Err)
}

type Err struct {
	errors []KeyError
}

func (e *Err) Error() string {
	if len(e.errors) == 1 {
		return e.errors[0].Error()
	}
	return fmt.Sprintf("%s, ...", e.errors[0].Error())
}

func (e *Err) Errors() []KeyError {
	return e.errors
}
