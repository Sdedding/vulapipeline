package organize

import "fmt"

type ErrPeerNotFound struct {
	Key  string
	Name string
}

func (e *ErrPeerNotFound) Error() string {
	return fmt.Sprintf("vula: peer with %s = '%s' not found", e.Key, e.Name)
}

type CompoundError []error

func (c *CompoundError) Error() string {
	switch len(*c) {
	case 0:
		return ""
	case 1:
		return (*c)[0].Error()
	default:
		return fmt.Sprintf("%v", (*c)[0])
	}
}

type compoundErrorBuilder struct {
	errors CompoundError
}

func (b *compoundErrorBuilder) Add(err error) {
	if err != nil {
		b.errors = append(b.errors, err)
	}
}

func (b *compoundErrorBuilder) Build() error {
	switch len(b.errors) {
	case 0:
		return nil
	case 1:
		return b.errors[0]
	default:
		return &b.errors
	}
}
