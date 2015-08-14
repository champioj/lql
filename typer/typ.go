package typer

import (
	"fmt"
	"strconv"
	"strings"
)

type Typ interface {
	Len() int
	Name() string
	String(b []byte) string
	FromString(b []byte, s string) error
	Get(b []byte) interface{}
	Set(b []byte, data interface{})
}

type Adder interface {
	Add(b, i, j []byte)
}

type Multer interface {
	Mult(b, i, j []byte)
}

type Unquoter interface {
	Unquote(s string) (string, error)
}

func init() {
	types = make(map[string]Typ, 0)
	Register(Int4{})
	Register(NewString(32))
}

var types map[string]Typ

func Register(typ Typ) {
	types[typ.Name()] = typ
}

func Get(s string) (Typ, bool) {
	t, ok := types[s]
	return t, ok
}

type Int4 struct{}

func (t Int4) Len() int {
	return 4
}

func (t Int4) Name() string {
	return "Int4"
}

func (t Int4) String(b []byte) string {
	return strconv.Itoa(int(b[0])<<24 + int(b[1])<<16 + int(b[2])<<8 + int(b[3]))
}

func (t Int4) FromString(b []byte, s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("typ: cannot convert %v", s)
	}
	d := int32(i)
	b[0] = byte(d >> 24)
	b[1] = byte(d >> 16)
	b[2] = byte(d >> 8)
	b[3] = byte(d >> 0)
	return nil
}

func (t Int4) Get(b []byte) interface{} {
	return int32(int(b[0])<<24 + int(b[1])<<16 + int(b[2])<<8 + int(b[3]))
}

func (t Int4) Set(b []byte, data interface{}) {
	d := data.(int32)
	b[0] = byte(d >> 24)
	b[1] = byte(d >> 16)
	b[2] = byte(d >> 8)
	b[3] = byte(d >> 0)
}

func (t Int4) Add(b, i, j []byte) {
	v1 := t.Get(i).(int32)
	v2 := t.Get(j).(int32)
	t.Set(b, v1+v2)
}

func (t Int4) Mult(b, i, j []byte) {
	v1 := t.Get(i).(int32)
	v2 := t.Get(j).(int32)
	t.Set(b, v1*v2)
}

type String struct {
	len int
}

func NewString(l int) String {
	return String{l}
}

func (t String) Len() int {
	return t.len
}

func (t String) Name() string {
	return "String" + strconv.Itoa(t.len)
}

func (t String) String(b []byte) string {
	return strings.Trim(string(b), "\x00")
}

func (t String) FromString(b []byte, s string) error {
	for i, bc := range []byte(s)[0:min(len(s), t.len-1)] {
		b[i] = bc
	}
	return nil
}

func (t String) Get(b []byte) interface{} {
	return string(b)
}

func (t String) Set(b []byte, data interface{}) {
	s := data.(string)
	for i, bc := range []byte(s)[0:min(len(s), t.len-1)] {
		b[i] = bc
	}
}

func (t String) Unquote(s string) (string, error) {
	var err error
	s, err = strconv.Unquote(s)
	if err != nil {
		return "", fmt.Errorf("lql: cannot unqote %v : %v", s, err)
	}
	return s, nil
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}
