package main

import (
	"bufio"
	"fmt"
	"github.com/champioj/lql"
	"io"
	"os"
	"strings"
)

const prog1 = `
Load S, ../testdata/S.csv
Load P, ../testdata/P.csv
Load SP, ../testdata/SP.csv

Print P
Print S

Pushv Int4, 20
Pusha STATUS
OpEq Int4
Restrict rs, S
Print rs

Pusha WEIGHT
Pusha WEIGHT
OpMult Int4
Pushv Int4, 150
OpGT Int4
Restrict rs, P
Print rs

Pusha WEIGHT
Pushv Int4, 5
OpAdd Int4
Pushv Int4, 20
OpGT Int4
Pushv String32, "Red"
Pusha COLOR
OpEq String32
OpOr
Restrict rs, P
Print rs
`

func main() {
	m := lql.NewMachine()

	var reader *bufio.Reader
	if false {
		reader = bufio.NewReader(os.Stdin)
	} else {
		reader = bufio.NewReader(strings.NewReader(prog1))
	}
	wd, _ := os.Getwd()
	fmt.Println("Hello, please enter some command (cwdir:", wd, "):")
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
		info, err := m.Execute(line)
		if info != "" {
			fmt.Println(info)
		}
		if err != nil {
			fmt.Println("machine error: ", err)
		}
	}
	fmt.Print("Goodbye!")
}
