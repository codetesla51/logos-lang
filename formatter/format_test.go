package formatter

import (
	"testing"

	"github.com/codetesla51/golexer/golexer"
	"github.com/codetesla51/logos/parser"
)

func format(input string) string {
	lexer := golexer.NewLexerWithConfig(input, "../tokens.json")
	p := parser.NewParser(lexer)
	program := p.Parse()
	f := New()
	return f.Format(program)
}

func TestFormatLetStatement(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"let x = 5", "let x = 5\n", "basic let"},
		{"let x = 5 + 3", "let x = 5 + 3\n", "let with expression"},
		{"let x = true", "let x = true\n", "let bool"},
		{"let x = null", "let x = null\n", "let null"},
		{"let x = \"hello\"", "let x = \"hello\"\n", "let string"},
		{"let x = `raw`", "let x = `raw`\n", "let raw string"},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}

func TestFormatReturnStatement(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"return 5", "return 5\n", "return int"},
		{"return true", "return true\n", "return bool"},
		{"return x + 1", "return x + 1\n", "return expression"},
		{"return null", "return null\n", "return null"},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}

func TestFormatIfExpression(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			"if true { 5 }",
			"if true {\n    5\n}\n",
			"basic if",
		},
		{
			"if true { 5 } else { 10 }",
			"if true {\n    5\n} else {\n    10\n}\n",
			"if else",
		},
		{
			"if x > 5 { 10 }",
			"if x > 5 {\n    10\n}\n",
			"if with condition",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}

func TestFormatFunctionLiteral(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			"fn() { 5 }",
			"fn() {\n    5\n}\n",
			"no params",
		},
		{
			"fn(x) { x }",
			"fn(x) {\n    x\n}\n",
			"one param",
		},
		{
			"fn(x, y) { x + y }",
			"fn(x, y) {\n    x + y\n}\n",
			"two params",
		},
		{
			"let add = fn(x, y) { x + y }",
			"let add = fn(x, y) {\n    x + y\n}\n",
			"named function",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}

func TestFormatForLoop(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			"for { break }",
			"for {\n    break\n}\n",
			"infinite loop",
		},
		{
			"for x < 5 { x += 1 }",
			"for x < 5 {\n    x += 1\n}\n",
			"condition loop",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}

func TestFormatForIn(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			"for item in arr { print(item) }",
			"for item in arr {\n    print(item)\n}\n",
			"basic for in",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}

func TestFormatSwitch(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			"switch x { case 1 { 10 } }",
			"switch x {\n    case 1 {\n        10\n    }\n}\n",
			"basic switch",
		},
		{
			"switch x { case 1 { 10 } default { 0 } }",
			"switch x {\n    case 1 {\n        10\n    }\n    default {\n        0\n    }\n}\n",
			"switch with default",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}

func TestFormatArray(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"[]", "[]\n", "empty array"},
		{"[1, 2, 3]", "[1, 2, 3]\n", "int array"},
		{"[1, 2, 3]", "[1, 2, 3]\n", "basic array"},
		{"[\"a\", \"b\"]", "[\"a\", \"b\"]\n", "string array"},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}

func TestFormatTable(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			"table{}",
			"table{}\n",
			"empty table",
		},
		{
			"table{\"x\": 1}",
			"table{\n    \"x\": 1,\n}\n",
			"single pair",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}

func TestFormatDotExpression(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"foo.bar", "foo.bar\n", "basic dot"},
		{"foo.bar.baz", "foo.bar.baz\n", "chained dot"},
		{"res.ok", "res.ok\n", "result dot"},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}

func TestFormatUse(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"use \"std/math\"", "use \"std/math\"\n", "basic use"},
		{"use \"std/array\"", "use \"std/array\"\n", "use array"},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}

func TestFormatIndentation(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			"fn(x) { if x > 0 { return x } }",
			"fn(x) {\n    if x > 0 {\n        return x\n    }\n}\n",
			"nested indentation",
		},
		{
			"for x < 5 { if x > 3 { break } }",
			"for x < 5 {\n    if x > 3 {\n        break\n    }\n}\n",
			"nested for if",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := format(tc.input)
			if got != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, got)
			}
		})
	}
}
