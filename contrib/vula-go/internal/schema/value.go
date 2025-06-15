package schema

import (
	"reflect"
)

func readValue(v, t any) (err error) {
	switch v := v.(type) {
	case Reader:
		err = v.Read(t)
	case *bool:
		err = readBool(*v, t)
	case bool:
		err = readBool(v, t)
	case *int64:
		err = readInt64(*v, t)
	case int64:
		err = readInt64(v, t)
	case *string:
		err = readString(*v, t)
	case string:
		err = readString(v, t)
	default:
		rv := reflect.ValueOf(v)
		switch {
		case isStructAddr(rv):
			err = readStruct(rv.Elem(), t)
		case isStruct(rv):
			err = readStruct(rv, t)
		default:
			err = errRead(v, t)
		}
	}
	return
}

func writeValue(v, t any) (err error) {
	switch t := t.(type) {
	case Writer:
		err = t.Write(v)
	case *bool:
		err = writeBool(v, t)
	case *int64:
		err = writeInt64(v, t)
	case *string:
		err = writeString(v, t)
	default:
		rt := reflect.ValueOf(t)
		switch {
		case isStructAddr(rt):
			err = writeStruct(v, rt)
		default:
			err = errWrite(v, t)
		}
	}
	return
}
