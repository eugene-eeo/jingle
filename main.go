package main

import (
	"jingle/repl"
	"os"
)

func main() {
	repl.Start(os.Stdin, os.Stdout)
}
