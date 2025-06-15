package schema

import (
	"reflect"

	"gopkg.in/yaml.v3"
)

type Map[K comparable, V any] map[K]V

var (
	_ Reader      = Map[any, any]{}
	_ Writer      = &Map[any, any]{}
	_ PathReader  = Map[any, any]{}
	_ PathWriter  = &Map[any, any]{}
	_ PathAdder   = &Map[any, any]{}
	_ PathRemover = &Map[any, any]{}
)

func (m Map[K, V]) Read(t any) error {
	return ReadMap(m, t)
}

func (m *Map[K, V]) Write(t any) error {
	return WriteMap(t, m)
}

func (m Map[K, V]) ReadPath(path []string, t any) error {
	return MapReadPath(m, path, t)
}

func (m *Map[K, V]) WritePath(v any, path []string) error {
	return MapWritePath(v, path, m)
}

func (m *Map[K, V]) AddPath(v any, path []string) error {
	return MapAddPath(v, path, m)
}

func (m *Map[K, V]) RemovePath(v any, path []string) error {
	return MapRemovePath(v, path, m)
}

func ReadMap[M ~map[K]V, K comparable, V any](v M, t any) (err error) {
	switch t := t.(type) {
	case *any:
		*t = v
	case *string:
		var b []byte
		b, err = yaml.Marshal(v)
		*t = string(b)
	case *[]byte:
		*t, err = yaml.Marshal(v)
	case *[]StringMapItem:
		*t = make([]StringMapItem, 0, len(v))
		var item StringMapItem
		for key, value := range v {
			err = readValue(key, &item.Key)
			if err != nil {
				return err
			}
			err = readValue(value, &item.Value)
			if err != nil {
				return err
			}
			*t = append(*t, item)
		}
	default:
		rt := reflect.ValueOf(t)
		if rt.Kind() == reflect.Pointer && rt.Elem().Kind() == reflect.Map {
			mapKey := reflect.New(rt.Elem().Type().Key())
			mapValue := reflect.New(rt.Elem().Type().Elem())
			if rt.Elem().IsNil() {
				rt.Elem().Set(reflect.MakeMap(rt.Elem().Type()))
			}
			for k, v := range v {
				err = readValue(k, mapKey.Interface())
				if err != nil {
					return err
				}
				err = readValue(v, mapValue.Interface())
				if err != nil {
					return
				}
				rt.Elem().SetMapIndex(mapKey.Elem(), mapValue.Elem())
			}
		} else {
			err = errRead(v, t)
		}
	}
	return
}

func WriteMap[M ~map[K]V, K comparable, V any](v any, t *M) (err error) {
	switch v := v.(type) {
	case map[K]V:
		*t = v
	case string:
		err = yaml.Unmarshal([]byte(v), t)
	case []byte:
		err = yaml.Unmarshal(v, t)
	case []StringMapItem:
		var (
			mapKey   K
			mapValue V
		)
		m := make(map[K]V, len(v))
		for _, item := range v {
			err = writeValue(item.Key, &mapKey)
			if err != nil {
				return
			}
			err = writeValue(item.Value, &mapValue)
			if err != nil {
				return
			}
			m[mapKey] = mapValue
		}
		*t = m
	default:
		m := make(map[K]V)
		rv := reflect.Indirect(reflect.ValueOf(v))
		if rv.Kind() == reflect.Map {
			var (
				mapKey   K
				mapValue V
				iter     = rv.MapRange()
			)
			for iter.Next() {
				err = readValue(iter.Key().Interface(), &mapKey)
				if err != nil {
					return
				}
				err = readValue(iter.Value().Interface(), &mapValue)
				if err != nil {
					return
				}
				m[mapKey] = mapValue
			}
			*t = m
		} else {
			err = errWrite(v, *t)
		}

	}
	return
}

func MapReadPath[M ~map[K]V, K comparable, V any](m M, path []string, t any) (err error) {
	if len(path) == 0 {
		return ReadMap[M](m, t)
	}

	var mapKey K
	err = readValue(path[0], &mapKey)
	if err != nil {
		return
	}
	mapValue, ok := m[mapKey]
	if !ok {
		return &ErrKey{Key: path[0]}
	}
	return Read(mapValue, path[1:], t)
}

func MapWritePath[M ~map[K]V, K comparable, V any](v any, path []string, m *M) (err error) {
	if len(path) == 0 {
		return readValue(v, m)
	}

	var mapKey K
	err = writeValue(path[0], &mapKey)
	if err != nil {
		return err
	}

	mapValue := (*m)[mapKey]
	err = Write(v, path[1:], &mapValue)
	(*m)[mapKey] = mapValue
	return
}

func MapAddPath[M ~map[K]V, K comparable, V any](v any, path []string, m *M) (err error) {
	if len(path) > 0 {
		return applyToMapItem(v, path, m, Add)
	}
	values, err := convertToMap[K, V](v)

	if *m == nil {
		*m = make(map[K]V)
	}

	for k, v := range values {
		(*m)[k] = v
	}
	return
}

func MapRemovePath[M ~map[K]V, K comparable, V any](v any, path []string, m *M) (err error) {
	if len(path) > 0 {
		return applyToMapItem(v, path, m, Remove)
	}

	var mapKey K
	err = readValue(v, &mapKey)
	if err != nil {
		return
	}
	delete(*m, mapKey)
	return
}

func applyToMapItem[M ~map[K]V, K comparable, V any](v any, path []string, m *M, f func(v any, path []string, t any) error) (err error) {
	var mapKey K
	err = readValue(path[0], &mapKey)
	if err != nil {
		return
	}
	mapValue, ok := (*m)[mapKey]
	if !ok {
		return &ErrKey{Key: path[0]}
	}
	err = f(v, path[1:], &mapValue)
	(*m)[mapKey] = mapValue
	return
}
