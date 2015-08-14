package lql

import (
	"bytes"
	"fmt"

	"github.com/champioj/lql/typer"
)

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func eval(s []byte, header Header, tuple Tuple, line opLine) ([]byte, error) {
	switch line.op {
	case Pusha:
		attr, pos, err := header.AttrByName(line.arg)
		if err != nil {
			return nil, fmt.Errorf("lql: error executing : %v", err)
		}
		s = append(s, tuple[pos:pos+attr.Typ.Len()]...)
		break
	case Pushv:
		s = append(s, line.value...)
		break
	case OpAdd:
		typ, ok := typer.Get(line.arg)
		if !ok {
			fmt.Errorf("lql: unknown type")
		}
		adder, ok := typ.(typer.Adder)
		if !ok {
			fmt.Errorf("lql: type %v is not an Adder", typ.Len())
		}
		buf1 := s[len(s)-typ.Len()*2 : len(s)-typ.Len()]
		buf2 := s[len(s)-typ.Len():]
		adder.Add(buf1, buf1, buf2)
		s = s[:len(s)-typ.Len()]
		break
	case OpMult:
		typ, ok := typer.Get(line.arg)
		if !ok {
			fmt.Errorf("lql: unknown type")
		}
		multer, ok := typ.(typer.Multer)
		if !ok {
			fmt.Errorf("lql: type %v is not a Multer", typ.Len())
		}
		buf1 := s[len(s)-typ.Len()*2 : len(s)-typ.Len()]
		buf2 := s[len(s)-typ.Len():]
		multer.Mult(buf1, buf1, buf2)
		s = s[:len(s)-typ.Len()]
		break
	case OpGT:
		typ, ok := typer.Get(line.arg)
		if !ok {
			return nil, fmt.Errorf("lql: unknown type")
		}
		var vb1, vb2 []byte
		vb1, s = s[len(s)-typ.Len():], s[:len(s)-typ.Len()]
		vb2, s = s[len(s)-typ.Len():], s[:len(s)-typ.Len()]
		if bytes.Compare(vb1, vb2) > 0 {
			s = append(s, byte(0))
		} else {
			s = append(s, byte(1))
		}

		break
	case OpOr:
		v1 := s[len(s)-1]
		v2 := s[len(s)-2]
		if v1 > 0 || v2 > 0 {
			s[len(s)-2] = 1
		} else {
			s[len(s)-2] = 0
		}
		s = s[:len(s)-1]
		break
	case OpEq:
		typ, ok := typer.Get(line.arg)
		if !ok {
			return nil, fmt.Errorf("lql: unknown type")
		}
		var vb1, vb2 []byte
		vb1, s = s[len(s)-typ.Len():], s[:len(s)-typ.Len()]
		vb2, s = s[len(s)-typ.Len():], s[:len(s)-typ.Len()]
		if bytes.Compare(vb1, vb2) == 0 {
			s = append(s, byte(1))
		} else {
			s = append(s, byte(0))
		}
		break
	default:
		panic("not a operator")
		break
	}
	return s, nil
}

func (m *Machine) restrict(r Relation, line []opLine) *Relation {
	newR := NewRelation(r.H.Copy())

	for _, t := range r.B {
		stack := make([]byte, 0)

		for _, l := range line {
			var err error
			stack, err = eval(stack, r.H, t, l)
			if err != nil {
				panic(err) // TODO rethink error handling
			}
		}
		if len(stack) != 1 {
			panic("stack must have only one byte (bolean) to restrict")
		}
		if stack[0] > 0 {
			newR.Add(t.Copy())
		}
	}
	return newR
}

func project(r Relation, names []string) *Relation {
	toKeep := []struct {
		iHeader int
		iTuple  int
	}{}

	newHeader := make(Header, 0, len(names))

	iTuple := 0
	for i, h := range r.H {
		if contains(names, h.Name) {
			eq := struct {
				iHeader int
				iTuple  int
			}{i, iTuple}
			toKeep = append(toKeep, eq)
			newHeader = append(newHeader, h)
		}
		iTuple += h.Typ.Len()
	}
	rDst := NewRelation(newHeader)

	for _, t := range r.B {
		cNew := 0
		cOld := 0
		newT := make(Tuple, newHeader.Len())
		for _, keep := range toKeep {
			lenToCopy := r.H[keep.iHeader].Typ.Len()
			cOld = keep.iTuple
			copy(newT[cNew:cNew+lenToCopy], t[cOld:cOld+lenToCopy])
			cNew += lenToCopy
		}

		rDst.Add(newT)
	}
	return rDst
}

func rename(r Relation, old string, new string) *Relation {
	// TODO add errors (old string notfound and new string already there)
	// should they be here or one level top?
	newHeader := make(Header, 0, len(r.H))

	for _, h := range r.H {
		if h.Name == old {
			h.Name = new
		}
		newHeader = append(newHeader, h)
	}
	tDst := NewRelation(newHeader)
	tDst.B = r.B
	return tDst
}

func join(r1 Relation, r2 Relation) *Relation {
	pairs := []struct {
		iHeader1 int
		iTuple1  int
		iHeader2 int
		iTuple2  int
	}{}

	iTuple1 := 0
	for i1, h1 := range r1.H {
		iTuple2 := 0
		for i2, h2 := range r2.H {
			if h1.Equals(h2) {
				eq := struct {
					iHeader1 int
					iTuple1  int
					iHeader2 int
					iTuple2  int
				}{i1, iTuple1, i2, iTuple2}
				pairs = append(pairs, eq)
			}
			iTuple2 += h2.Typ.Len()
		}
		iTuple1 += h1.Typ.Len()
	}

	newHeader := make(Header, len(r1.H))
	// we can safely copy the whole header of the first relation
	copy(newHeader, r1.H)
	// and add the non-redonant column
headers:
	for i, h := range r2.H {
		for _, pair := range pairs {
			if i == pair.iHeader2 {
				continue headers
			}
		}
		newHeader = append(newHeader, h)
	}

	rDst := NewRelation(newHeader)
	tLen := newHeader.Len()

	for _, t1 := range r1.B {
		for _, t2 := range r2.B {
			equals := true
			for _, pair := range pairs { // TODO make one or two small helpers :)
				//fmt.Printf("%v:%v\n", t1[pair.iTuple1:pair.iTuple1+r1.H[pair.iHeader1].Typ.length], t2[pair.iTuple2:pair.iTuple2+r2.H[pair.iHeader2].Typ.length])
				if bytes.Compare(t1[pair.iTuple1:pair.iTuple1+r1.H[pair.iHeader1].Typ.Len()],
					t2[pair.iTuple2:pair.iTuple2+r2.H[pair.iHeader2].Typ.Len()]) != 0 {
					equals = false
				}
			}
			if !equals {
				continue
			}

			newT := make(Tuple, tLen)
			copy(newT, t1)

			cNew := len(t1)
			cT2 := 0

			for _, pair := range pairs {
				lenToCopy := pair.iTuple2 - cT2
				if lenToCopy == 0 {
					continue
				}
				copy(newT[cNew:cNew+lenToCopy],
					t2[cT2:cT2+lenToCopy])
				cNew += lenToCopy
				cT2 += r2.H[pair.iHeader2].Typ.Len()
			}
			copy(newT[cNew:], t2[cT2:]) // copy the rest
			rDst.Add(newT)
		}
	}
	return rDst
}
