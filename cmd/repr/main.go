package main

import (
	"bufio"
	"fmt"
	"jingle/eval"
	"jingle/parser"
	"jingle/scanner"
	"os"
)

func printError(str string) {
	fmt.Fprintf(os.Stdout, "\x1b[1;31m%s\x1b[0m\n", str)
}

func main() {
	fn := "<stdin>"
	ev := eval.NewContext()
	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(os.Stdout, "> ")
		os.Stdout.Sync()
		if !sc.Scan() {
			break
		}
		scanner := scanner.New(fn, sc.Text())
		scanner.ScanAll()
		if scanner.Errors() != nil {
			for _, err := range scanner.Errors() {
				printError(err.Error())
			}
			continue
		}
		p := parser.New(fn, scanner.Tokens())
		prog, err := p.Parse()
		if err != nil {
			printError(err.Error())
			continue
		}
		val := ev.Eval(prog)
		if err, ok := val.(*eval.Error); ok {
			if z, ok := err.Reason.(*eval.String); ok {
				printError(z.String())
			} else {
				printError(fmt.Sprintf("%+v", err))
			}
		} else {
			fmt.Printf("%+v\n", val)
		}
	}
}
