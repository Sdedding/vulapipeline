package schema

import (
	"testing"
)

func TestReadMap(t *testing.T) {
	// arrange
	m := map[string]int64{"one": 1, "two": 2}
	mt := map[string]string{}

	// act
	err := ReadMap(m, &mt)
	if err != nil {
		t.Error(err)
	}

	// assert
	if v := mt["one"]; v != "1" {
		t.Errorf(`mt["one"] = "%s"`, v)
	}
	if v := mt["two"]; v != "2" {
		t.Errorf(`mt["two"] = "%s"`, v)
	}
}
