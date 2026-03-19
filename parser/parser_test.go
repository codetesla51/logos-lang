package parser

import (
	"testing"

	"github.com/codetesla51/golexer/golexer"
)

// this might come in handing for finding lexer bugs
//
//	l2 := golexer.NewLexer(tc.input)
//
//	for {
//		tok := l2.NextToken()
//		fmt.Printf("Type: %s | Literal: %s\n", tok.Type, tok.Literal)
//		if tok.Type == golexer.EOF {
//			break
//		}
//	}
func TestParser(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedStmts int
		expectedType  interface{}
		validate      func(t *testing.T, stmt Statement)
	}{
		{
			name:          "InfixExpression",
			input:         "5 + 3",
			expectedStmts: 1,
			validate: func(t *testing.T, stmt Statement) {
				exprStmt, ok := stmt.(*ExpressionStatement)
				if !ok {
					t.Fatalf("expected *ExpressionStatement, got %T", stmt)
				}
				infixExp, ok := exprStmt.Expression.(*InfixExpression)
				if !ok {
					t.Fatalf("expected *InfixExpression, got %T", exprStmt.Expression)
				}
				if infixExp.Operator != "+" {
					t.Fatalf("expected operator '+', got %q", infixExp.Operator)
				}
				if infixExp.Left.String() != "5" {
					t.Fatalf("expected left '5', got %q", infixExp.Left.String())
				}
				if infixExp.Right.String() != "3" {
					t.Fatalf("expected right '3', got %q", infixExp.Right.String())
				}
			},
		},
		{
			name:          "IntegerLiteral",
			input:         "42",
			expectedStmts: 1,
			validate: func(t *testing.T, stmt Statement) {
				exprStmt, ok := stmt.(*ExpressionStatement)
				if !ok {
					t.Fatalf("expected *ExpressionStatement, got %T", stmt)
				}
				intLit, ok := exprStmt.Expression.(*IntegerLiteral)
				if !ok {
					t.Fatalf("expected *IntegerLiteral, got %T", exprStmt.Expression)
				}
				if intLit.Value != 42 {
					t.Fatalf("expected value 42, got %d", intLit.Value)
				}
			},
		},
		{
			name:          "FloatLiteral",
			input:         "3.14",
			expectedStmts: 1,
			validate: func(t *testing.T, stmt Statement) {
				exprStmt, ok := stmt.(*ExpressionStatement)
				if !ok {
					t.Fatalf("expected *ExpressionStatement, got %T", stmt)
				}
				floatLit, ok := exprStmt.Expression.(*FloatLiteral)
				if !ok {
					t.Fatalf("expected *FloatLiteral, got %T", exprStmt.Expression)
				}
				if floatLit.Value != 3.14 {
					t.Fatalf("expected value 3.14, got %f", floatLit.Value)
				}
			},
		},
		{
			name:          "Identifier",
			input:         "x",
			expectedStmts: 1,
			validate: func(t *testing.T, stmt Statement) {
				exprStmt, ok := stmt.(*ExpressionStatement)
				if !ok {
					t.Fatalf("expected *ExpressionStatement, got %T", stmt)
				}
				ident, ok := exprStmt.Expression.(*Identifier)
				if !ok {
					t.Fatalf("expected *Identifier, got %T", exprStmt.Expression)
				}
				if ident.Value != "x" {
					t.Fatalf("expected identifier 'x', got %q", ident.Value)
				}
			},
		},
		{
			name:          "PrefixExpression",
			input:         "-5",
			expectedStmts: 1,
			validate: func(t *testing.T, stmt Statement) {
				exprStmt, ok := stmt.(*ExpressionStatement)
				if !ok {
					t.Fatalf("expected *ExpressionStatement, got %T", stmt)
				}
				prefixExp, ok := exprStmt.Expression.(*PrefixExpression)
				if !ok {
					t.Fatalf("expected *PrefixExpression, got %T", exprStmt.Expression)
				}
				if prefixExp.Operator != "-" {
					t.Fatalf("expected operator '-', got %q", prefixExp.Operator)
				}
				if prefixExp.Right.String() != "5" {
					t.Fatalf("expected right '5', got %q", prefixExp.Right.String())
				}
			},
		},
		{
			name:          "BooleanLiteral",
			input:         "true",
			expectedStmts: 1,
			validate: func(t *testing.T, stmt Statement) {
				exprStmt, ok := stmt.(*ExpressionStatement)
				if !ok {
					t.Fatalf("expected *ExpressionStatement, got %T", stmt)
				}
				boolLit, ok := exprStmt.Expression.(*BooleanLiteral)
				if !ok {
					t.Fatalf("expected *BooleanLiteral, got %T", exprStmt.Expression)
				}
				if !boolLit.Value {
					t.Fatalf("expected value true, got false")
				}
			},
		},
		{
			name:          "LetStatement",
			input:         "let x = 10;",
			expectedStmts: 1,
			validate: func(t *testing.T, stmt Statement) {
				letStmt, ok := stmt.(*LetStatement)
				if !ok {
					t.Fatalf("expected *LetStatement, got %T", stmt)
				}
				if letStmt.Name.Value != "x" {
					t.Fatalf("expected name 'x', got %q", letStmt.Name.Value)
				}
				intLit, ok := letStmt.Value.(*IntegerLiteral)
				if !ok {
					t.Fatalf("expected *IntegerLiteral, got %T", letStmt.Value)
				}
				if intLit.Value != 10 {
					t.Fatalf("expected value 10, got %d", intLit.Value)
				}
			},
		},
		{
			name:          "ReturnStatement",
			input:         "return 42",
			expectedStmts: 1,
			validate: func(t *testing.T, stmt Statement) {
				retStmt, ok := stmt.(*ReturnStatement)
				if !ok {
					t.Fatalf("expected *ReturnStatement, got %T", stmt)
				}
				intLit, ok := retStmt.ReturnValue.(*IntegerLiteral)
				if !ok {
					t.Fatalf("expected *IntegerLiteral, got %T", retStmt.ReturnValue)
				}
				if intLit.Value != 42 {
					t.Fatalf("expected value 42, got %d", intLit.Value)
				}
			},
		},
		{
			name:          "IfExpression",
			input:         "if (x > 5) { 10 } else { 20 }",
			expectedStmts: 1,
			validate: func(t *testing.T, stmt Statement) {
				exprStmt, ok := stmt.(*ExpressionStatement)
				if !ok {
					t.Fatalf("expected *ExpressionStatement, got %T", stmt)
				}
				ifExp, ok := exprStmt.Expression.(*IfExpression)
				if !ok {
					t.Fatalf("expected *IfExpression, got %T", exprStmt.Expression)
				}
				if ifExp.Condition == nil {
					t.Fatalf("expected condition, got nil")
				}
				if ifExp.Consequence == nil {
					t.Fatalf("expected consequence block, got nil")
				}
				if ifExp.Alternative == nil {
					t.Fatalf("expected alternative block, got nil")
				}
			},
		},
		{
			name:          "FunctionLiteral",
			input:         "fn(x, y) { x + y }",
			expectedStmts: 1,
			validate: func(t *testing.T, stmt Statement) {
				exprStmt, ok := stmt.(*ExpressionStatement)
				if !ok {
					t.Fatalf("expected *ExpressionStatement, got %T", stmt)
				}
				fnLit, ok := exprStmt.Expression.(*FunctionLiteral)
				if !ok {
					t.Fatalf("expected *FunctionLiteral, got %T", exprStmt.Expression)
				}
				if len(fnLit.Parameters) != 2 {
					t.Fatalf("expected 2 parameters, got %d", len(fnLit.Parameters))
				}
				if fnLit.Parameters[0].Value != "x" {
					t.Fatalf("expected first param 'x', got %q", fnLit.Parameters[0].Value)
				}
				if fnLit.Parameters[1].Value != "y" {
					t.Fatalf("expected second param 'y', got %q", fnLit.Parameters[1].Value)
				}
				if fnLit.Body == nil {
					t.Fatalf("expected function body, got nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := golexer.NewLexer(tt.input)
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if program == nil {
				t.Fatalf("Parse() returned nil")
			}
			if len(program.Statements) != tt.expectedStmts {
				t.Fatalf("expected %d statement(s), got %d", tt.expectedStmts, len(program.Statements))
			}

			tt.validate(t, program.Statements[0])
		})
	}
}
func TestParserErrors(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
	}{
		{"5 +", true},
		{"let x", true},
		{"if (x", true},
		{"fn(", true},
		{"(5 + 3", true},
		{"5 + 3", false},
		{"let x = 5;", false},
		{"if (true) { 5 }", false},
		{"fn() { 5 }", false},
		{"fn(x) { x + 1 }", false},
	}
	for _, tt := range tests {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		p.Parse()
		hasErrors := len(p.Errors()) > 0
		if hasErrors != tt.shouldError {
			if tt.shouldError {
				t.Fatalf("input=%q: expected errors but got none", tt.input)
			} else {
				t.Fatalf("input=%q: expected no errors but got: %v", tt.input, p.Errors())
			}
		}

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
		{"5 - 3", "(5 - 3)"},
		{"10 / 2 * 3", "((10 / 2) * 3)"},
		{"2 * 3 - 4", "((2 * 3) - 4)"},
		{"10 - 5 - 2", "((10 - 5) - 2)"},
		{"(10 - 5) * (2 + 3)", "((10 - 5) * (2 + 3))"},
		{"1 + 2 * 3 - 4", "((1 + (2 * 3)) - 4)"},
	}
	for _, tt := range tests {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestBooleanLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"true", "true"},
		{"false", "false"},
		{"!true", "(!true)"},
		{"!false", "(!false)"},
		{"!!true", "(!(!true))"},
		{"true", "true"},
		{"false", "false"},
		{"!true", "(!true)"},
		{"true == false", "(true == false)"},
		{"true != false", "(true != false)"},
	}
	for _, tt := range tests {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestIfExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"if (true) { 5 }", "iftrue{5}"},
		{"if (false) { 5 } else { 10 }", "iffalse{5}else{10}"},
		{"if (x > 5) { x + 1 }", "if(x > 5){(x + 1)}"},
		{"if (x < 10) { x * 2 } else { 0 }", "if(x < 10){(x * 2)}else{0}"},
		{"if (true) { 1 + 2 }", "iftrue{(1 + 2)}"},
		{"if (x == 5) { 100 }", "if(x == 5){100}"},
		{"if (x != y) { x + y }", "if(x != y){(x + y)}"},
		{"if (a > b) { a } else { b }", "if(a > b){a}else{b}"},
		{"if (5 > 3) { 5 > 2 }", "if(5 > 3){(5 > 2)}"},
		{"if (x > 5) { 1 } else if (x == 5) { 2 } else { 3 }", "if(x > 5){1}else{if(x == 5){2}else{3}}"},
		{"if (x > 5) { 1 } else if (x == 3) { 2 }", "if(x > 5){1}else{if(x == 3){2}}"},
		{"if (true) { 1 } else if (false) { 2 } else if (true) { 3 }", "iftrue{1}else{iffalse{2}else{iftrue{3}}}"},
		{"if (true) { if (false) { 1 } else { 2 } }", "iftrue{iffalse{1}else{2}}"},
	}
	for _, tt := range tests {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-5", "(-5)"},
		{"!true", "(!true)"},
		{"!false", "(!false)"},
		{"-5 + 3", "((-5) + 3)"},

		{"-x", "(-x)"},
		{"!x", "(!x)"},
		{"-1 * 2", "((-1) * 2)"},
		{"-(5 + 3)", "(-(5 + 3))"},
		{"!(x > 5)", "(!(x > 5))"},
	}
	for _, tt := range tests {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}
func TestFunctionLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fn() { }", "fn(){}"},
		{"fn(x) { x }", "fn(x){x}"},
		{"fn(x) { x + 1; }", "fn(x){(x + 1)}"},
		{"fn(x, y) { x + y; }", "fn(x, y){(x + y)}"},
		{"fn(a, b, c) { a * b + c; }", "fn(a, b, c){((a * b) + c)}"},
		{"fn(x) { x * 2 + 1 }", "fn(x){((x * 2) + 1)}"},
		{"let f = fn(x) { x }; f", "let f = fn(x){x};f"},
		{"fn(x) { fn(y) { x + y } }", "fn(x){fn(y){(x + y)}}"},
		{"fn(a, b) { a - b }", "fn(a, b){(a - b)}"},
		{"fn(n) { if (n > 0) { n } else { 0 } }", "fn(n){if(n > 0){n}else{0}}"},
	}
	for _, tt := range tests {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"return 5", "return 5;"},
		{"return 10", "return 10;"},
		{"return x", "return x;"},
		{"return x + y", "return (x + y);"},
		{"return x * 2 + 1", "return ((x * 2) + 1);"},
		{"return true", "return true;"},
		{"return false", "return false;"},
		{"return if (x > 5) { 10 } else { 20 }", "return if(x > 5){10}else{20};"},
		{"return fn(x) { x }", "return fn(x){x};"},
		{"return x && y", "return (x && y);"},
	}
	for _, tt := range tests {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x && y", "(x && y)"},
		{"x || y", "(x || y)"},
		{"true && false", "(true && false)"},
		{"true || false", "(true || false)"},
		{"x && y && z", "((x && y) && z)"},
		{"x || y || z", "((x || y) || z)"},
		{"x && y || z", "((x && y) || z)"},
		{"x || y && z", "(x || (y && z))"},
		{"(x && y) || (a && b)", "((x && y) || (a && b))"},
		{"!x && y", "((!x) && y)"},
		{"x && !y", "(x && (!y))"},
		{"x > 5 && y < 10", "((x > 5) && (y < 10))"},
		{"x == 5 || y == 10", "((x == 5) || (y == 10))"},
		{"if (x && y) { 1 } else { 2 }", "if(x && y){1}else{2}"},
		{"if (x || y) { 1 } else { 2 }", "if(x || y){1}else{2}"},
		{"fn(a, b) { a && b }", "fn(a, b){(a && b)}"},
		{"fn(a, b) { a || b }", "fn(a, b){(a || b)}"},
		{"let result = x && y;", "let result = (x && y);"},
		{"let result = x || y;", "let result = (x || y);"},
		{"return x && y", "return (x && y);"},
	}
	for _, tt := range tests {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestFunctionCalls(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// Basic function calls
		{"add()", "add()"},
		{"f()", "f()"},
		{"multiply()", "multiply()"},

		// Single argument
		{"add(5)", "add(5)"},
		{"sqrt(16)", "sqrt(16)"},
		{"f(x)", "f(x)"},
		{"abs(-5)", "abs((-5))"},

		// Multiple arguments
		{"add(2, 3)", "add(2, 3)"},
		{"multiply(4, 5)", "multiply(4, 5)"},
		{"power(2, 8)", "power(2, 8)"},
		{"min(10, 20, 30)", "min(10, 20, 30)"},

		// Arguments with expressions
		{"add(2 + 3, 4 * 5)", "add((2 + 3), (4 * 5))"},
		{"f(x + y, a - b)", "f((x + y), (a - b))"},
		{"max(x * 2, y / 3)", "max((x * 2), (y / 3))"},

		// Nested function calls
		{"f(g())", "f(g())"},
		{"add(mul(2, 3), 4)", "add(mul(2, 3), 4)"},
		{"f(g(h()))", "f(g(h()))"},
		{"add(mul(2, 3), div(10, 5))", "add(mul(2, 3), div(10, 5))"},

		// Function calls with identifiers and literals
		{"len(arr)", "len(arr)"},
		{"print(x)", "print(x)"},
		{"say(\"hello\")", "say(\"hello\")"},
		{"say(`raw string`)", "say(`raw string`)"},

		// Function calls with boolean arguments
		{"if_else(true, 1, 2)", "if_else(true, 1, 2)"},
		{"check(x > 5)", "check((x > 5))"},
		{"validate(x && y)", "validate((x && y))"},

		// Function calls in expressions
		{"add(2, 3) + 5", "(add(2, 3) + 5)"},
		{"mul(3, 4) * 2", "(mul(3, 4) * 2)"},
		{"f(x) + g(y)", "(f(x) + g(y))"},
		{"add(1, 2) == 3", "(add(1, 2) == 3)"},

		// Function calls in let statements
		{"let x = f();", "let x = f();"},
		{"let y = add(2, 3);", "let y = add(2, 3);"},
		{"let result = max(a, b, c);", "let result = max(a, b, c);"},

		// Function calls in return statements
		{"return f()", "return f();"},
		{"return add(1, 2)", "return add(1, 2);"},
		{"return mul(x, y)", "return mul(x, y);"},

		// Function calls in if conditions
		{"if (x && y) { 1 } else { 2 }", "if(x && y){1}else{2}"},
		{"if (x || y) { 1 } else { 2 }", "if(x || y){1}else{2}"},

		// Function calls in function bodies
		{"fn() { f() }", "fn(){f()}"},
		{"fn(x) { add(x, 1) }", "fn(x){add(x, 1)}"},
		{"fn(x, y) { mul(x, y) }", "fn(x, y){mul(x, y)}"},

		// Complex nested cases
		{"f(g(x), h(y))", "f(g(x), h(y))"},
		{"add(mul(2, 3), div(10, 5), sub(7, 2))", "add(mul(2, 3), div(10, 5), sub(7, 2))"},
		{"len(reverse(sort(arr)))", "len(reverse(sort(arr)))"},

		// Function calls with logical operators
		{"f(x && y)", "f((x && y))"},
		{"f(x || y)", "f((x || y))"},
		{"g(a && b, c || d)", "g((a && b), (c || d))"},

		// Function calls with prefix operators
		{"f(-x)", "f((-x))"},
		{"g(!x)", "g((!x))"},
		{"h(-a, !b)", "h((-a), (!b))"},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}
func TestArrayLiterals(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"[1, 2, 3]", "[1, 2, 3]"},
		{"[]", "[]"},
		{"[1 + 2, 3 * 4]", "[(1 + 2), (3 * 4)]"},
		{"[true, false, true]", "[true, false, true]"},
		{"[fn(x) { x }, fn(y) { y }]", "[fn(x){x}, fn(y){y}]"},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestModuloOperator(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// Basic modulo
		{"5 % 2", "(5 % 2)"},
		{"10 % 3", "(10 % 3)"},
		{"x % y", "(x % y)"},

		// Modulo with expressions
		{"a + b % c", "(a + (b % c))"},
		{"a % b + c", "((a % b) + c)"},
		{"a * b % c", "((a * b) % c)"},
		{"a % b * c", "((a % b) * c)"},

		// Modulo in comparison
		{"a % b == 0", "((a % b) == 0)"},
		{"x % 2 > 0", "((x % 2) > 0)"},

		// Modulo in array indexing
		{"arr[i % 5]", "(arr[(i % 5)])"},
		{"arr[0] % 2", "((arr[0]) % 2)"},

		// Modulo in function calls
		{"f(a % b)", "f((a % b))"},
		{"f(x % 2, y % 3)", "f((x % 2), (y % 3))"},

		// Chained modulo
		{"a % b % c", "((a % b) % c)"},

		// Complex expressions with modulo
		{"let x = a % b;", "let x = (a % b);"},
		{"return a % b;", "return (a % b);"},
		{"if (x % 2 == 0) { 1 }", "if((x % 2) == 0){1}"}}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestArrayIndexing(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// Basic indexing
		{"arr[0]", "(arr[0])"},
		{"arr[1]", "(arr[1])"},
		{"x[i]", "(x[i])"},

		// Index with expressions
		{"arr[i + 1]", "(arr[(i + 1)])"},
		{"arr[i - 1]", "(arr[(i - 1)])"},
		{"arr[i * 2]", "(arr[(i * 2)])"},
		{"arr[i / 2]", "(arr[(i / 2)])"},

		// Nested indexing
		{"arr[arr[0]]", "(arr[(arr[0])])"},
		{"matrix[i][j]", "((matrix[i])[j])"},

		// Indexing with function calls
		{"arr[f()]", "(arr[f()])"},
		{"arr[len(x)]", "(arr[len(x)])"},

		// Indexing in expressions
		{"arr[0] + arr[1]", "((arr[0]) + (arr[1]))"},
		{"arr[i] * 2", "((arr[i]) * 2)"},
		{"arr[0] == 5", "((arr[0]) == 5)"},

		// Indexing in let statements
		{"let x = arr[0];", "let x = (arr[0]);"},
		{"let y = arr[i + 1];", "let y = (arr[(i + 1)]);"},

		// Indexing in return statements
		{"return arr[0]", "return (arr[0]);"},
		{"return arr[i]", "return (arr[i]);"},

		// Indexing in if conditions
		{"if (arr[0] > 5) { 1 }", "if((arr[0]) > 5){1}"},
		{"if (arr[i] == x) { 1 } else { 2 }", "if((arr[i]) == x){1}else{2}"},

		// Indexing in function calls
		{"f(arr[0])", "f((arr[0]))"},
		{"f(arr[0], arr[1])", "f((arr[0]), (arr[1]))"},
		{"f(arr[i], g(arr[j]))", "f((arr[i]), g((arr[j])))"},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}
func TestForStatements(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		//todo make surre tests pass
		{"let i = 0; for i < 5 { i = i + 1; }", "let i = 0;for (i < 5) {(i = (i + 1))}"},
		{"let x = 10; for x > 0 { x = x - 1; }", "let x = 10;for (x > 0) {(x = (x - 1))}"},
		// Compound assignments in loops
		{"let i = 0; for i < 10 { i += 1; }", "let i = 0;for (i < 10) {(i += 1)}"},
		{"let x = 1; for x < 100 { x *= 2; }", "let x = 1;for (x < 100) {(x *= 2)}"},
		{"let x = 100; for x > 0 { x /= 2; }", "let x = 100;for (x > 0) {(x /= 2)}"},
		// Multiple statements in loop
		{"let i = 0; for i < 5 { i = i + 1; print(i); }", "let i = 0;for (i < 5) {(i = (i + 1));print(i)}"},
		{"let i = 0; for i < 3 { print(i); i = i + 1; }", "let i = 0;for (i < 3) {print(i);(i = (i + 1))}"},
		// Nested loops with assignment
		{"let i = 0; for i < 2 { let j = 0; for j < 2 { j = j + 1; } i = i + 1; }", "let i = 0;for (i < 2) {let j = 0;for (j < 2) {(j = (j + 1))};(i = (i + 1))}"},
		//Assignment with expressions
		{"let i = 0; for i < 10 { i = i + 2 * 3; }", "let i = 0;for (i < 10) {(i = (i + (2 * 3)))}"},
		{"let x = 1; for x < 100 { x = x + x; }", "let x = 1;for (x < 100) {(x = (x + x))}"},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}
func TestTableLiterals(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// Empty table
		{`table{}`, `table{}`},
		// Single string key
		{`table{"key": 1}`, `table{"key":1}`},
		// Single integer key
		{`table{1: "one"}`, `table{1:"one"}`},
		// Single boolean key
		{`table{true: 1}`, `table{true:1}`},
		// Multiple pairs - order now guaranteed
		{`table{"a": 1, "b": 2}`, `table{"a":1, "b":2}`},
		{`table{"a": 1, "b": 2, "c": 3}`, `table{"a":1, "b":2, "c":3}`},
		// Expression value
		{`table{"sum": 1 + 2}`, `table{"sum":(1 + 2)}`},
		// Expression key
		{`table{1 + 1: "two"}`, `table{(1 + 1):"two"}`},
		// Nested table as value
		{`table{"inner": table{"x": 1}}`, `table{"inner":table{"x":1}}`},
		// Nested table multiple pairs
		{`table{"a": 1, "inner": table{"x": 1, "y": 2}}`, `table{"a":1, "inner":table{"x":1, "y":2}}`},
		// Identifier value
		{`let x = 5; table{"val": x}`, `let x = 5;table{"val":x}`},
		// Identifier key
		{`let k = "key"; table{k: 42}`, `let k = "key";table{k:42}`},
		// Mixed key types
		{`table{"str": 1, 2: "int"}`, `table{"str":1, 2:"int"}`},
		// Boolean value
		{`table{"flag": true}`, `table{"flag":true}`},
		// Multiple mixed expression values
		{`table{"a": 1 + 2, "b": 3 * 4}`, `table{"a":(1 + 2), "b":(3 * 4)}`},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}
func TestForInStatements(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"for item in list { let x = item; }", "for item in list {let x = item;}"}, {"for item in getItems() { item; }", "for item in getItems() {item}"},
		{"for item in list { print(item); item; }", "for item in list {print(item);item}"},
		{"for i in list { for j in i { j; } }", "for i in list {for j in i {j}}"},
		{"for num in numbers { print(num); }", "for num in numbers {print(num)}"},
		{"for x in a + b { x; }", "for x in (a + b) {x}"},
		// test index for TestForInStatements
		{"for i, item in list { item; }", "for i, item in list {item}"},
		{"for i, item in getItems() { item; }", "for i, item in getItems() {item}"},
		{"for i, num in numbers { print(num); }", "for i, num in numbers {print(num)}"},
		{"for i, x in a + b { x; }", "for i, x in (a + b) {x}"},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}
func TestNullExpression(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"null", "null"},
		{"let x = null", "let x = null;"},
		{"let x = null; let y = null;", "let x = null;let y = null;"},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}
func TestArrowFunctions(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"fn(x) -> x", "fn(x){return x;}"},
		{"fn(x) -> x + 1", "fn(x){return (x + 1);}"},
		{"fn(x, y) -> x + y", "fn(x, y){return (x + y);}"},
		{"fn() -> 42", "fn(){return 42;}"},
		{"let add = fn(a, b) -> a + b", "let add = fn(a, b){return (a + b);};"},
		{"let double = fn(x) -> x * 2", "let double = fn(x){return (x * 2);};"},
		{"fn(x) -> print(x)", "fn(x){return print(x);}"},
		{"fn(x) -> x > 0", "fn(x){return (x > 0);}"},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}
func TestStringLiterals(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{`"hello"`, `"hello"`},
		{`"world"`, `"world"`},
		{`""`, `""`},
		{"let x = \"hello\";", `let x = "hello";`},
		{"let x = \"hello\"; let y = \"world\";", `let x = "hello";let y = "world";`},
		{"`raw string`", "`raw string`"},
		{"`hello world`", "`hello world`"},
		{"let x = `raw`;", "let x = `raw`;"},
		{`"hello" == "world"`, `("hello" == "world")`},
		{`fn(x) { x }`, `fn(x){x}`},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestFloatLiterals(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"3.14", "3.14"},
		{"0.5", "0.5"},
		{"2.0", "2.0"},
		{"let x = 3.14;", "let x = 3.14;"},
		{"3.14 + 1.0", "(3.14 + 1.0)"},
		{"3.14 * 2.0", "(3.14 * 2.0)"},
		{"3.14 > 2.0", "(3.14 > 2.0)"},
		{"return 3.14", "return 3.14;"},
		{"fn(x) { 3.14 }", "fn(x){3.14}"},
		{"let x = 1.5 + 2.5;", "let x = (1.5 + 2.5);"},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestNumberFormats(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"0xFF", "0xFF"},
		{"0x1A2B", "0x1A2B"},
		{"0b1010", "0b1010"},
		{"0B1111", "0B1111"},
		{"0o777", "0o777"},
		{"0755", "0755"},
		{"let hex = 0xFF;", "let hex = 0xFF;"},
		{"let bin = 0b1010;", "let bin = 0b1010;"},
		{"let oct = 0o777;", "let oct = 0o777;"},
		{"0xFF + 0b1010", "(0xFF + 0b1010)"},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestCompoundAssignments(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"x += 1", "(x += 1)"},
		{"x -= 1", "(x -= 1)"},
		{"x *= 2", "(x *= 2)"},
		{"x /= 2", "(x /= 2)"},
		{"x %= 3", "(x %= 3)"},
		{"x += 1 + 2", "(x += (1 + 2))"},
		{"x *= 2 + 3", "(x *= (2 + 3))"},
		{"let x = 5; x += 1", "let x = 5;(x += 1)"},
		{"let x = 10; x -= 3", "let x = 10;(x -= 3)"},
		{"let x = 2; x *= 4", "let x = 2;(x *= 4)"},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestSwitchStatements(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// default only
		{
			`switch x { default { 0 } }`,
			`switch x {default {0}}`,
		},
		// multiple cases
		{
			`switch x { case 1 { 10 } case 2 { 20 } }`,
			`switch x {case 1 {10}case 2 {20}}`,
		},
		// cases + default
		{
			`switch x { case 1 { 10 } case 2 { 20 } default { 0 } }`,
			`switch x {case 1 {10}case 2 {20}default {0}}`,
		},
		// switch on expression
		{
			`switch x + 1 { case 2 { 10 } default { 0 } }`,
			`switch (x + 1) {case 2 {10}default {0}}`,
		},
		// switch on string
		{
			`switch "hello" { case "hello" { "matched" } }`,
			`switch "hello" {case "hello" {"matched"}}`,
		},
		// break inside case
		{
			`switch x { case 1 { break } }`,
			`switch x {case 1 {break;}}`,
		},
		// nested switch
		{
			`switch x { case 1 { switch y { case 2 { 42 } } } }`,
			`switch x {case 1 {switch y {case 2 {42}}}}`,
		},
		// switch inside for
		{
			`for { switch x { case 1 { break } } }`,
			`for {switch x {case 1 {break;}}}`,
		},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}

func TestArrowFunctionAdvanced(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// Arrow returning arrow
		{"fn(x) -> fn(y) -> x + y", "fn(x){return fn(y){return (x + y);};}"},
		// Arrow in array
		{"[fn(x) -> x, fn(y) -> y * 2]", "[fn(x){return x;}, fn(y){return (y * 2);}]"},
		// Arrow as function argument
		{"map(fn(x) -> x * 2)", "map(fn(x){return (x * 2);})"},
		// Arrow with boolean body
		{"fn(x, y) -> x && y", "fn(x, y){return (x && y);}"},
		// Arrow with comparison
		{"fn(x) -> x == 0", "fn(x){return (x == 0);}"},
		// Arrow called immediately
		{"fn(x) -> x + 1", "fn(x){return (x + 1);}"},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		if len(p.Errors()) != 0 {
			t.Fatalf("input=%q: parser has %d errors: %v", tt.input, len(p.Errors()), p.Errors())
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if program.String() != tt.expected {
			t.Fatalf("input=%q: expected=%q, got=%q", tt.input, tt.expected, program.String())
		}
	}
}
func TestErrorRecovery(t *testing.T) {
	testCases := []struct {
		input         string
		expectedStmts int
		expectErrors  bool
	}{
		{"let x = ; let y = 5;", 2, true},
		{"let x = 5; let y = ; let z = 10;", 3, true},
		{"let x = ; let y = 10;", 2, true},
		{"5 + ; let x = 10;", 2, true},
		{"let x = ; let y = ; let z = 5;", 3, true},
		{"let x = ; let y = 5; let z = 10;", 3, true},
	}
	for _, tt := range testCases {
		l := golexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.Parse()
		hasErrors := len(p.Errors()) > 0
		if hasErrors != tt.expectErrors {
			if tt.expectErrors {
				t.Fatalf("input=%q: expected errors but got none", tt.input)
			} else {
				t.Fatalf("input=%q: expected no errors but got: %v", tt.input, p.Errors())
			}
		}
		if program == nil {
			t.Fatalf("input=%q: Parse() returned nil", tt.input)
		}
		if tt.expectedStmts > 0 {
			if len(program.Statements) != tt.expectedStmts {
				t.Fatalf("input=%q: expected %d statements after recovery, got %d", tt.input, tt.expectedStmts, len(program.Statements))
			}
		}
	}
}
func TestUse(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			`use "file.lgs"`,
			`use"file.lgs"`,
			"basic use",
		},
		{
			`use "math"`,
			`use"math"`,
			"use without extension",
		},
		{
			`use "lib/utils"`,
			`use"lib/utils"`,
			"use with path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			l := golexer.NewLexer(tc.input)
			p := NewParser(l)
			program := p.Parse()
			if len(p.Errors()) != 0 {
				t.Fatalf("input=%q: parser has %d errors: %v", tc.input, len(p.Errors()), p.Errors())
			}
			if program == nil {
				t.Fatalf("input=%q: Parse() returned nil", tc.input)
			}
			if program.String() != tc.expected {
				t.Fatalf("input=%q: expected=%q, got=%q", tc.input, tc.expected, program.String())
			}
		})
	}
}
func TestDotExpression(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "simple dot access",
			input:    `foo.bar`,
			expected: `(foo.bar)`,
		},
		{
			desc:     "chained dot access",
			input:    `foo.bar.baz`,
			expected: `((foo.bar).baz)`,
		},
		{
			desc:     "dot access on call",
			input:    `foo().bar`,
			expected: `(foo().bar)`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			lexer := golexer.NewLexer(tc.input)
			p := NewParser(lexer)
			program := p.Parse()
			if len(program.Statements) == 0 {
				t.Fatalf("[%s] no statements parsed", tc.desc)
			}
			expr, ok := program.Statements[0].(*ExpressionStatement)
			if !ok {
				t.Fatalf("[%s] expected ExpressionStatement", tc.desc)
			}
			if expr.Expression.String() != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, expr.Expression.String())
			}
		})
	}
}
func TestSpawnStatement(t *testing.T) {
	testCase := []struct {
		input  string
		expect string
		decs   string
	}{
		{"spawn {}", "spawn{}", "spawn empty block"},
		{"spawn { run() }", "spawn{run()}", "spawn single call"},
		{"spawn { httpGet(url) }", "spawn{httpGet(url)}", "spawn http call"},
		{"spawn { print(x) }", "spawn{print(x)}", "spawn print"},
		{"spawn { let x = 1 }", "spawn{let x = 1;}", "spawn let statement"},
		{"spawn { httpGet(url) fileCopy(src, dst) }", "spawn{httpGet(url);fileCopy(src, dst)}", "spawn multiple statements"},
		{"spawn { if x { print(x) } }", "spawn{ifx{print(x)}}", "spawn with if"},
		{"spawn { for i in items { print(i) } }", "spawn{for i in items {print(i)}}", "spawn with for in"},
		// nested spawn
		{"spawn { spawn { print(x) } }", "spawn{spawn{print(x)}}", "nested spawn"},

		// spawn with return
		{"spawn { return x }", "spawn{return x;}", "spawn with return"},

		// empty collection
		{"spawn for i in [] {}", "spawn for i in [] {}", "spawn for empty array"},
	}
	for _, tc := range testCase {
		t.Run(tc.decs, func(t *testing.T) {
			l := golexer.NewLexer(tc.input)
			p := NewParser(l)
			program := p.Parse()
			if len(p.Errors()) != 0 {
				t.Fatalf("input=%q: parser has %d errors: %v", tc.input, len(p.Errors()), p.Errors())
			}
			if program == nil {
				t.Fatalf("input=%q: Parse() returned nil", tc.input)
			}
			if program.String() != tc.expect {
				t.Fatalf("input=%q: expected=%q, got=%q", tc.input, tc.expect, program.String())
			}
		})
	}

}
func TestTenary(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"let x = true ? 1 : 2", "let x = (true ? 1 : 2);"},
		{"let x = false ? 1 : 2", "let x = (false ? 1 : 2);"},
		{"let x = a ? b : c", "let x = (a ? b : c);"},
		{"let x = a + b ? c : d", "let x = ((a + b) ? c : d);"},
		{"let x = a || b ? c : d", "let x = ((a || b) ? c : d);"},
		{"a ? b : c", "(a ? b : c)"},
		{"a ? b ? c : d : e", "(a ? (b ? c : d) : e)"},
		{"a ? b : c ? d : e", "(a ? b : (c ? d : e))"},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			l := golexer.NewLexer(tc.input)
			p := NewParser(l)
			program := p.Parse()
			if len(p.Errors()) != 0 {
				t.Fatalf("input=%q: parser has %d errors: %v", tc.input, len(p.Errors()), p.Errors())
			}
			if program == nil {
				t.Fatalf("input=%q: Parse() returned nil", tc.input)
			}
			if tc.expected != program.String() {
				t.Fatalf("input=%q: expected=%q, got=%q", tc.input, tc.expected, program.String())
			}
		})
	}
}
func TestIncrementDecrement(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"x++", "(x += 1)"},
		{"x--", "(x -= 1)"},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {

			l := golexer.NewLexer(tc.input)
			p := NewParser(l)
			program := p.Parse()
			if len(p.Errors()) != 0 {
				t.Fatalf("input=%q: parser has %d errors: %v", tc.input, len(p.Errors()), p.Errors())
			}
			if program == nil {
				t.Fatalf("input=%q: Parse() returned nil", tc.input)
			}
			if tc.expected != program.String() {
				t.Fatalf("input=%q: expected=%q, got=%q", tc.input, tc.expected, program.String())
			}
		})
	}
}
func TestPipe(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"x |> f", "f(x)"},
		{"x |> f |> g", "g(f(x))"},
		{"x |> f |> g |> h", "h(g(f(x)))"},
		{"x |> f(a) |> g(b) |> h(c)", "h(g(f(x, a), b), c)"},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			l := golexer.NewLexer(tc.input)
			p := NewParser(l)
			program := p.Parse()
			if len(p.Errors()) != 0 {
				t.Fatalf("input=%q: parser has %d errors: %v", tc.input, len(p.Errors()), p.Errors())
			}
			if program == nil {
				t.Fatalf("input=%q: Parse() returned nil", tc.input)
			}
			if tc.expected != program.String() {
				t.Fatalf("input=%q: expected=%q, got=%q", tc.input, tc.expected, program.String())
			}
		})
	}
}
func TestInterpolation(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"\"hello ${name}\"", "\"hello \"name"},
		{"\"result is ${1 + 2}\"", "\"result is \"(1 + 2)"},
		{"\"${a} and ${b}\"", "\"\"a\" and \"b"},
		{"\"hello world\"", "\"hello world\""},
		{"\"${name} is here\"", "\"\"name\" is here\""},
		{"\"${upper(name)}\"", "\"\"upper(name)"},
		{"\"${name}\"", "\"\"name"},
		{"\"${1 + 2 + 3}\"", "\"\"((1 + 2) + 3)"},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			l := golexer.NewLexer(tc.input)
			p := NewParser(l)
			program := p.Parse()
			if len(p.Errors()) != 0 {
				t.Fatalf("input=%q: parser has %d errors: %v", tc.input, len(p.Errors()), p.Errors())
			}
			if program == nil {
				t.Fatalf("input=%q: Parse() returned nil", tc.input)
			}
			if tc.expected != program.String() {
				t.Fatalf("input=%q: expected=%q, got=%q", tc.input, tc.expected, program.String())
			}
		})
	}
}
func TestTryExp(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"let result = try riskyFunction()", "let result = try riskyFunction();"},
		{"try riskyFunction()", "try riskyFunction()"},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			l := golexer.NewLexer(tc.input)
			p := NewParser(l)
			program := p.Parse()
			if len(p.Errors()) != 0 {
				t.Fatalf("input=%q: parser has %d errors: %v", tc.input, len(p.Errors()), p.Errors())
			}
			if program == nil {
				t.Fatalf("input=%q: Parse() returned nil", tc.input)
			}
			if tc.expected != program.String() {
				t.Fatalf("input=%q: expected=%q, got=%q", tc.input, tc.expected, program.String())
			}
		})
	}
}
func TestConstStatement(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		isConst  bool
	}{
		{"const x = 10", "const x = 10;", true},
		{"let y = 10", "let y = 10;", false},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			l := golexer.NewLexer(tc.input)
			p := NewParser(l)
			program := p.Parse()
			if len(p.Errors()) != 0 {
				t.Fatalf("input=%q: parser has %d errors: %v", tc.input, len(p.Errors()), p.Errors())
			}
			if program == nil {
				t.Fatalf("input=%q: Parse() returned nil", tc.input)
			}
			if tc.expected != program.String() {
				t.Fatalf("input=%q: expected=%q, got=%q", tc.input, tc.expected, program.String())
			}
			stmt, ok := program.Statements[0].(*LetStatement)
			if !ok {
				t.Fatalf("input=%q: expected *ast.LetStatement, got %T", tc.input, program.Statements[0])
			}
			if stmt.IsConst != tc.isConst {
				t.Fatalf("input=%q: expected IsConst=%v, got %v", tc.input, tc.isConst, stmt.IsConst)
			}
		})
	}
}
