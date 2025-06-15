package schema

import (
	"reflect"
	"strings"
	"sync"
	"unicode"

	"gopkg.in/yaml.v3"
)

func readStruct(rv reflect.Value, t any) (err error) {
	switch t := t.(type) {
	case *any:
		*t = rv.Interface()
	case *string:
		var b []byte
		b, err = yaml.Marshal(rv.Interface())
		if err == nil {
			*t = string(b)
		}
	case *[]byte:
		*t, err = yaml.Marshal(rv.Interface())
	case *[]StringMapItem:
		err = readStructIntoStringMapping(rv, t)
	default:
		rt := reflect.ValueOf(t)
		if rt.Kind() == reflect.Pointer && rt.Elem().Type() == rv.Type() {
			rt.Elem().Set(rv)
		} else {
			err = errRead(rv.Interface(), t)
		}
	}
	return
}

func writeStruct(v any, addr reflect.Value) (err error) {
	switch v := v.(type) {
	case string:
		err = yaml.Unmarshal([]byte(v), addr.Interface())
	case []byte:
		err = yaml.Unmarshal(v, addr.Interface())
	case []StringMapItem:
		err = writeStructFromStringMapping(v, addr)
	default:
		rt := reflect.ValueOf(v)
		if rt.Type() == addr.Type() {
			addr.Elem().Set(rt.Elem())
		} else {
			err = errWrite(v, addr.Elem().Interface())
		}
	}
	return
}

func getYamlStructField(rv reflect.Value, yamlFieldName string) (field yamlStructField, err error) {

	fields := getYamlStructFieldsCached(rv.Type())
	field, ok := fields[yamlFieldName]
	if !ok {
		err = &ErrKey{Key: yamlFieldName}
	}
	return
}

func structReadPath(rv reflect.Value, path []string, t any) (err error) {
	if len(path) == 0 {
		return readStruct(rv, t)
	}

	field, err := getYamlStructField(rv, path[0])
	if err != nil {
		return
	}

	fieldValue := rv.Field(field.index).Interface()
	if len(path) == 1 {
		return readValue(fieldValue, t)
	}

	return Read(fieldValue, path[1:], t)
}

func structWritePath(v any, path []string, addr reflect.Value) (err error) {
	if len(path) == 0 {
		return writeStruct(v, addr)
	}

	field, err := getYamlStructField(addr.Elem(), path[0])
	if err != nil {
		return
	}

	fieldAddr := addr.Elem().Field(field.index).Addr().Interface()
	return Write(v, path[1:], fieldAddr)
}

func structAddPath(v any, path []string, addr reflect.Value) error {
	if len(path) == 0 {
		m, err := convertToMap[string, any](v)
		if err != nil {
			return err
		}
		for key, value := range m {
			field, err := getYamlStructField(addr.Elem(), key)
			if err != nil {
				return err
			}
			fieldAddr := addr.Elem().Field(field.index).Addr().Interface()
			err = readValue(value, fieldAddr)
			if err != nil {
				return err
			}
		}
		return nil
	}

	field, err := getYamlStructField(addr.Elem(), path[0])
	if err != nil {
		return err
	}
	fieldAddr := addr.Elem().Field(field.index).Addr().Interface()
	return Add(v, path[1:], fieldAddr)
}

func structRemovePath(v any, path []string, addr reflect.Value) error {
	if len(path) == 0 {
		fieldName := ""
		err := readValue(v, &fieldName)
		if err != nil {
			return err
		}
		field, err := getYamlStructField(addr.Elem(), fieldName)
		if err != nil {
			return err
		}
		addr.Field(field.index).SetZero()
		return nil
	}

	field, err := getYamlStructField(addr.Elem(), path[0])
	if err != nil {
		return err
	}
	fieldAddr := addr.Elem().Field(field.index).Addr().Interface()
	return Remove(v, path[1:], fieldAddr)
}

func isStruct(rv reflect.Value) bool {
	return rv.Kind() == reflect.Struct
}

func isStructAddr(rv reflect.Value) bool {
	return rv.Kind() == reflect.Pointer && rv.Elem().Kind() == reflect.Struct
}

func convertToMap[K comparable, V any](v any) (m map[K]V, err error) {
	if v, ok := v.(map[K]V); ok {
		return v, nil
	}

	var (
		mapKey   K
		mapValue V
	)
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Map {
		m = make(map[K]V)
		iter := rv.MapRange()
		for iter.Next() {
			err = writeValue(iter.Key().Interface(), &mapKey)
			if err != nil {
				return
			}
			err = writeValue(iter.Value().Interface(), &mapValue)
			if err != nil {
				return
			}
			m[mapKey] = mapValue
		}
	} else {
		err = writeValue(v, &mapKey)
		if err != nil {
			return
		}
		err = writeValue(true, &mapValue)
		if err != nil {
			return
		}
		m = map[K]V{mapKey: mapValue}
	}
	return
}

func readStructIntoStringMapping(rv reflect.Value, t *[]StringMapItem) error {
	fields := getYamlStructFieldsCached(rv.Type())
	*t = make([]StringMapItem, 0, len(fields))
	value := ""
	for key, field := range fields {
		fieldValue := rv.Field(field.index)
		if field.omitEmpty && fieldValue.IsZero() {
			continue
		}
		err := readValue(fieldValue.Interface(), &value)
		if err != nil {
			return err
		}
		*t = append(*t, StringMapItem{key, value})
	}
	return nil
}

func writeStructFromStringMapping(v []StringMapItem, addr reflect.Value) error {
	fields := getYamlStructFieldsCached(addr.Elem().Type())
	for _, item := range v {
		fields, ok := fields[item.Key]
		if !ok {
			return &ErrKey{Key: item.Key}
		}
		fieldAddr := addr.Elem().Field(fields.index).Addr()
		err := writeValue(item.Value, fieldAddr.Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	yamlStructFieldCacheMux sync.Mutex
	yamlStructFieldCache    = map[reflect.Type]map[string]yamlStructField{}
)

type yamlStructField struct {
	index     int
	omitEmpty bool
}

// getYamlStructFieldsCached provides a cache around getYamlStructFields
func getYamlStructFieldsCached(t reflect.Type) map[string]yamlStructField {
	yamlStructFieldCacheMux.Lock()
	defer yamlStructFieldCacheMux.Unlock()

	fields := yamlStructFieldCache[t]
	if fields == nil {
		fields = getYamlStructFields(t)
		yamlStructFieldCache[t] = fields
	}
	return fields
}

// getYamlFieldIndices returns a map from the yaml field names to the struct field indices
// t must be a struct type
func getYamlStructFields(t reflect.Type) map[string]yamlStructField {
	numField := t.NumField()
	fieldIndices := make(map[string]yamlStructField, numField)
	for i := 0; i < numField; i++ {
		name, omitEmpty := parseYamlStructTag(t.Field(i))
		fieldIndices[name] = yamlStructField{i, omitEmpty}
	}

	return fieldIndices
}

// getNameForYamlTag converts a yaml field tag to the name that the yaml marshaler would
func parseYamlStructTag(field reflect.StructField) (name string, omitEmpty bool) {
	tag, hasTag := field.Tag.Lookup("yaml")
	if hasTag {
		commaIndex := strings.IndexByte(tag, ',')
		if commaIndex > 0 {
			name = tag[:commaIndex]
			omitEmpty = strings.Contains(tag[commaIndex+1:], "omitempty") // TODO: better parsing
		} else {
			name = tag
		}
	} else {
		name = firstToLower(field.Name)
	}
	return
}

func firstToLower(s string) string {
	if len(s) == 0 {
		return s
	}
	chars := []rune(s)
	chars[0] = unicode.ToLower(chars[0])
	return string(chars)
}
