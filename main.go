package main

import (
	"fmt"
	"interpreter/evaluator"
	"interpreter/lexer"
	"interpreter/object"
	"interpreter/parser"
	"io"
	"os"
)

func main() {
	// repl.Start(os.Stdin, os.Stdout)
	filePath := "program.txt" // Replace with the actual path to your file

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	source_code := string(content)

	lexer := lexer.New(source_code)
	pars := parser.New(lexer)

	program := pars.ParseProgram()
	if len(pars.Errors()) != 0 {
		PrintParserErrors(os.Stdout, pars.Errors())
		return
	}

	env := object.NewEnvironment()
	evaluator.Eval(program, env)
}

func PrintParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
