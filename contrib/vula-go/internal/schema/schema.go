package schema

import (
	"reflect"
)

type Reader interface {
	Read(t any) error
}

type Writer interface {
	Write(v any) error
}

type PathReader interface {
	ReadPath(path []string, t any) error
}

type PathWriter interface {
	WritePath(v any, path []string) error
}

type PathAdder interface {
	AddPath(v any, path []string) error
}

type PathRemover interface {
	RemovePath(v any, path []string) error
}

func Read(v any, path []string, t any) (err error) {
	if len(path) == 0 {
		return readValue(v, t)
	}
	if v, ok := v.(PathReader); ok {
		return v.ReadPath(path, t)
	}

	rv := reflect.ValueOf(v)
	switch {
	case isStruct(rv):
		err = structReadPath(rv, path, t)
	case isStructAddr(rv):
		err = structReadPath(rv.Elem(), path, t)
	default:
		err = errIndex(v)
	}
	return
}

func Write(v any, path []string, t any) (err error) {
	if len(path) == 0 {
		return writeValue(v, t)
	}
	if t, ok := t.(PathWriter); ok {
		return t.WritePath(v, path)
	}

	rt := reflect.ValueOf(t)
	switch {
	case isStructAddr(rt):
		err = structWritePath(v, path, rt)
	default:
		err = errIndex(t)
	}
	return
}

func Add(v any, path []string, t any) (err error) {
	if t, ok := t.(PathAdder); ok {
		return t.AddPath(v, path)
	}

	rt := reflect.ValueOf(t)
	switch {
	case isStructAddr(rt):
		err = structAddPath(v, path, rt)
	default:
		err = errIndex(t)
	}
	return
}

func Remove(v any, path []string, t any) (err error) {
	if t, ok := t.(PathRemover); ok {
		return t.RemovePath(v, path)
	}
	rt := reflect.ValueOf(t)
	switch {
	case isStructAddr(rt):
		err = structRemovePath(v, path, rt)
	default:
		err = errIndex(t)
	}
	return
}
