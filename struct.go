package lql

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/champioj/lql/typer"
)

type Attr struct {
	Name string
	Typ  typer.Typ
}

type Header []Attr
type Tuple []byte
type Body []Tuple

type Relation struct {
	H Header
	B Body
}

func NewRelation(attrs []Attr) *Relation {
	h := Header(attrs)
	b := make([]Tuple, 0)
	return &Relation{h, b}
}

func (r *Relation) NewTuple(values []string) (Tuple, error) {
	c := 0
	tuple := make(Tuple, r.H.Len())
	for i, v := range r.H {
		err := v.Typ.FromString(tuple[c:], values[i])
		if err != nil {
			return Tuple{}, fmt.Errorf("lql: unable to add tuple, reason:", err)
		}
		c += v.Typ.Len()
	}

	return tuple, nil
}

func (r *Relation) Add(tuple Tuple) {
	// TODO it's certainly quite slow
	for _, v := range r.B {
		if bytes.Compare(tuple, v) == 0 {
			return
		}
	}

	r.B = append(r.B, tuple)
}

func (h Header) Len() int {
	c := 0
	for _, v := range h {
		c += v.Typ.Len()
	}
	return c
}

func (a Attr) Equals(a2 Attr) bool {
	// TODO if typ is null, use it as anything?
	return a.Typ.Name() == a2.Typ.Name() && a.Name == a2.Name
}

func (h Header) Pos(a Attr) (int, error) {
	c := 0
	for _, v := range h {
		if a.Equals(v) {
			return c, nil
		}
		c += v.Typ.Len()
	}
	return -1, fmt.Errorf("lql serious: attribute %s not found", a.Name)
}

func (h Header) Copy() Header {
	newAttr := make(Header, 0, len(h))
	for _, v := range h {
		newAttr = append(newAttr, v)
	}
	return newAttr
}

func (h Header) AttrByName(name string) (attr Attr, pos int, err error) {
	found := false
	c := 0
	for _, v := range h {
		if v.Name == name {
			if found == true {
				return attr, pos, fmt.Errorf("lql: duplicate name")
			}
			found = true
			attr = v
			pos = c
		}
		c += v.Typ.Len()
	}
	if found == false {
		return attr, pos, fmt.Errorf("lql serious: attribute %s not found", name)
	}
	return attr, pos, nil
}

func (t Tuple) Copy() Tuple {
	newTuple := make(Tuple, len(t))
	copy(newTuple, t)
	return newTuple
}

func (r Relation) String() string {
	w := new(tabwriter.Writer)

	buf := new(bytes.Buffer)
	w.Init(buf, 0, 10, 1, '\t', 0)
	for _, v := range r.H {
		fmt.Fprintf(w, "%s:%s\t", v.Name, v.Typ.Name())
	}
	fmt.Fprint(w)
	for _, vt := range r.B {
		i := 0
		fmt.Fprintln(w)
		for _, vh := range r.H {
			typLen := vh.Typ.Len()
			fmt.Fprintf(w, "%v\t", vh.Typ.String(vt[i:i+typLen]))
			i += typLen
		}
	}
	w.Flush()
	return buf.String()
}
