package goprint // import "vimagination.zapto.org/goprint"

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Type struct {
	v interface{}
}

func Wrap(v interface{}) *Type {
	return &Type{
		v: v,
	}
}

func (t *Type) Format(s fmt.State, v rune) {
	format(reflect.ValueOf(t.v), s, s.Flag('+'), false)
}

var (
	ptr          = []byte{'&'}
	dot          = []byte{'.'}
	colon        = []byte{':'}
	comma        = []byte{','}
	bracketOpen  = []byte{'['}
	bracketClose = []byte{']'}
	braceOpen    = []byte{'{'}
	braceClose   = []byte{'}'}
	newLine      = []byte{'\n'}
	indent       = []byte{'	'}
	nilt         = []byte{'n', 'i', 'l'}
	truet        = []byte{'t', 'r', 'u', 'e'}
	falset       = []byte{'f', 'a', 'l', 's', 'e'}
)

func printName(w io.Writer, typ reflect.Type) {
	if pkg := typ.PkgPath(); pkg != "" {
		var pos int
		if p := strings.LastIndexByte(pkg, '/'); p >= 0 {
			pos = p + 1
		}
		io.WriteString(w, pkg[pos:])
		w.Write(dot)
	}
	io.WriteString(w, typ.Name())
}

func format(v reflect.Value, w io.Writer, verbose, inArray bool) {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			w.Write(nilt)
			return
		}
		w.Write(ptr)
		format(v.Elem(), w, verbose, false)
	case reflect.Struct:
		typ := v.Type()
		if !inArray {
			printName(w, typ)
		}
		w.Write(braceOpen)
		ip := indentPrinter{w}
		var any bool
		for i := 0; i < typ.NumField(); i++ {
			vf := v.Field(i)
			if vf.IsZero() && !verbose {
				continue
			}
			tf := typ.Field(i)
			ip.Write(newLine)
			ip.WriteString(tf.Name)
			ip.Write(colon)
			format(vf, &ip, verbose, false)
			ip.Write(comma)
			any = true
		}
		if any {
			w.Write(newLine)
		}
		w.Write(braceClose)
	case reflect.Array:
	case reflect.Slice:
		if v.IsNil() {
			w.Write(nilt)
			return
		}
		typ := v.Type()
		n := typ.Name()
		if n != "" {
			io.WriteString(w, typ.Name())
		} else {
			w.Write(bracketOpen)
			w.Write(bracketClose)
			printName(w, typ.Elem())
		}
		w.Write(braceOpen)
		if l := v.Len(); l > 0 {
			ip := indentPrinter{w}
			for i := 0; i < l; i++ {
				ip.Write(newLine)
				format(v.Index(i), &ip, verbose, true)
				ip.Write(comma)
			}
			w.Write(newLine)
		}
		w.Write(braceClose)
	case reflect.Map:
	case reflect.Interface:
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		io.WriteString(w, strconv.FormatUint(v.Uint(), 10))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		io.WriteString(w, strconv.FormatInt(v.Int(), 10))
	case reflect.String:
		io.WriteString(w, strconv.Quote(v.String()))
	case reflect.Bool:
		if v.Bool() {
			w.Write(truet)
		} else {
			w.Write(falset)
		}
	default:
	}
}

type indentPrinter struct {
	io.Writer
}

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
