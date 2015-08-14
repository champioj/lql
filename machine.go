package lql

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/champioj/lql/typer"
)

type Machine struct {
	rels map[string]Relation
	pred []opLine
}

type opLine struct {
	op    opCode
	arg   string
	value []byte
}

//go:generate stringer -type=opCode
type opCode uint8

const (
	// Load relationDst, csv_pathfile
	Load opCode = iota + 1
	// Print relationSrc
	Print
	// Join relationDst, relationSrc1, relationSrc2
	Join
	// Rename relationDst, relationSrc, attrOld, attrNew
	Rename
	// Project relationDst, relationSrc, attr1, attr2, ..., attrN
	Project
	// Restrict relationDst, relationSrc
	Restrict

	// Pushv typ, value
	Pushv
	// Pusha attr
	Pusha

	// OpEq typ
	OpEq
	// OpGT typ
	OpGT

	// OpAnd
	OpAnd
	// OpOr
	OpOr

	// OpAdd typ
	OpAdd
	// OpMult typ
	OpMult
)

var opCodeStr = map[string]opCode{
	"Load":     Load,
	"Print":    Print,
	"Join":     Join,
	"Rename":   Rename,
	"Project":  Project,
	"Restrict": Restrict,
	"Pushv":    Pushv,
	"Pusha":    Pusha,
	"OpEq":     OpEq,
	"OpGT":     OpGT,
	"OpAnd":    OpAnd,
	"OpOr":     OpOr,
	"OpAdd":    OpAdd,
	"OpMult":   OpMult,
}

func NewMachine() *Machine {
	m := Machine{}
	m.rels = make(map[string]Relation, 0)
	m.pred = make([]opLine, 0)
	return &m
}

func extract(line string) (op opCode, args []string, err error) {
	opStr := strings.SplitN(line, " ", 2)
	ok := false
	op, ok = opCodeStr[opStr[0]]
	if !ok {
		return 0, nil, fmt.Errorf("lql: unknown opCode: %v", opStr[0])
	}
	if len(opStr) > 1 {
		args = strings.Split(opStr[1], ",")
	} else {
		args = make([]string, 0)
	}
	for k, _ := range args {
		args[k] = strings.TrimSpace(args[k])
	}
	return
}

func (m *Machine) Execute(line string) (info string, err error) {
	op, args, err := extract(line)
	if err != nil {
		return "", fmt.Errorf("lql: error executing %v: %v", line, err)
	}
	switch op {
	case Load:
		if len(args) != 2 {
			return "", fmt.Errorf("lql: 2 argument needed for Load, %d provided", len(args))
		}

		file, err := os.Open(args[1])
		if err != nil {
			return "", fmt.Errorf("lql: Error when opening file: '%s'", err)
		}
		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			return "", fmt.Errorf("lql: Error reading file: %s", err)
		}

		header, records := records[0], records[1:]
		attrs := make(Header, 0, len(header))

		for _, v := range header {
			nameType := strings.Split(v, ":")
			if len(nameType) != 2 {
				return "", fmt.Errorf("lql: column header %s must have a name and a type %s", v)
			}
			typ, ok := typer.Get(nameType[1])
			if !ok {
				return "", fmt.Errorf("lql: type %s not registered", nameType[1])
			}
			attrs = append(attrs, Attr{nameType[0], typ})
		}

		table := NewRelation(attrs)
		for _, v := range records {
			tuple, err := table.NewTuple(v)
			if err != nil {
				return "", fmt.Errorf("lql: unable to add tuple, reason: %s", err)
			}
			table.Add(tuple)
		}
		m.rels[args[0]] = *table
		break
	case Print:
		if len(args) != 1 {
			return "", fmt.Errorf("lql: 1 argument needed for Print, %d provided", len(args))
		}
		table, ok := m.rels[args[0]]
		if !ok {
			return "", fmt.Errorf("lql: unknown table, %s ", args[0])
		}
		//fmt.Println(table)
		info = fmt.Sprint(table)
	case Join:
		if len(args) != 3 {
			return "", fmt.Errorf("lql: Join must have 3 arguments")
		}
		table1, ok1 := m.rels[args[1]]
		if !ok1 {
			return "", fmt.Errorf("lql: table %s unknow", args[1])
		}
		table2, ok2 := m.rels[args[2]]
		if !ok2 {
			return "", fmt.Errorf("lql: table %s unknow", args[2])
		}
		newTable := join(table1, table2)
		m.rels[args[0]] = *newTable
		break
	case Rename:
		if len(args) != 4 {
			return "", fmt.Errorf("lql: Rename must have 4 arguments")
		}
		tableSrc, ok1 := m.rels[args[1]]
		if !ok1 {
			return "", fmt.Errorf("lql: table %s unknow", args[1])
		}
		newTable := rename(tableSrc, args[2], args[3])
		m.rels[args[0]] = *newTable
		break
	case Project:
		if len(args) < 3 {
			return "", fmt.Errorf("lql: Project must have at least 3 arguments")
		}
		tableSrc, ok1 := m.rels[args[1]]
		if !ok1 {
			return "", fmt.Errorf("lql: table %s unknow", args[1])
		}
		newTable := project(tableSrc, args[2:])
		m.rels[args[0]] = *newTable
	case Restrict:
		if len(args) != 2 {
			return "", fmt.Errorf("lql: Restrict must have 2 arguments")
		}
		tableSrc, ok1 := m.rels[args[1]]
		if !ok1 {
			return "", fmt.Errorf("lql: table %s unknow", args[1])
		}
		newTable := m.restrict(tableSrc, m.pred)
		m.pred = m.pred[0:0]
		m.rels[args[0]] = *newTable
		break
	case OpEq:
		if len(args) != 1 {
			return "", fmt.Errorf("lql: Op must have 1 arguments")
		}
		l := opLine{op: OpEq, arg: args[0], value: nil}
		m.pred = append(m.pred, l)
		break
	case OpGT:
		if len(args) != 1 {
			return "", fmt.Errorf("lql: Op must have 1 arguments")
		}
		l := opLine{op: OpGT, arg: args[0], value: nil}
		m.pred = append(m.pred, l)
		break
	case OpAnd:
		if len(args) != 0 {
			return "", fmt.Errorf("lql: Op must have 0 arguments")
		}
		l := opLine{op: OpAnd, arg: "", value: nil}
		m.pred = append(m.pred, l)
		break
	case OpOr:
		if len(args) != 0 {
			return "", fmt.Errorf("lql: Op must have 0 arguments")
		}
		l := opLine{op: OpOr, arg: "", value: nil}
		m.pred = append(m.pred, l)
		break
	case OpAdd:
		if len(args) != 1 {
			return "", fmt.Errorf("lql: Op must have 1 arguments")
		}
		l := opLine{op: OpAdd, arg: args[0], value: nil}
		m.pred = append(m.pred, l)
		break
	case OpMult:
		if len(args) != 1 {
			return "", fmt.Errorf("lql: Op must have 1 arguments")
		}
		l := opLine{op: OpMult, arg: args[0], value: nil}
		m.pred = append(m.pred, l)
		break
	case Pusha:
		if len(args) != 1 {
			return "", fmt.Errorf("lql: Pushr must have 1 arguments")
		}
		l := opLine{op: Pusha, arg: args[0], value: nil}
		m.pred = append(m.pred, l)
		break
	case Pushv:
		if len(args) != 2 {
			return "", fmt.Errorf("lql: Pushv must have 2 arguments")
		}
		lTyp, ok := typer.Get(args[0])
		if !ok {
			return "", fmt.Errorf("lql: Typ %v unknown", args[0])
		}
		if t, ok := lTyp.(typer.Unquoter); ok {
			unquoted, err := t.Unquote(args[1])
			args[1] = unquoted
			if err != nil {
				return "", fmt.Errorf("lql: Could not unqote %v", args[1])
			}
		}
		v := make([]byte, lTyp.Len())
		err := lTyp.FromString(v, args[1])
		if err != nil {
			return "", fmt.Errorf("lql: Error parsing value %v of type %v", args[1], args[0])
		}
		l := opLine{op: Pushv, arg: args[0], value: v}
		m.pred = append(m.pred, l)
		break
	default:
		return "", fmt.Errorf("lql: Op code '%s' not implemented", op)
		break
	}
	return
}
