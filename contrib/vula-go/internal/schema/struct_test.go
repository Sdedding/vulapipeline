package schema

import (
	"testing"

	"gopkg.in/yaml.v3"
)

type testYamlStruct struct {
	Field1 string
	Field2 int64 `yaml:"field2"`
	Field3 bool  `yaml:"field3,omitempty"`
}

func newTestYamlStruct() testYamlStruct {
	return testYamlStruct{"1", 2, true}
}

func TestReadPathStruct(t *testing.T) {
	// arrange
	s := newTestYamlStruct()

	// act
	field1 := ""
	err := Read(&s, []string{"field1"}, &field1)
	if err != nil {
		t.Error(err)
	}
	field2 := int64(0)
	err = Read(s, []string{"field2"}, &field2)
	if err != nil {
		t.Error(err)
	}
	field3 := false
	err = Read(s, []string{"field3"}, &field3)
	if err != nil {
		t.Error(err)
	}

	// assert
	if field1 != "1" {
		t.Errorf("field1 = '%s'", field1)
	}

	if field2 != int64(2) {
		t.Errorf("field2 = %d", field2)
	}

	if field3 != true {
		t.Errorf("field3 = %v", field3)
	}
}

func TestYamlStructTags(t *testing.T) {
	// arrange
	content := `
field1: one
field2: 2
field3: true
`
	var s testYamlStruct

	// act
	err := yaml.Unmarshal([]byte(content), &s)
	if err != nil {
		t.Error(err)
	}

	// assert
	if s.Field1 != "one" {
		t.Errorf("s.Field1 = '%s'", s.Field1)
	}

	if s.Field2 != 2 {
		t.Errorf("s.Field2 = %d", s.Field2)
	}

	if s.Field3 != true {
		t.Errorf("s.Field3 = %v", s.Field3)
	}
}
