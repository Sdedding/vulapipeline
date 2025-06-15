package organize

import (
	"fmt"
)

type Color uint8

const (
	DefaultColor Color = iota
	RedColor
	GreenColor
	BlueColor
	YellowColor
)

type Weight uint8

const (
	NormalWeight Weight = iota
	BoldWeight
)

type StyledString struct {
	Text   string
	Color  Color
	Weight Weight
}

func Text(s string) StyledString {
	return StyledString{s, DefaultColor, NormalWeight}
}

func Red(s string) StyledString {
	return StyledString{s, RedColor, NormalWeight}
}

func Green(s string) StyledString {
	return StyledString{s, GreenColor, NormalWeight}
}

func Blue(s string) StyledString {
	return StyledString{s, BlueColor, NormalWeight}
}

func Yellow(s string) StyledString {
	return StyledString{s, YellowColor, NormalWeight}
}

func Bold(s string) StyledString {
	return StyledString{s, DefaultColor, BoldWeight}
}

func Colored(s string, c Color) StyledString {
	return StyledString{s, c, NormalWeight}
}

type StyledOutput []StyledString

func (o *StyledOutput) AddNewline() {
	*o = append(*o, StyledString{Text: "\n"})
}

func (o *StyledOutput) AddIndentation() {
	o.Add(Text("  ")) // Two spaces for indentation
}

func (o *StyledOutput) Add(s ...StyledString) {
	*o = append(*o, s...)
}

func (o *StyledOutput) AddKeyValue(key StyledString, value ...StyledString) {
	isEmpty := true
	for _, s := range value {
		if s.Text != "" {
			isEmpty = false
			break
		}
	}

	if isEmpty {
		return
	}

	if key.Text != "" {
		o.Add(StyledString{key.Text, key.Color, BoldWeight}, Text(": "))
		o.Add(value...)
	} else {
		o.Add(value...)
	}
}

func (o *StyledOutput) AppendStyledOutput(other *StyledOutput) {
	*o = append(*o, *other...)
}

func JoinStyledString(sep string, t ...string) StyledString {
	s := []byte{}
	for i, v := range t {
		if i != 0 {
			s = append(s, sep...)
		}
		s = append(s, v...)
	}
	return Text(string(s))
}

func JoinStyledStringer[T fmt.Stringer](sep string, t ...T) StyledString {
	s := []byte{}
	for i, v := range t {
		if i != 0 {
			s = append(s, sep...)
		}
		s = append(s, v.String()...)
	}
	return Text(string(s))
}

func JoinStyled(sep string, t ...StyledString) []StyledString {
	s := []StyledString{}
	for i, v := range t {
		if i != 0 {
			s = append(s, Text(sep))
		}
		s = append(s, v)
	}
	return s
}

func (o *StyledOutput) ToConsole() []byte {
	buffer := make([]byte, 0)
	for _, styledString := range *o {
		colorCode := ""
		switch styledString.Color {
		case RedColor:
			colorCode = "\033[31m"
		case GreenColor:
			colorCode = "\033[32m"
		case BlueColor:
			colorCode = "\033[34m"
		case YellowColor:
			colorCode = "\033[33m"
		default:
		}
		buffer = append(buffer, colorCode...)

		if styledString.Weight == BoldWeight {
			buffer = append(buffer, "\033[1m"...)
		}

		buffer = append(buffer, styledString.Text...)
		buffer = append(buffer, "\033[0m"...)
	}
	return buffer
}
