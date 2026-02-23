package parser

import (
	"testing"

	"github.com/codetesla51/golexer/golexer"
)

func TestParser(t *testing.T) {
	input := "5 + 3"
	l := golexer.NewLexer(input)
	p := NewParser(l)
	program := p.Parse()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors", len(p.Errors()))
	}
	if program == nil {
		t.Fatalf("Parse() returned nil")
	}
	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ExpressionStatement. got=%T", program.Statements[0])
	}
	exp, ok := stmt.Expression.(*InfixExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *InfixExpression. got=%T", stmt.Expression)
	}
	if exp.Operator != "+" {
		t.Fatalf("exp.Operator is not '+'. got=%q", exp.Operator)
	}
	if exp.Left.String() != "5" {
		t.Fatalf("exp.Left.String() is not '5'. got=%q", exp.Left.String())
	}
	if exp.Right.String() != "3" {
		t.Fatalf("exp.Right.String() is not '3'. got=%q", exp.Right.String())
	}
}
func TestParserErrors(t *testing.T) {
	input := "5 +"
	l := golexer.NewLexer(input)
	p := NewParser(l)
	program := p.Parse()
	if len(p.Errors()) == 0 {
		t.Fatalf("parser should have errors but got none")
	}
	if program != nil {
		t.Fatalf("Parse() should return nil when there are errors. got=%v", program)
	}
}
func TestParserPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5 + 3", "(5 + 3)"},
		{"5 + 3 * 2", "(5 + (3 * 2))"},
		{"5 * 3 + 2", "((5 * 3) + 2)"},
		{"(5 + 3) * 2", "((5 + 3) * 2)"},
	}
	for _, tt := range tests {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors", len(p.Errors()))
		}
		if program.String() != tt.expected {
			t.Fatalf("expected=%q, got=%q", tt.expected, program.String())
		}
	}
}
