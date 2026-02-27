package interpreter

import (
	"strings"
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
		{"x + 5;", "ERROR [1:1]: identifier not found: x", "error bubbles up"},
		{"5 / 0;", "ERROR [1:3]: division by zero", "division by zero error"},
		{"5 % 0;", "ERROR [1:3]: modulo by zero", "modulo by zero error"},
		{"5 + true;", "ERROR [1:3]: type mismatch: INTEGER + BOOLEAN", "type mismatch error"},
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
		{"5.0 / 0.0", "ERROR [1:5]: division by zero", "float division by zero"},
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
		{`"hello" - "world"`, "ERROR [1:9]: unknown operator: STRING - STRING", "invalid string operator"},
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
		{"10 % 0", "ERROR [1:4]: modulo by zero", "modulo by zero"},
		{"10 / 0", "ERROR [1:4]: division by zero", "division by zero"},
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
		{"if true { 10 }", "10", "true condition"},
		{"if false { 10 }", "null", "false condition no alternative returns null"},
		{"if 1 { 10 }", "10", "integer is truthy"},
		{"if null { 10 }", "null", "null is falsy"},

		// with else
		{"if true { 10 } else { 20 }", "10", "true takes consequence"},
		{"if false { 10 } else { 20 }", "20", "false takes alternative"},

		// condition is expression
		{"if 5 > 3 { 10 }", "10", "expression condition true"},
		{"if 5 < 3 { 10 } else { 20 }", "20", "expression condition false"},
		{"if 5 == 5 { 10 }", "10", "equality condition"},

		// condition uses variables
		{"let x = 5; if x > 3 { 10 }", "10", "variable in condition"},
		{"let x = 1; if x > 3 { 10 } else { 20 }", "20", "variable condition false"},

		// block returns last value
		{"if true { 5; 10; 15; }", "15", "block returns last statement"},
		{"if false { 5 } else { 10; 20; }", "20", "else block returns last statement"},

		// early return inside block (stays wrapped)
		{"if true { return 5; 10; }", "5", "return inside if stops block"},
		{"if false { 10 } else { return 20; 30; }", "20", "return inside else stops block"},

		// error in condition bubbles up
		{"if x { 10 }", "ERROR [1:4]: identifier not found: x", "error in condition bubbles up"},
		{"if 5 + true { 10 }", "ERROR [1:6]: type mismatch: INTEGER + BOOLEAN", "type mismatch in condition"},

		// nested if
		{"if true { if true { 10 } }", "10", "nested if"},
		{"if true { if false { 10 } else { 20 } }", "20", "nested if else"},
		{"if true { if false { 10 } }", "null", "nested if no alternative null"},
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
func TestFunctionLiteral(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"fn(x) { x }", "fn(x) {\nx\n}", "function literal returns function object"},
		{"fn(x, y) { x + y }", "fn(x, y) {\n(x + y)\n}", "multi param function literal"},
		{"fn() { 5 }", "fn() {\n5\n}", "no param function literal"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			lexer := golexer.NewLexer(tc.input)
			p := parser.NewParser(lexer)
			program := p.Parse()
			i := NewInterpreter()
			result := i.Eval(program, i.Env)

			if result.Type() != FUNCTION_OBJ {
				t.Errorf("[%s] expected FUNCTION_OBJ, got %s", tc.desc, result.Type())
			}
		})
	}
}

func TestEvalFunctionCalls(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"let add = fn(x, y) { x + y }; add(5, 10)", "15", "basic two param call"},
		{"let double = fn(x) { x * 2 }; double(5)", "10", "single param call"},
		{"let identity = fn(x) { x }; identity(5)", "5", "identity function"},
		{"let noParam = fn() { 5 }; noParam()", "5", "no param call"},
		{"let f = fn(x) { 1; 2; x }; f(5)", "5", "implicit return is last value"},
		{"let f = fn(x) { return x; 999 }; f(5)", "5", "explicit return exits early"},
		{"notAFunction(5)", "ERROR [1:1]: identifier not found: notAFunction", "undefined function"},
		{"let x = 5; x(5)", "ERROR: not a function: INTEGER", "call non function"},
		{"true(5)", "ERROR: not a function: BOOLEAN", "call boolean as function"},
		{"let add = fn(x, y) { x + y }; add(x, 5)", "ERROR [1:35]: identifier not found: x", "error in first arg"},
		{"let add = fn(x, y) { x + y }; add(5, x)", "ERROR [1:38]: identifier not found: x", "error in second arg"},
		{"let x = 10; let f = fn(x) { x }; f(5); x", "10", "function param doesnt leak to outer env"},
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

func TestApplyFunction(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			`let factorial = fn(n) {
				if n == 0 { return 1 }
				n * factorial(n - 1)
			};
			factorial(5)`,
			"120",
			"recursive factorial",
		},
		{
			`let fib = fn(n) {
				if n <= 1 { return n }
				fib(n - 1) + fib(n - 2)
			};
			fib(10)`,
			"55",
			"recursive fibonacci",
		},
		{
			`let apply = fn(f, x) { f(x) };
			let double = fn(x) { x * 2 };
			apply(double, 5)`,
			"10",
			"function as argument",
		},
		{
			`let makeAdder = fn(x) { fn(y) { x + y } };
			let add5 = makeAdder(5);
			add5(3)`,
			"8",
			"function as return value closure",
		},
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
func TestBangOperator(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"!true", "false", "bang true"},
		{"!false", "true", "bang false"},
		{"!null", "true", "bang null is true"},
		{"!5", "false", "bang integer is false"},
		{"!0", "false", "bang zero is false"},
		{"!!true", "true", "double bang true"},
		{"!!false", "false", "double bang false"},
		{"!!5", "true", "double bang integer"},
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

func TestMinusPrefixOperator(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		// integer
		{"-5", "-5", "negate positive integer"},
		{"-10", "-10", "negate ten"},
		{"--5", "5", "double negate integer"},
		{"-0", "0", "negate zero"},

		// float
		{"-5.0", "-5", "negate positive float"},
		{"-10.5", "-10.5", "negate float with decimal"},
		{"--5.0", "5", "double negate float"},
		{"-0.0", "-0", "negate zero float"},

		// errors
		{"-true", "ERROR [1:1]: unknown operator: -BOOLEAN", "negate boolean"},
		{"-null", "ERROR [1:1]: unknown operator: -NULL", "negate null"},
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
func TestArrayLiteral(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		// basic arrays
		{"[1, 2, 3]", "[1, 2, 3]", "integer array"},
		{"[1.0, 2.0, 3.0]", "[1, 2, 3]", "float array"},
		{`["hello", "world"]`, "[hello, world]", "string array"},
		{"[true, false]", "[true, false]", "boolean array"},
		{"[]", "[]", "empty array"},

		// mixed types
		{`[1, "hello", true]`, `[1, hello, true]`, "mixed type array"},

		// expressions as elements
		{"[1 + 2, 3 * 4]", "[3, 12]", "expression elements"},
		{"let x = 5; [x, x + 1]", "[5, 6]", "variable elements"},

		// error in element bubbles up
		{"[1, x, 3]", "ERROR [1:5]: identifier not found: x", "error in element bubbles up"},

		// nested arrays
		{"[[1, 2], [3, 4]]", "[[1, 2], [3, 4]]", "nested arrays"},
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
func TestAndOr(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		// &&
		{"true && true", "true", "true and true"},
		{"true && false", "false", "true and false"},
		{"false && true", "false", "false and true"},
		{"false && false", "false", "false and false"},

		// ||
		{"true || true", "true", "true or true"},
		{"true || false", "true", "true or false"},
		{"false || true", "true", "false or true"},
		{"false || false", "false", "false or false"},

		// short circuit &&
		{"false && 1", "false", "short circuit and does not eval right"},
		{"true && true", "true", "and evals right when left is true"},

		// short circuit ||
		{"true || 1", "true", "short circuit or does not eval right"},
		{"false || true", "true", "or evals right when left is false"},

		// with expressions
		{"1 == 1 && 2 == 2", "true", "expression and expression"},
		{"1 == 2 && 2 == 2", "false", "false expression and true expression"},
		{"1 == 2 || 2 == 2", "true", "false expression or true expression"},
		{"1 == 2 || 2 == 3", "false", "false expression or false expression"},

		// chained
		{"true && true && true", "true", "chained and"},
		{"true && false && true", "false", "chained and with false"},
		{"false || false || true", "true", "chained or"},
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
func TestArrayIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"let arr = [1, 2, 3]; arr[0]", int64(1)},
		{"let arr = [1, 2, 3]; arr[1]", int64(2)},
		{"let arr = [1, 2, 3]; arr[2]", int64(3)},
		{"let arr = [1, 2, 3]; arr[0] + arr[1]", int64(3)},
		{"let arr = [1, 2, 3]; arr[1] * arr[2]", int64(6)},
		{"let i = 0; let arr = [1, 2, 3]; arr[i]", int64(1)},
		{"let arr = [10, 20, 30]; arr[2] - arr[0]", int64(20)},
		{"let arr = [1, 2, 3]; arr[-1]", "index out of bounds: index -1, length 3"},
		{"let arr = [1, 2, 3]; arr[3]", "index out of bounds: index 3, length 3"},
		{"let arr = [1, 2, 3]; arr[100]", "index out of bounds: index 100, length 3"},
		{"let arr = [\"a\", \"b\", \"c\"]; arr[0]", "a"},
		{"let arr = [true, false, true]; arr[1]", false},
		{"let arr = [1, 2, 3]; let i = 2; arr[i]", int64(3)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := golexer.NewLexerWithConfig(tt.input, "tokens.json")
			p := parser.NewParser(lexer)
			program := p.Parse()
			inter := NewInterpreter()
			result := inter.Eval(program, inter.Env)

			switch expected := tt.expected.(type) {
			case int64:
				intObj, ok := result.(*Integar)
				if !ok {
					t.Fatalf("expected INTEGER, got %T (%s)", result, result.String())
				}
				if intObj.Value != expected {
					t.Errorf("expected %d, got %d", expected, intObj.Value)
				}
			case string:
				errObj, ok := result.(*Error)
				if ok {
					if !strings.Contains(errObj.Message, expected) {
						t.Errorf("expected error containing %q, got %q", expected, errObj.Message)
					}
					return
				}
				strObj, ok := result.(*String)
				if !ok {
					t.Fatalf("expected STRING, got %T (%s)", result, result.String())
				}
				if strObj.Value != expected {
					t.Errorf("expected %q, got %q", expected, strObj.Value)
				}
			case bool:
				boolObj, ok := result.(*Bool)
				if !ok {
					t.Fatalf("expected BOOLEAN, got %T (%s)", result, result.String())
				}
				if boolObj.Value != expected {
					t.Errorf("expected %t, got %t", expected, boolObj.Value)
				}
			}
		})
	}
}
