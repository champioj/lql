package lql

import (
	"os"
	"testing"

	"github.com/champioj/lql/typer"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func newTestMachine() *Machine {
	m := NewMachine()
	m.RegisterTyp(typer.Int4{})
	m.RegisterTyp(typer.String32{})
	lines := []string{
		"Load S, testdata/S.csv",
		"Load P, testdata/P.csv",
		"Load SP, testdata/SP.csv",
	}
	for _, v := range lines {
		_, err := m.Execute(v)
		if err != nil {
			panic(err)
		}
	}
	return m
}

func TestPrint(t *testing.T) {
	m := newTestMachine()
	info, err := m.Execute("Print S")
	if err != nil {
		t.Fatalf("Error during execution: %v", err)
	}
	wanted := `PNO:Int4	PNAME:String32	COLOR:String32	WEIGHT:Int4	CITY:String32	
1	Nut		Red		12		London		
2	Bolt		Green		17		Paris		
3	Screw		Blue		17		Oslo		
4	Screw		Red		14		London		
5	Cam		Blue		12		Paris		
6	Cog		Red		19		London`
	if info != wanted {
		t.Fatalf("Got: %q\nWanted: %q", info, wanted)
	}
	return
}

// TODO these kind of test should be generated from a template
// perhaps there should be a way to print in a normalized way?
func TestJoin(t *testing.T) {
	m := newTestMachine()
	info, err := m.Execute("Join dst, S, SP")
	if err != nil {
		t.Fatalf("Error during execution: %v", err)
	}
	info, err = m.Execute("Print dst")
	if err != nil {
		t.Fatalf("Error during execution: %v", err)
	}

	wanted := `PNO:Int4	PNAME:String32	COLOR:String32	WEIGHT:Int4	CITY:String32	SNO:Int4	QTY:Int4	
1	Nut		Red		12		London		1	1	
1	Nut		Red		12		London		2	1	
2	Bolt		Green		17		Paris		1	2	
2	Bolt		Green		17		Paris		2	2	
2	Bolt		Green		17		Paris		3	2	
2	Bolt		Green		17		Paris		4	2	
3	Screw		Blue		17		Oslo		1	3	
4	Screw		Red		14		London		4	4	
5	Cam		Blue		12		Paris		1	5	
5	Cam		Blue		12		Paris		4	5	
6	Cog		Red		19		London		1	6`
	if err != nil {
		t.Fatalf("Error during execution: %v", err)
	}
	if info != wanted {
		t.Fatalf("Got: %q\nWanted: %q", info, wanted)
	}
	return
}
