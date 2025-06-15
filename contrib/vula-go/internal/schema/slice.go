package schema

import (
	"reflect"
	"slices"
)

type Slice[T comparable] []T

var (
	_ Reader      = Slice[any]{}
	_ Writer      = &Slice[any]{}
	_ PathReader  = Slice[any]{}
	_ PathWriter  = &Slice[any]{}
	_ PathAdder   = &Slice[any]{}
	_ PathRemover = &Slice[any]{}
)

func (v Slice[T]) Read(t any) error {
	return ReadSlice(v, t)
}

func (t *Slice[T]) Write(v any) error {
	return WriteSlice(v, t)
}

func (v Slice[T]) ReadPath(path []string, t any) error {
	return SliceReadPath(v, path, t)
}

func (t *Slice[T]) WritePath(v any, path []string) error {
	return SliceWritePath(v, path, t)
}

func (t *Slice[T]) AddPath(v any, path []string) error {
	return SliceAddPath(v, path, t)
}

func (t *Slice[T]) RemovePath(v any, path []string) error {
	return SliceRemovePath(v, path, t)
}

func ReadSlice[S ~[]T, T any](v S, t any) (err error) {
	switch t := t.(type) {
	case *any:
		*t = v
	case *[]T:
		*t = []T(v)
	default:
		rt := reflect.ValueOf(t)
		switch {
		case rt.Kind() == reflect.Pointer && rt.Elem().Kind() == reflect.Slice:
			rt.Elem().Set(reflect.MakeSlice(rt.Elem().Type(), 0, len(v)))
			item := reflect.New(rt.Elem().Type().Elem())
			for i := range v {
				err = readValue(v[i], item.Interface())
				if err != nil {
					return
				}
				rt.Elem().Set(reflect.Append(rt.Elem(), item.Elem()))
			}
		default:
			err = errRead(v, t)
		}
	}
	return
}

func WriteSlice[S ~[]T, T any](v any, t *S) (err error) {
	*t = nil

	switch v := v.(type) {
	case []T:
		*t = S(v)
	default:
		rv := reflect.ValueOf(v)
		switch {
		case rv.Kind() == reflect.Slice:
			length := rv.Len()
			var item T
			for i := 0; i < length; i++ {
				err = writeValue(rv.Index(i).Interface(), &item)
				if err != nil {
					return
				}
				*t = append(*t, item)
			}
		default:
			err = errWrite(v, *t)
		}
	}
	return nil
}

func SliceReadPath[S ~[]T, T any](v S, path []string, t any) error {
	if len(path) > 0 {
		return errIndex(v)
	}
	return readValue(v, t)
}

func SliceWritePath[S ~[]T, T any](v any, path []string, s *S) error {
	if len(path) > 0 {
		return errIndex(*s)
	}
	return writeValue(v, s)
}

func SliceAddPath[S ~[]T, T comparable](v any, path []string, s *S) error {
	if len(path) > 0 {
		return errIndex(*s)
	}

	var item T
	err := writeValue(v, &item)
	if err != nil {
		return err
	}
	if !slices.Contains(*s, item) {
		*s = append(*s, item)
	}
	return nil
}

func SliceRemovePath[S ~[]T, T comparable](v any, path []string, s *S) error {
	if len(path) > 0 {
		return errIndex(*s)
	}

	var item T
	err := writeValue(v, &item)
	if err != nil {
		return err
	}
	index := slices.Index(*s, item)
	if index >= 0 {
		*s = slices.Delete(*s, index, index+1)
	}
	return nil
}
