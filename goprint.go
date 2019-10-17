package goprint // import "vimagination.zapto.org/goprint"

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Type struct {
	v     interface{}
	pkgFn func(io.Writer, reflect.Type)
}

func PkgName(w io.Writer, typ reflect.Type) {
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

func Wrap(v interface{}, pkgFn func(io.Writer, reflect.Type)) *Type {
	if pkgFn == nil {
		pkgFn = PkgName
	}
	return &Type{
		v:     v,
		pkgFn: pkgFn,
	}
}

func (t *Type) Format(s fmt.State, v rune) {
	t.format(reflect.ValueOf(t.v), s, s.Flag('+'), false)
}

var (
	ptr          = []byte{'&'}
	ptrp         = []byte{'*'}
	dot          = []byte{'.'}
	colon        = []byte{':'}
	comma        = []byte{','}
	bracketOpen  = []byte{'['}
	bracketClose = []byte{']'}
	braceOpen    = []byte{'{'}
	braceClose   = []byte{'}'}
	parenOpen    = []byte{'('}
	parenClose   = []byte{')'}
	newLine      = []byte{'\n'}
	indent       = []byte{'	'}
	nilt         = []byte{'n', 'i', 'l'}
	truet        = []byte{'t', 'r', 'u', 'e'}
	falset       = []byte{'f', 'a', 'l', 's', 'e'}
	mapt         = []byte{'m', 'a', 'p'}
	structt      = []byte{'s', 't', 'r', 'u', 'c', 't'}
	tagStart     = []byte{' ', '`'}
	tagEnd       = tagStart[1:]
	space        = tagStart[:1]
	interfacet   = []byte{'i', 'n', 't', 'e', 'r', 'f', 'a', 'c', 'e'}
	funct        = []byte{'f', 'u', 'n', 'c'}
	ellipsis     = []byte{'.', '.', '.'}
	complext     = []byte{'i'}
	complexa     = []byte{' ', '+', ' '}
)

func (t *Type) format(v reflect.Value, w io.Writer, verbose, inArray bool) {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			w.Write(nilt)
			return
		}
		w.Write(ptr)
		t.format(v.Elem(), w, verbose, false)
	case reflect.Struct:
		typ := v.Type()
		if !inArray {
			t.formatType(w, typ)
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
			t.format(vf, &ip, verbose, false)
			ip.Write(comma)
			any = true
		}
		if any {
			w.Write(newLine)
		}
		w.Write(braceClose)
	case reflect.Array:
		t.formatType(w, v.Type())
		w.Write(braceOpen)
		if l := v.Len(); l > 0 {
			ip := indentPrinter{w}
			for i := 0; i < l; i++ {
				ip.Write(newLine)
				t.format(v.Index(i), &ip, verbose, true)
				ip.Write(comma)
			}
			w.Write(newLine)
		}
		w.Write(braceClose)
	case reflect.Slice:
		if v.IsNil() {
			w.Write(nilt)
			return
		}
		t.formatType(w, v.Type())
		w.Write(braceOpen)
		if l := v.Len(); l > 0 {
			ip := indentPrinter{w}
			for i := 0; i < l; i++ {
				ip.Write(newLine)
				t.format(v.Index(i), &ip, verbose, true)
				ip.Write(comma)
			}
			w.Write(newLine)
		}
		w.Write(braceClose)
	case reflect.Map:
		if v.IsNil() {
			w.Write(nilt)
			return
		}
		t.formatType(w, v.Type())
		w.Write(braceOpen)
		keys := v.MapKeys()
		if len(keys) > 0 {
			ip := indentPrinter{w}
			for _, k := range v.MapKeys() {
				ip.Write(newLine)
				t.format(k, w, verbose, false)
				w.Write(colon)
				t.format(v.MapIndex(k), w, verbose, true)
				w.Write(comma)
			}
			w.Write(newLine)
		}
		w.Write(braceClose)
	case reflect.Interface:
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		io.WriteString(w, strconv.FormatUint(v.Uint(), 10))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		io.WriteString(w, strconv.FormatInt(v.Int(), 10))
	case reflect.Float32, reflect.Float64:
		io.WriteString(w, strconv.FormatFloat(v.Float(), 'f', -1, 64))
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		r := real(c)
		i := imag(c)
		if r != 0 {
			io.WriteString(w, strconv.FormatFloat(r, 'f', -1, 64))
			if i != 0 {
				w.Write(complexa)
			}
		}
		if i != 0 {
			io.WriteString(w, strconv.FormatFloat(i, 'f', -1, 64))
			w.Write(complext)
		}
	case reflect.UnsafePointer:
		io.WriteString(w, strconv.FormatUint(uint64(v.Pointer()), 10))
	case reflect.String:
		io.WriteString(w, strconv.Quote(v.String()))
	case reflect.Bool:
		if v.Bool() {
			w.Write(truet)
		} else {
			w.Write(falset)
		}
	case reflect.Func:
		w.Write(nilt)
	default:
	}
}

func (t *Type) formatType(w io.Writer, rt reflect.Type) {
	for {
		if n := rt.Name(); n != "" {
			t.pkgFn(w, rt)
			return
		}
		switch rt.Kind() {
		case reflect.Ptr:
			w.Write(ptrp)
		case reflect.Struct:
			w.Write(structt)
			w.Write(braceOpen)
			if l := rt.NumField(); l > 0 {
				ip := indentPrinter{w}
				for i := 0; i < l; i++ {
					ip.Write(newLine)
					f := rt.Field(i)
					if !f.Anonymous {
						io.WriteString(&ip, f.Name)
						ip.Write(space)
					}
					t.formatType(&ip, f.Type)
					if f.Tag != "" {
						ip.Write(tagStart)
						io.WriteString(&ip, string(f.Tag))
						ip.Write(tagEnd)
					}
				}
				w.Write(newLine)
			}
			w.Write(braceClose)
			return
		case reflect.Array:
			w.Write(bracketOpen)
			io.WriteString(w, strconv.FormatInt(int64(rt.Len()), 10))
			w.Write(bracketClose)
		case reflect.Slice:
			w.Write(bracketOpen)
			w.Write(bracketClose)
		case reflect.Map:
			w.Write(mapt)
			w.Write(bracketOpen)
			t.formatType(w, rt.Key())
			w.Write(bracketClose)
			t.formatType(w, rt.Elem())
			return
		case reflect.Interface:
			w.Write(interfacet)
			w.Write(braceOpen)
			if l := rt.NumMethod(); l > 0 {
				ip := indentPrinter{w}
				for i := 0; i < l; i++ {
					ip.Write(newLine)
					m := rt.Method(i)
					io.WriteString(&ip, m.Name)
					ip.Write(space)
					t.formatType(&ip, m.Type)
				}
				w.Write(newLine)
			}
			w.Write(braceClose)
			return
		case reflect.Func:
			w.Write(funct)
			w.Write(space)
			io.WriteString(w, rt.Name())
			w.Write(parenOpen)
			if in := rt.NumIn(); in > 0 {
				for i := 0; i < in; i++ {
					if i > 0 {
						w.Write(comma)
					}
					if i == in-1 && rt.IsVariadic() {
						w.Write(ellipsis)
						t.formatType(w, rt.In(i).Elem())
					} else {
						t.formatType(w, rt.In(i))
					}
				}
			}
			w.Write(parenClose)
			out := rt.NumOut()
			if out != 1 {
				w.Write(parenOpen)
			}
			for i := 0; i < out; i++ {
				if i > 0 {
					w.Write(comma)
				}
				t.formatType(w, rt.Out(i))
			}
			if out > 1 {
				w.Write(parenClose)
			}
			return
		}
		rt = rt.Elem()
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
