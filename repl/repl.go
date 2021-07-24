package repl

import (
	"bufio"
	"fmt"
	"io"
	"jingle/evaluator"
	"jingle/lexer"
	"jingle/object"
	"jingle/parser"
	"os"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	env.Set("exit", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			os.Exit(0)
			return evaluator.NULL
		},
	})
	io.WriteString(out, JINGLE_BELL)

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		l.Filename = "<stdin>"
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

const JINGLE_BELL = `    _
   /\` + "`" + `--.
  |o-|   )>=====o
   \/.--'
`

func printParserErrors(out io.Writer, errors []string) {
	// io.WriteString(out, JINGLE_BELL)
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
