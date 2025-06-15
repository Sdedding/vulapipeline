package schema

import (
	"bytes"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type base64Test struct {
	Field Base64 `yaml:"field"`
}

func TestBase64Unmarshal(t *testing.T) {
	// arrange
	content := `field: AAAA`
	s := base64Test{}

	// act
	err := yaml.Unmarshal([]byte(content), &s)
	if err != nil {
		t.Error(err)
	}

	// assert
	if !bytes.Equal(make([]byte, 3), s.Field) {
		t.Errorf("s.Field = %v", s.Field)
	}
}

func TestBase64Marshal(t *testing.T) {
	// arrange
	s := base64Test{make(Base64, 3)}

	// act
	data, err := yaml.Marshal(s)
	if err != nil {
		t.Error(err)
	}

	// assert
	stringData := strings.TrimSpace(string(data))
	if stringData != "field: AAAA" {
		t.Errorf("stringData = %s", stringData)
	}
}

func TestBase64Read(t *testing.T) {
	// arrange
	s := base64Test{make(Base64, 3)}
	target := ""

	// act
	err := Read(s, []string{"field"}, &target)
	if err != nil {
		t.Error(err)
	}

	// assert
	if target != "AAAA" {
		t.Errorf("target = %s", target)
	}
}
