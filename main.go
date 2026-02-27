package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/codetesla51/golexer/golexer"
	"github.com/codetesla51/logos/interpreter"
	"github.com/codetesla51/logos/parser"
)

const PROMPT = ">> "

func eval(input string, inter *interpreter.Interpreter) interpreter.Object {
	lexer := golexer.NewLexerWithConfig(input, "tokens.json")
	p := parser.NewParser(lexer)
	program := p.Parse()

	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		return nil
	}

	return inter.Eval(program, inter.Env)
}

func runFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %s\n", err)
		os.Exit(1)
	}
	inter := interpreter.NewInterpreter()
	lexer := golexer.NewLexerWithConfig(string(data), "tokens.json")
	p := parser.NewParser(lexer)
	program := p.Parse()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		return
	}
	result := inter.Eval(program, inter.Env)
	if result != nil && result.Type() == interpreter.ERROR_OBJ {
		fmt.Fprintln(os.Stderr, result.String())
	}
}

func runREPL() {
	scanner := bufio.NewScanner(os.Stdin)
	inter := interpreter.NewInterpreter()
	fmt.Println("Logos REPL (ctrl+c to exit)")
	for {
		fmt.Print(PROMPT)
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		result := eval(line, inter)
		if result == nil {
			continue
		}
		switch result.Type() {
		case interpreter.NULL_OBJ, interpreter.FUNCTION_OBJ:
			// don't print
		default:
			fmt.Println(result.String())
		}
	}
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		runREPL()
		return
	}

	path := args[0]
	if !strings.HasSuffix(path, ".lgs") {
		fmt.Fprintf(os.Stderr, "error: expected .lgs file, got %q\n", path)
		os.Exit(1)
	}

	runFile(path)
}
