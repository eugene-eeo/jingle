package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"jingle/lexer"
	"jingle/parser"
	"os"
	"reflect"
	"strings"
)

const (
	ERR_FORMAT = "\x1b[1;31m"
	ERR_RESET  = "\x1b[0m"
)

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
		w := bufio.NewWriter(os.Stdout)
		Frprint(w, program, 0)
		w.Flush()
		fmt.Printf("\n")
	}
}

type w interface {
	io.StringWriter
	io.Writer
}

func Frprint(out w, value interface{}, level int) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Ptr:
		fmt.Fprintf(out, "&(")
		Frprint(out, v.Elem().Interface(), level)
		fmt.Fprintf(out, ")")
	case reflect.Slice:
		indent := strings.Repeat("  ", level)
		fmt.Fprintf(out, "%T[", value)
		for i := 0; i < v.Len(); i++ {
			fmt.Fprintf(out, "\n%s", indent)
			Frprint(out, v.Index(i).Interface(), level+1)
			out.WriteString(",")
			fmt.Fprintf(out, "\n")
		}
		fmt.Fprintf(out, "%s]", indent)
	case reflect.Struct:
		indent := strings.Repeat("  ", level)
		fmt.Fprintf(out, "%T{", value)
		t := reflect.TypeOf(value)
		for i := 0; i < t.NumField(); i++ {
			if !t.Field(i).Anonymous {
				name := t.Field(i).Name
				fmt.Fprintf(out, "\n%s%s: ", indent, name)
				Frprint(out, v.FieldByName(name).Interface(), level+1)
				out.WriteString(",")
			}
		}
		fmt.Fprintf(out, "\n%s}", indent)
	default:
		fmt.Fprint(out, value)
	}
}
