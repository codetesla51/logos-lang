package main

import (
	"fmt"

	"github.com/codetesla51/golexer/golexer"
	"github.com/codetesla51/logos/parser"
)

func main() {
	testCases := []string{
		// Decimal integers
		"let x = 42;",
		"let y = 0;",
		"let z = 1000;",

		// Floats
		"let a = 3.14;",
		"let b = 0.5;",
		"let c = 2.0;",

		// Hexadecimal
		"let hex = 0xFF;",
		"let hex2 = 0x1A2B;",

		// Binary
		"let bin = 0b1010;",
		"let bin2 = 0B1111;",

		// Octal
		"let oct = 0o777;",
		"let oct2 = 0755;",

		// Booleans
		"let flag = true;",
		"let other = false;",

		// Expressions
		"5 + 3 * 2;",
		"-42 + 10;",
		"!true;",

		// If expressions
		"if (x > 5) { x + 1; } else { x - 1; }",

		// Complex
		"let result = if (0xFF > 100) { true } else { false };",
		"let func = fn(x, y) { x + y; };",
		"let add = fn(x,y) {return x && y; };",
		"add(1,2);",
		"let name = \"Alice\";",
		"let raw = `This is a raw string\nwith a newline.`;",
		"let arr = [1, 2, 3, 4];",
		"arr[i + 4]",
		"let num = 10 % 3;",
		"let i = 0; for (i < 10) { i = i + 1; }",
		"switch (expression) { case value { statements } default { statements } }",
		"let name = null",
		"for (item in collection) { /* statements */ }",
		"let result = fn(a,b) -> a + b;",
		"table{key1: value1, key2: value2}",
		"table{key1: value1, key2: value2}[key1]",
		"table{}",
		"switch (x) { case 1 { 10 } }",
		"",
		"let x = if (x >= y){return y}", // Syntax error
	}

	for _, input := range testCases {
		fmt.Printf("\nInput:  %s\n", input)
		lexer := golexer.NewLexerWithConfig(input, "tokens.json")
		p := parser.NewParser(lexer)
		program := p.Parse()

		if len(p.Errors()) != 0 {
			fmt.Printf("Errors (%d):\n", len(p.Errors()))
			for _, err := range p.Errors() {
				fmt.Printf("   - %s\n", err)
			}
			continue
		}

		if program == nil {
			fmt.Printf("Parse returned nil\n")
			continue
		}

		fmt.Printf("AST:    %s\n", program.String())
	}

}
