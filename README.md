# goprint
--
    import "vimagination.zapto.org/goprint"

Package goprint allows for the printing of Go Values in Go Code

## Usage

#### type Opt

```go
type Opt func(*config)
```

Opt is a printing option

#### func  ArrayReplacer

```go
func ArrayReplacer(rf func(io.Writer, reflect.Type, int, reflect.Value) bool) Opt
```
ArrayReplacer allows for some values in an array or slice to be replaced.

This is intended to aid the replacing of constants with named constants or vars,
though other applications are applicable

Whatever is written to the Writer will display in place of the constant.

The return value should be set to true is the func has written anything, or
false otherwise

#### func  PkgName

```go
func PkgName(pf func(io.Writer, reflect.Type)) Opt
```
PkgName sets the function that will write the package name for the type given.

This should be used to override the default which takes the Base of the URL as
the import name.

#### func  StructFilter

```go
func StructFilter(sf func(reflect.Type, string) bool) Opt
```
StructFilter allows for only some fields of a Struct to be printed.

The func recieves the struct type and the field name, and should return true to
print the field, and false to not print it

#### func  StructReplacer

```go
func StructReplacer(rf func(io.Writer, reflect.Type, string, reflect.Value) bool) Opt
```
StructReplacer allows for some values in struct fields to be replaced.

This is intended to aid the replacing of constants with named constants or vars,
though other applications are applicable

Whatever is written to the Writer will display in place of the constant.

The return value should be set to true is the func has written anything, or
false otherwise

#### type Type

```go
type Type struct {
}
```

Type in a wrapped type with its print configuration

#### func  Wrap

```go
func Wrap(v interface{}, opts ...Opt) *Type
```
Wrap creates the type printer.

#### func (*Type) Format

```go
func (t *Type) Format(s fmt.State, v rune)
```
Format implements the fmt.Formatter interface

#### func (*Type) WriteTo

```go
func (t *Type) WriteTo(w io.Writer) (int64, error)
```
WriteTo implements the io.WriterTo interface
