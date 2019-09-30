package goprint // import "vimagination.zapto.org/goprint"

import (
	"fmt"
	"io"
	"reflect"
)

type Type struct {
	v reflect.Value
}

func Wrap(v interface{}) *Type {
	return &Type{
		v: reflect.ValueOf(v),
	}
}

func (t *Type) Format(s fmt.State, v rune) {

}

type indentPrinter struct {
	io.Writer
}

var indent = []byte{'	'}

func (i *indentPrinter) Write(p []byte) (int, error) {
	var (
		total int
		last  int
	)
	for n, c := range p {
		if c == '\n' {
			m, err := i.Writer.Write(p[last : n+1])
			total += m
			if err != nil {
				return total, err
			}
			_, err = i.Writer.Write(indent)
			if err != nil {
				return total, err
			}
			last = n + 1
		}
	}
	if last != len(p) {
		m, err := i.Writer.Write(p[last:])
		total += m
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func (i *indentPrinter) Print(args ...interface{}) {
	fmt.Fprint(i, args...)
}

func (i *indentPrinter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(i, format, args...)
}

func (i *indentPrinter) WriteString(s string) (int, error) {
	return i.Write([]byte(s))
}
