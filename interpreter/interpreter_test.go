package interpreter

import (
	"testing"

	"github.com/codetesla51/golexer/golexer"
	"github.com/codetesla51/logos/parser"
)

func TestEvalProgram(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		// last value returned
		{"5; 10; 15;", "15", "returns last statement"},
		{"5; 10;", "10", "returns last of two statements"},

		// early return
		{"return 5; 10;", "5", "early return stops execution"},
		{"return 10; return 20;", "10", "first return wins"},

		// error propagation
		{"x + 5;", "ERROR: identifier not found: x", "error bubbles up"},
		{"5 / 0;", "ERROR: division by zero", "division by zero error"},
		{"5 % 0;", "ERROR: modulo by zero", "modulo by zero error"},
		{"5 + true;", "ERROR: type mismatch: INTEGER + BOOLEAN", "type mismatch error"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			lexer := golexer.NewLexer(tc.input)
			p := parser.NewParser(lexer)
			program := p.Parse()
			i := NewInterpreter()
			result := i.Eval(program, i.Env)

			if result.String() != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, result.String())
			}
		})
	}
}

func TestFloatInfix(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"5.0 + 2.5", "7.5", "float addition"},
		{"5.0 - 2.5", "2.5", "float subtraction"},
		{"5.0 * 2.0", "10", "float multiplication"},
		{"5.0 / 2.0", "2.5", "float division"},
		{"5.0 / 0.0", "ERROR: division by zero", "float division by zero"},
		{"5.0 == 5.0", "true", "float equality true"},
		{"5.0 == 4.0", "false", "float equality false"},
		{"5.0 != 4.0", "true", "float not equal"},
		{"5.0 > 3.0", "true", "float greater than"},
		{"5.0 < 3.0", "false", "float less than"},
		{"5.0 >= 5.0", "true", "float greater or equal"},
		{"5.0 <= 4.0", "false", "float less or equal"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			lexer := golexer.NewLexer(tc.input)
			p := parser.NewParser(lexer)
			program := p.Parse()
			i := NewInterpreter()
			result := i.Eval(program, i.Env)

			if result.String() != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, result.String())
			}
		})
	}
}

func TestStringInfix(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{`"hello" + " world"`, "hello world", "string concatenation"},
		{`"hello" == "hello"`, "true", "string equality true"},
		{`"hello" == "world"`, "false", "string equality false"},
		{`"hello" != "world"`, "true", "string not equal"},
		{`"hello" - "world"`, "ERROR: unknown operator: STRING - STRING", "invalid string operator"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			lexer := golexer.NewLexer(tc.input)
			p := parser.NewParser(lexer)
			program := p.Parse()
			i := NewInterpreter()
			result := i.Eval(program, i.Env)

			if result.String() != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, result.String())
			}
		})
	}
}

func TestIntegerInfix(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"10 % 3", "1", "modulo basic"},
		{"10 % 5", "0", "modulo even division"},
		{"10 % 0", "ERROR: modulo by zero", "modulo by zero"},
		{"10 / 0", "ERROR: division by zero", "division by zero"},
		{"10 == 10", "true", "integer equality"},
		{"10 != 5", "true", "integer not equal"},
		{"10 >= 10", "true", "greater or equal"},
		{"10 <= 9", "false", "less or equal false"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			lexer := golexer.NewLexer(tc.input)
			p := parser.NewParser(lexer)
			program := p.Parse()
			i := NewInterpreter()
			result := i.Eval(program, i.Env)

			if result.String() != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, result.String())
			}
		})
	}
}
func TestIfExpression(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		// basic truthy/falsy
		{"if (true) { 10 }", "10", "true condition"},
		{"if (false) { 10 }", "null", "false condition no alternative returns null"},
		{"if (1) { 10 }", "10", "integer is truthy"},
		{"if (null) { 10 }", "null", "null is falsy"},

		// with else
		{"if (true) { 10 } else { 20 }", "10", "true takes consequence"},
		{"if (false) { 10 } else { 20 }", "20", "false takes alternative"},

		// condition is expression
		{"if (5 > 3) { 10 }", "10", "expression condition true"},
		{"if (5 < 3) { 10 } else { 20 }", "20", "expression condition false"},
		{"if (5 == 5) { 10 }", "10", "equality condition"},

		// condition uses variables
		{"let x = 5; if (x > 3) { 10 }", "10", "variable in condition"},
		{"let x = 1; if (x > 3) { 10 } else { 20 }", "20", "variable condition false"},

		// block returns last value
		{"if (true) { 5; 10; 15; }", "15", "block returns last statement"},
		{"if (false) { 5 } else { 10; 20; }", "20", "else block returns last statement"},

		// early return inside block (stays wrapped)
		{"if (true) { return 5; 10; }", "5", "return inside if stops block"},
		{"if (false) { 10 } else { return 20; 30; }", "20", "return inside else stops block"},

		// error in condition bubbles up
		{"if (x) { 10 }", "ERROR: identifier not found: x", "error in condition bubbles up"},
		{"if (5 + true) { 10 }", "ERROR: type mismatch: INTEGER + BOOLEAN", "type mismatch in condition"},

		// nested if
		{"if (true) { if (true) { 10 } }", "10", "nested if"},
		{"if (true) { if (false) { 10 } else { 20 } }", "20", "nested if else"},
		{"if (true) { if (false) { 10 } }", "null", "nested if no alternative null"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			lexer := golexer.NewLexer(tc.input)
			p := parser.NewParser(lexer)
			program := p.Parse()
			i := NewInterpreter()
			result := i.Eval(program, i.Env)

			if result.String() != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, result.String())
			}
		})
	}
}
