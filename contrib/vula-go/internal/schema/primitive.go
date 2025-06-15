package schema

import (
	"strconv"
)

func readBool(v bool, t any) (err error) {
	switch t := t.(type) {
	case *any:
		*t = v
	case *bool:
		*t = v
	case *string:
		*t = strconv.FormatInt(convertBoolToInt64(v), 10)
	default:
		err = errRead(t, v)
	}
	return
}

func writeBool(v any, t *bool) (err error) {
	switch v := v.(type) {
	case bool:
		*t = v
	case string:
		*t, err = strconv.ParseBool(v)
	default:
		err = errWrite(v, *t)
	}
	return
}

func readInt64(v int64, t any) (err error) {
	switch t := t.(type) {
	case *any:
		*t = v
	case *int64:
		*t = v
	case *string:
		*t = strconv.FormatInt(v, 10)
	default:
		err = errRead(v, t)
	}
	return
}

func writeInt64(v any, t *int64) (err error) {
	switch v := v.(type) {
	case int64:
		*t = v
	case string:
		*t, err = strconv.ParseInt(v, 10, 64)
	default:
		err = errWrite(v, *t)
	}
	return
}

func readString(v string, t any) (err error) {
	switch t := t.(type) {
	case *any:
		*t = v
	case *string:
		*t = v
	default:
		err = errRead(v, t)
	}
	return
}

func writeString(v any, t *string) (err error) {
	switch v := v.(type) {
	case string:
		*t = v
	default:
		err = errWrite(v, *t)
	}
	return
}

func convertBoolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}
