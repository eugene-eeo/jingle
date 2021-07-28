package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"jingle/lexer"
	"jingle/parser"
	"jingle/token"
	"os"
	"reflect"
	"strings"
)

const (
	OK_FORMAT  = "\x1b[1;32m"
	ERR_FORMAT = "\x1b[1;31m"
	ERR_RESET  = "\x1b[0m"
)

func printOkEnd() {
	fmt.Printf("%s-------------------%s\n",
		OK_FORMAT,
		ERR_RESET,
	)
}

func printOkStart() {
	fmt.Printf("%s------ OK ---------%s\n",
		OK_FORMAT,
		ERR_RESET,
	)
}

func printError(err error) {
	fmt.Printf("%s------ ERROR ------%s\n%s\n%s-------------------%s\n",
		ERR_FORMAT,
		ERR_RESET,
		err,
		ERR_FORMAT,
		ERR_RESET,
	)
}

func main() {
	var deep bool
	flag.BoolVar(&deep, "deep", false, "recursively print ast")
	flag.Parse()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		buf := bytes.NewBuffer(scanner.Bytes())
		lexer, err := lexer.TryNew("<stdin>", buf)
		if err != nil {
			printError(err)
			continue
		}
		p := parser.New(lexer)
		program, err := p.Parse()
		if err != nil {
			printError(err)
			continue
		}
		if deep {
			printOkStart()
			w := bufio.NewWriter(os.Stdout)
			Frprint(w, program, 0)
			w.WriteString("\n")
			w.Flush()
			printOkEnd()
		} else {
			fmt.Println(program.String())
		}
	}
}

type w interface {
	io.StringWriter
	io.Writer
}

func Frprint(out w, value interface{}, level int) {
	if token, ok := value.(token.Token); ok {
		// compact formatting for token
		fmt.Fprintf(out, "%s(%d:%d:%q)", token.Type, token.LineNo, token.Column, token.Literal)
		return
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Ptr:
		fmt.Fprintf(out, "&")
		Frprint(out, v.Elem().Interface(), level)
	case reflect.Slice:
		indent := strings.Repeat("  ", level+1)
		fmt.Fprintf(out, "%T[", value)
		last := v.Len() - 1
		for i := 0; i < v.Len(); i++ {
			fmt.Fprintf(out, "\n%s", indent)
			Frprint(out, v.Index(i).Interface(), level+1)
			if i != last {
				out.WriteString(",")
			}
		}
		fmt.Fprintf(out, "\n%s]", strings.Repeat("  ", level))
	case reflect.Struct:
		indent := strings.Repeat("  ", level+1)
		fmt.Fprintf(out, "%T{", value)
		t := reflect.TypeOf(value)
		for i := 0; i < t.NumField(); i++ {
			if !t.Field(i).Anonymous {
				if i > 1 {
					out.WriteString(",") // from the previous iteration
				}
				name := t.Field(i).Name
				fmt.Fprintf(out, "\n%s%s: ", indent, name)
				Frprint(out, v.FieldByName(name).Interface(), level+1)
			}
		}
		fmt.Fprintf(out, "\n%s}", strings.Repeat("  ", level))
	default:
		fmt.Fprint(out, value)
	}
}
