# goprint
--
    import "vimagination.zapto.org/goprint"

Package goprint allows for the printing of Go Values in Go Code.

## Usage

#### type Opt

```go
type Opt func(*Printer)
```

Opt is a printing option.

#### func  ArrayReplacer

```go
func ArrayReplacer(rf func(io.Writer, reflect.Type, int, reflect.Value) bool) Opt
```
ArrayReplacer allows for some values in an array or slice to be replaced.

This is intended to aid the replacing of constants with named constants or vars,
though other applications are applicable

Whatever is written to the Writer will display in place of the constant.

The return value should be set to true is the func has written anything, or
false otherwise.

#### func  FromPrinter

```go
func FromPrinter(p Printer) Opt
```
FromPrinter copies the configuration from an existing Printer.

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
print the field, and false to not print it.

#### func  StructReplacer

```go
func StructReplacer(rf func(io.Writer, reflect.Type, string, reflect.Value) bool) Opt
```
StructReplacer allows for some values in struct fields to be replaced.

This is intended to aid the replacing of constants with named constants or vars,
though other applications are applicable

Whatever is written to the Writer will display in place of the constant.

The return value should be set to true is the func has written anything, or
false otherwise.

#### type Printer

```go
type Printer struct {
}
```

Printer represents the options to print out a value.

#### func  New

```go
func New(opts ...Opt) *Printer
```
New creates a new Printer from the given options.

#### func (*Printer) Format

```go
func (p *Printer) Format(w io.Writer, val interface{}) (int64, error)
```
Format prints out the value to the writer.

#### func (*Printer) FormatVerbose

```go
func (p *Printer) FormatVerbose(w io.Writer, val interface{}) (int64, error)
```
FormatVerbose prints out the value to the writer adding more detail that doesn't
change the created value.

#### type Value

```go
type Value struct {
	Printer
}
```

Value in a wrapped value with its print configuration.

#### func  Wrap

```go
func Wrap(v interface{}, opts ...Opt) *Value
```
Wrap creates a Printer with an embedded value to be used with fmt-like
functions.

#### func (*Value) Format

```go
func (t *Value) Format(s fmt.State, v rune)
```
Format implements the fmt.Formatter interface.

#### func (*Value) WriteTo

```go
func (t *Value) WriteTo(w io.Writer) (int64, error)
```
WriteTo implements the io.WriterTo interface
