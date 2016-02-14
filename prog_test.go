package lql

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
	"strings"
	"testing"
)

type prog struct {
	Code    string
	Results map[string]string
}

// TODO this should be defined in the toml file
func newTestMachine() *Machine {
	m := NewMachine()
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

func TestProgs(t *testing.T) {
	progsPath := []string{
		"testdata/prog/basicop.toml",
	}
	for _, path := range progsPath {
		var progs map[string]prog
		_, err := toml.DecodeFile(path, &progs)
		if err != nil {
			t.Fatalf("Error decoding toml file: %v", err)
		}
		for name, v := range progs {
			t.Logf("Testing %v\n", name)
			testProg(t, v)
		}
	}
}

func testProg(t *testing.T, p prog) {
	m := newTestMachine()
	reader := bufio.NewReader(strings.NewReader(p.Code))
	for {
		var line string

		line, err := reader.ReadString('\n')

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("console input error: ", err)
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		_, err = m.Execute(line)
		if err != nil {
			fmt.Println("machine error: ", err)
		}
	}
	for relName, want := range p.Results {
		got, err := m.Execute(fmt.Sprint("Print ", relName))
		if err != nil {
			fmt.Println("machine error while comparing: ", err)
		}
		if got != want {
			t.Logf("...%v Fail\n", relName)
			t.Logf("Wanted:\n%v\nGot:\n%v\n", want, got)
			t.Fail()
		}
	}
}
