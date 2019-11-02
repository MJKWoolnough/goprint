// Package goprint allows for the printing of Go Values in Go Code
package goprint // import "vimagination.zapto.org/goprint"

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"vimagination.zapto.org/rwcount"
)

// Printer represents the options to print out a value
type Printer struct {
	pkgFn         func(io.Writer, reflect.Type)
	structFilter  func(reflect.Type, string) bool
	structReplace func(io.Writer, reflect.Type, string, reflect.Value) bool
	arrayReplace  func(io.Writer, reflect.Type, int, reflect.Value) bool
}

// New creates a new Printer from the given options
func New(opts ...Opt) *Printer {
	p := &Printer{
		pkgFn:         pkgName,
		structFilter:  noFilter,
		structReplace: noStructReplace,
		arrayReplace:  noArrayReplace,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Format prints out the value to the writer
func (p *Printer) Format(w io.Writer, val interface{}) (int64, error) {
	rw := rwcount.Writer{Writer: w}
	p.format(reflect.ValueOf(val), &rw, false, false)
	return rw.Count, rw.Err
}

// FormatVerbose prints out the value to the writer adding more detail that
// doesn't change the created value
func (p *Printer) FormatVerbose(w io.Writer, val interface{}) (int64, error) {
	rw := rwcount.Writer{Writer: w}
	p.format(reflect.ValueOf(val), &rw, true, false)
	return rw.Count, rw.Err
}

// Value in a wrapped value with its print configuration
type Value struct {
	v interface{}
	Printer
}

// Opt is a printing option
type Opt func(*Printer)

// PkgName sets the function that will write the package name for the type
// given.
//
// This should be used to override the default which takes the Base of the URL
// as the import name.
func PkgName(pf func(io.Writer, reflect.Type)) Opt {
	return func(p *Printer) {
		p.pkgFn = pf
	}
}

// StructFilter allows for only some fields of a Struct to be printed.
//
// The func recieves the struct type and the field name, and should return true
// to print the field, and false to not print it
func StructFilter(sf func(reflect.Type, string) bool) Opt {
	return func(p *Printer) {
		p.structFilter = sf
	}
}

// StructReplacer allows for some values in struct fields to be replaced.
//
// This is intended to aid the replacing of constants with named constants or
// vars, though other applications are applicable
//
// Whatever is written to the Writer will display in place of the constant.
//
// The return value should be set to true is the func has written anything, or
// false otherwise
func StructReplacer(rf func(io.Writer, reflect.Type, string, reflect.Value) bool) Opt {
	return func(p *Printer) {
		p.structReplace = rf
	}
}

// ArrayReplacer allows for some values in an array or slice to be replaced.
//
// This is intended to aid the replacing of constants with named constants or
// vars, though other applications are applicable
//
// Whatever is written to the Writer will display in place of the constant.
//
// The return value should be set to true is the func has written anything, or
// false otherwise
func ArrayReplacer(rf func(io.Writer, reflect.Type, int, reflect.Value) bool) Opt {
	return func(p *Printer) {
		p.arrayReplace = rf
	}
}

// FromPrinter copies the configuration from an existing Printer
func FromPrinter(p Printer) Opt {
	return func(q *Printer) {
		*q = p
	}
}

func pkgName(w io.Writer, typ reflect.Type) {
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

func noFilter(reflect.Type, string) bool { return true }

func noStructReplace(io.Writer, reflect.Type, string, reflect.Value) bool { return false }
func noArrayReplace(io.Writer, reflect.Type, int, reflect.Value) bool     { return false }

// Wrap creates a Printer with an embedded value to be used with fmt-like
// functions
func Wrap(v interface{}, opts ...Opt) *Value {
	p := Printer{
		pkgFn:         pkgName,
		structFilter:  noFilter,
		structReplace: noStructReplace,
		arrayReplace:  noArrayReplace,
	}
	for _, o := range opts {
		o(&p)
	}
	return &Value{
		v:       v,
		Printer: p,
	}
}

// Format implements the fmt.Formatter interface
func (t *Value) Format(s fmt.State, v rune) {
	t.format(reflect.ValueOf(t.v), s, s.Flag('+'), false)
}

// WriteTo implements the io.WriterTo interface
func (t *Value) WriteTo(w io.Writer) (int64, error) {
	rw := rwcount.Writer{
		Writer: w,
	}
	t.format(reflect.ValueOf(t.v), &rw, true, false)
	return rw.Count, rw.Err
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
	maket        = []byte{'m', 'a', 'k', 'e'}
	chant        = []byte{'c', 'h', 'a', 'n', ' '}
	appendt      = []byte{'a', 'p', 'p', 'e', 'n', 'd'}
	zero         = []byte{'0'}
)

func (t *Printer) format(v reflect.Value, w io.Writer, verbose, inArray bool) {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			w.Write(nilt)
			return
		}
		w.Write(ptr)
		c := v.Elem()
		switch c.Kind() {
		case reflect.Struct:
			t.format(c, w, verbose, false)
		default:
			w.Write(bracketOpen)
			w.Write(bracketClose)
			t.formatType(w, c.Type(), false)
			w.Write(braceOpen)
			t.format(c, w, verbose, true)
			w.Write(braceClose)
			w.Write(bracketOpen)
			w.Write(zero)
			w.Write(bracketClose)
		}
	case reflect.Struct:
		typ := v.Type()
		if !inArray {
			t.formatType(w, typ, false)
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
			if !t.structFilter(typ, tf.Name) {
				continue
			}
			ip.Write(newLine)
			ip.WriteString(tf.Name)
			ip.Write(colon)
			ip.Write(space)
			if !t.structReplace(w, typ, tf.Name, vf) {
				t.format(vf, &ip, verbose, false)
			}
			ip.Write(comma)
			any = true
		}
		if any {
			w.Write(newLine)
		}
		w.Write(braceClose)
	case reflect.Array:
		typ := v.Type()
		t.formatType(w, typ, false)
		w.Write(braceOpen)
		if l := v.Len(); l > 0 {
			ip := indentPrinter{w}
			for i := 0; i < l; i++ {
				ip.Write(newLine)
				e := v.Index(i)
				if !t.arrayReplace(&ip, typ, i, e) {
					t.format(v.Index(i), &ip, verbose, true)
				}
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
		typ := v.Type()
		fullCap := true
		l, c := v.Len(), v.Cap()
		if l < c && verbose {
			fullCap = false
			w.Write(appendt)
			w.Write(parenOpen)
			w.Write(maket)
			w.Write(parenOpen)
		}
		t.formatType(w, typ, false)
		if fullCap {
			w.Write(braceOpen)
		} else {
			w.Write(comma)
			w.Write(space)
			w.Write(zero)
			w.Write(comma)
			w.Write(space)
			io.WriteString(w, strconv.FormatInt(int64(c), 10))
			w.Write(parenClose)
			w.Write(comma)
		}
		if l > 0 {
			ip := indentPrinter{w}
			for i := 0; i < l; i++ {
				ip.Write(newLine)
				e := v.Index(i)
				if !t.arrayReplace(&ip, typ, i, e) {
					t.format(v.Index(i), &ip, verbose, true)
				}
				ip.Write(comma)
			}
			w.Write(newLine)
		}
		if fullCap {
			w.Write(braceClose)
		} else {
			w.Write(parenClose)
		}
	case reflect.Map:
		if v.IsNil() {
			w.Write(nilt)
			return
		}
		t.formatType(w, v.Type(), false)
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
		typ := v.Elem().Type()
		switch k := typ.Kind(); k {
		case reflect.Struct, reflect.Slice:
			t.format(v.Elem(), w, false, false)
		case reflect.Ptr:
			if !v.Elem().IsNil() && typ.Name() == "" {
				t.format(v.Elem(), w, false, false)
				break
			}
			fallthrough
		default:
			t.formatType(w, typ, false)
			w.Write(parenOpen)
			t.format(v.Elem(), w, false, false)
			w.Write(parenClose)
		}
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
	case reflect.Chan:
		if v.IsNil() {
			w.Write(nilt)
			return
		}
		w.Write(maket)
		w.Write(parenOpen)
		t.formatType(w, v.Type(), false)
		if l := v.Cap(); l > 0 {
			w.Write(comma)
			w.Write(space)
			io.WriteString(w, strconv.FormatInt(int64(l), 10))
		}
		w.Write(parenClose)
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

func (t *Printer) formatType(w io.Writer, rt reflect.Type, inInterface bool) {
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
			if l := rt.NumField(); l > 0 {
				w.Write(space)
				w.Write(braceOpen)
				ip := indentPrinter{w}
				for i := 0; i < l; i++ {
					ip.Write(newLine)
					f := rt.Field(i)
					if !f.Anonymous {
						io.WriteString(&ip, f.Name)
						ip.Write(space)
					}
					t.formatType(&ip, f.Type, false)
					if f.Tag != "" {
						ip.Write(tagStart)
						io.WriteString(&ip, string(f.Tag))
						ip.Write(tagEnd)
					}
				}
				w.Write(newLine)
			} else {
				w.Write(braceOpen)
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
			t.formatType(w, rt.Key(), false)
			w.Write(bracketClose)
			t.formatType(w, rt.Elem(), false)
			return
		case reflect.Interface:
			w.Write(interfacet)
			if l := rt.NumMethod(); l > 0 {
				w.Write(space)
				w.Write(braceOpen)
				ip := indentPrinter{w}
				for i := 0; i < l; i++ {
					ip.Write(newLine)
					m := rt.Method(i)
					io.WriteString(&ip, m.Name)
					t.formatType(&ip, m.Type, true)
				}
				w.Write(newLine)
			} else {
				w.Write(braceOpen)
			}
			w.Write(braceClose)
			return
		case reflect.Chan:
			w.Write(chant)
		case reflect.Func:
			if !inInterface {
				w.Write(funct)
			}
			w.Write(parenOpen)
			if in := rt.NumIn(); in > 0 {
				for i := 0; i < in; i++ {
					if i > 0 {
						w.Write(comma)
						w.Write(space)
					}
					if i == in-1 && rt.IsVariadic() {
						w.Write(ellipsis)
						t.formatType(w, rt.In(i).Elem(), false)
					} else {
						t.formatType(w, rt.In(i), false)
					}
				}
			}
			w.Write(parenClose)
			w.Write(space)
			out := rt.NumOut()
			if out != 1 {
				w.Write(parenOpen)
			}
			for i := 0; i < out; i++ {
				if i > 0 {
					w.Write(comma)
					w.Write(space)
				}
				t.formatType(w, rt.Out(i), false)
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
