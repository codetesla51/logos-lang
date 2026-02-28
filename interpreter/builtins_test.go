package interpreter

import (
	"fmt"
	"os"
	"testing"

	"github.com/codetesla51/golexer/golexer"
	"github.com/codetesla51/logos/parser"
)

func TestPrint(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "print string",
			input:    `print("hello")`,
			expected: "null",
		},
		{
			desc:     "print number",
			input:    `print(42)`,
			expected: "null",
		},
		{
			desc:     "print multiple args",
			input:    `print("hello", "world")`,
			expected: "null",
		},
		{
			desc:     "print bool",
			input:    `print(true)`,
			expected: "null",
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

func TestInput(t *testing.T) {
	testCases := []struct {
		desc      string
		input     string
		userInput string
		expected  string
	}{
		{
			desc:      "input no prompt",
			input:     `input()`,
			userInput: "hello\n",
			expected:  "hello",
		},
		{
			desc:      "input with prompt",
			input:     `input("name: ")`,
			userInput: "john\n",
			expected:  "john",
		},
		{
			desc:      "input with spaces",
			input:     `input()`,
			userInput: "hello world\n",
			expected:  "hello world",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// swap stdin with a pipe
			old := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			w.WriteString(tc.userInput)
			w.Close()

			lexer := golexer.NewLexer(tc.input)
			p := parser.NewParser(lexer)
			program := p.Parse()
			i := NewInterpreter()
			result := i.Eval(program, i.Env)

			os.Stdin = old

			if result.String() != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, result.String())
			}
		})
	}
}
func TestFileBuiltins(t *testing.T) {
	// setup temp dir
	tmpDir, _ := os.MkdirTemp("", "lgs-test-*")
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		desc     string
		input    string
		expected string
		setup    func()
	}{
		// write
		{
			desc:     "write creates file",
			input:    fmt.Sprintf(`write("%s/test.txt", "hello")`, tmpDir),
			expected: "null",
		},
		// read
		{
			desc:     "read existing file",
			input:    fmt.Sprintf(`read("%s/test.txt")`, tmpDir),
			expected: "hello",
			setup: func() {
				os.WriteFile(tmpDir+"/test.txt", []byte("hello"), 0644)
			},
		},
		{
			desc:     "read nonexistent file returns null",
			input:    fmt.Sprintf(`read("%s/nope.txt")`, tmpDir),
			expected: "null",
		},
		// append
		{
			desc:     "append to file",
			input:    fmt.Sprintf(`append("%s/append.txt", "world")`, tmpDir),
			expected: "null",
			setup: func() {
				os.WriteFile(tmpDir+"/append.txt", []byte("hello "), 0644)
			},
		},
		{
			desc:     "read appended file",
			input:    fmt.Sprintf(`read("%s/append.txt")`, tmpDir),
			expected: "hello world",
			setup: func() {
				os.WriteFile(tmpDir+"/append.txt", []byte("hello world"), 0644)
			},
		},
		// exists
		{
			desc:     "exists returns true for existing file",
			input:    fmt.Sprintf(`exists("%s/exists.txt")`, tmpDir),
			expected: "true",
			setup: func() {
				os.WriteFile(tmpDir+"/exists.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "exists returns false for missing file",
			input:    fmt.Sprintf(`exists("%s/missing.txt")`, tmpDir),
			expected: "false",
		},
		// delete
		{
			desc:     "delete removes file",
			input:    fmt.Sprintf(`delete("%s/del.txt")`, tmpDir),
			expected: "null",
			setup: func() {
				os.WriteFile(tmpDir+"/del.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "delete nonexistent returns null",
			input:    fmt.Sprintf(`delete("%s/nope.txt")`, tmpDir),
			expected: "null",
		},
		// deleteAll
		{
			desc:     "deleteAll removes folder",
			input:    fmt.Sprintf(`deleteAll("%s/subdir")`, tmpDir),
			expected: "null",
			setup: func() {
				os.MkdirAll(tmpDir+"/subdir/nested", 0755)
				os.WriteFile(tmpDir+"/subdir/nested/file.txt", []byte("hi"), 0644)
			},
		},
		// rename
		{
			desc:     "rename file",
			input:    fmt.Sprintf(`rename("%s/old.txt", "%s/new.txt")`, tmpDir, tmpDir),
			expected: "null",
			setup: func() {
				os.WriteFile(tmpDir+"/old.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "rename nonexistent returns null",
			input:    fmt.Sprintf(`rename("%s/nope.txt", "%s/nope2.txt")`, tmpDir, tmpDir),
			expected: "null",
		},
		// mkdir
		{
			desc:     "mkdir creates folder",
			input:    fmt.Sprintf(`mkdir("%s/newdir")`, tmpDir),
			expected: "null",
		},
		{
			desc:     "mkdir creates nested folders",
			input:    fmt.Sprintf(`mkdir("%s/a/b/c")`, tmpDir),
			expected: "null",
		},
		// rmdir
		{
			desc:     "rmdir removes empty folder",
			input:    fmt.Sprintf(`rmdir("%s/emptydir")`, tmpDir),
			expected: "null",
			setup: func() {
				os.Mkdir(tmpDir+"/emptydir", 0755)
			},
		},
		{
			desc:     "rmdir nonexistent returns null",
			input:    fmt.Sprintf(`rmdir("%s/nope")`, tmpDir),
			expected: "null",
		},
		// cp
		{
			desc:     "cp copies file",
			input:    fmt.Sprintf(`cp("%s/src.txt", "%s/dst.txt")`, tmpDir, tmpDir),
			expected: "null",
			setup: func() {
				os.WriteFile(tmpDir+"/src.txt", []byte("copied"), 0644)
			},
		},
		{
			desc:     "cp nonexistent returns null",
			input:    fmt.Sprintf(`cp("%s/nope.txt", "%s/dst.txt")`, tmpDir, tmpDir),
			expected: "null",
		},
		// mv
		// {
		// 	desc:     "mv moves file",
		// 	input:    fmt
		//		.Sprintf(`mv("%s/mv_src.txt", "%s/mv_dst.txt")`, tmpDir, tmpDir),
		// 	expected: "null",
		// 	setup: func() {
		// 		os.Remove(tmpDir + "/mv_dst.txt")
		// 		os.WriteFile(tmpDir+"/mv_src.txt", []byte("moved"), 0644)
		// 	},
		// },
		// chmod
		{
			desc:     "chmod changes permissions",
			input:    fmt.Sprintf(`chmod("%s/chmod.txt", "0644")`, tmpDir),
			expected: "null",
			setup: func() {
				os.WriteFile(tmpDir+"/chmod.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "chmod invalid mode returns error",
			input:    fmt.Sprintf(`chmod("%s/chmod.txt", "badmode")`, tmpDir),
			expected: `ERROR: chmod() invalid mode "badmode", use octal like "0755"`,
		},
		// glob
		{
			desc:     "glob finds files",
			input:    fmt.Sprintf(`len(glob("%s/*.txt"))`, tmpDir),
			expected: "1",
			setup: func() {
				// clean tmpDir first then add one txt file
				os.RemoveAll(tmpDir)
				os.MkdirTemp("", "")
				os.MkdirAll(tmpDir, 0755)
				os.WriteFile(tmpDir+"/only.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "glob no matches returns empty array",
			input:    fmt.Sprintf(`len(glob("%s/*.xyz"))`, tmpDir),
			expected: "0",
		},
		// readDir
		{
			desc:     "readDir returns file names",
			input:    fmt.Sprintf(`len(readDir("%s"))`, tmpDir),
			expected: "1",
			setup: func() {
				os.RemoveAll(tmpDir)
				os.MkdirAll(tmpDir, 0755)
				os.WriteFile(tmpDir+"/file.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "readDir nonexistent returns null",
			input:    fmt.Sprintf(`readDir("%s/nope")`, tmpDir),
			expected: "null",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}
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
func TestStringBuiltins(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		// len
		{desc: "len of string", input: `len("hello")`, expected: "5"},
		{desc: "len of empty string", input: `len("")`, expected: "0"},
		{desc: "len of array", input: `len([1, 2, 3])`, expected: "3"},

		// upper
		{desc: "upper", input: `upper("hello")`, expected: "HELLO"},
		{desc: "upper already upper", input: `upper("HELLO")`, expected: "HELLO"},

		// lower
		{desc: "lower", input: `lower("HELLO")`, expected: "hello"},
		{desc: "lower already lower", input: `lower("hello")`, expected: "hello"},

		// trim
		{desc: "trim spaces", input: `trim("  hello  ")`, expected: "hello"},
		{desc: "trim no spaces", input: `trim("hello")`, expected: "hello"},

		// replace
		{desc: "replace word", input: `replace("hello world", "world", "there")`, expected: "hello there"},
		{desc: "replace all occurrences", input: `replace("aaa", "a", "b")`, expected: "bbb"},
		{desc: "replace no match", input: `replace("hello", "x", "y")`, expected: "hello"},

		// split
		{desc: "split by comma", input: `len(split("a,b,c", ","))`, expected: "3"},
		{desc: "split by space", input: `len(split("hello world", " "))`, expected: "2"},

		// join
		{desc: "join array", input: `join(["a", "b", "c"], "-")`, expected: "a-b-c"},
		{desc: "join with space", input: `join(["hello", "world"], " ")`, expected: "hello world"},

		// contains
		{desc: "contains true", input: `contains("hello world", "world")`, expected: "true"},
		{desc: "contains false", input: `contains("hello world", "xyz")`, expected: "false"},
		{desc: "contains array true", input: `contains([1, 2, 3], 2)`, expected: "true"},
		{desc: "contains array false", input: `contains([1, 2, 3], 9)`, expected: "false"},

		// startsWith
		{desc: "startsWith true", input: `startsWith("hello", "he")`, expected: "true"},
		{desc: "startsWith false", input: `startsWith("hello", "lo")`, expected: "false"},

		// endsWith
		{desc: "endsWith true", input: `endsWith("hello", "lo")`, expected: "true"},
		{desc: "endsWith false", input: `endsWith("hello", "he")`, expected: "false"},

		// indexOf
		{desc: "indexOf found", input: `indexOf("hello", "l")`, expected: "2"},
		{desc: "indexOf not found", input: `indexOf("hello", "x")`, expected: "-1"},

		// repeat
		{desc: "repeat string", input: `repeat("ha", 3)`, expected: "hahaha"},
		{desc: "repeat once", input: `repeat("hi", 1)`, expected: "hi"},
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
func TestTypeConversionBuiltins(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		// toInt
		{desc: "toInt from string", input: `toInt("42")`, expected: "42"},
		{desc: "toInt from float", input: `toInt(3.9)`, expected: "3"},
		{desc: "toInt from bool true", input: `toInt(true)`, expected: "1"},
		{desc: "toInt from bool false", input: `toInt(false)`, expected: "0"},
		{desc: "toInt from int", input: `toInt(42)`, expected: "42"},
		{desc: "toInt invalid string returns null", input: `toInt("abc")`, expected: "null"},

		// toFloat
		{desc: "toFloat from string", input: `toFloat("3.14")`, expected: "3.14"},
		{desc: "toFloat from int", input: `toFloat(1)`, expected: "1"},
		{desc: "toFloat from bool true", input: `toFloat(true)`, expected: "1"},
		{desc: "toFloat from bool false", input: `toFloat(false)`, expected: "0"},
		{desc: "toFloat from float", input: `toFloat(3.14)`, expected: "3.14"},
		{desc: "toFloat invalid string returns null", input: `toFloat("abc")`, expected: "null"},

		// toBool
		{desc: "toBool from int 1", input: `toBool(1)`, expected: "true"},
		{desc: "toBool from int 0", input: `toBool(0)`, expected: "false"},
		{desc: "toBool from float", input: `toBool(1.0)`, expected: "true"},
		{desc: "toBool from string true", input: `toBool("true")`, expected: "true"},
		{desc: "toBool from string false", input: `toBool("false")`, expected: "false"},
		{desc: "toBool from bool", input: `toBool(true)`, expected: "true"},
		{desc: "toBool invalid string returns null", input: `toBool("abc")`, expected: "null"},

		// toStr
		{desc: "toStr from int", input: `toStr(42)`, expected: "42"},
		{desc: "toStr from float", input: `toStr(3.14)`, expected: "3.14"},
		{desc: "toStr from bool", input: `toStr(true)`, expected: "true"},
		{desc: "toStr from string", input: `toStr("hello")`, expected: "hello"},
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
func TestArrayBuiltins(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		// push
		{desc: "push adds element", input: `len(push([1, 2, 3], 4))`, expected: "4"},
		{desc: "push result has correct len", input: `len(push([1, 2, 3], 4))`, expected: "4"},
		{desc: "push to empty array", input: `len(push([], 1))`, expected: "1"},
		// pop
		{desc: "pop returns last element", input: `pop([1, 2, 3])`, expected: "3"},
		{desc: "pop from single element", input: `pop([42])`, expected: "42"},
		{desc: "pop empty array returns null", input: `pop([])`, expected: "null"},

		// first
		{desc: "first returns first element", input: `first([1, 2, 3])`, expected: "1"},
		{desc: "first single element", input: `first([42])`, expected: "42"},
		{desc: "first empty array returns null", input: `first([])`, expected: "null"},

		// last
		{desc: "last returns last element", input: `last([1, 2, 3])`, expected: "3"},
		{desc: "last single element", input: `last([42])`, expected: "42"},
		{desc: "last empty array returns null", input: `last([])`, expected: "null"},

		// tail
		{desc: "tail returns all but first", input: `len(tail([1, 2, 3]))`, expected: "2"},
		{desc: "tail single element returns empty", input: `len(tail([1]))`, expected: "0"},
		{desc: "tail empty array returns null", input: `tail([])`, expected: "null"},

		// prepend
		{desc: "prepend adds to front", input: `len(prepend([1, 2, 3], 0))`, expected: "4"},
		{desc: "prepend to empty", input: `len(prepend([], 1))`, expected: "1"},
		{desc: "prepend correct element", input: `first(prepend([1, 2, 3], 0))`, expected: "0"},

		// reverse
		{desc: "reverse array", input: `first(reverse([1, 2, 3]))`, expected: "3"},
		{desc: "reverse single element", input: `first(reverse([1]))`, expected: "1"},
		{desc: "reverse empty array", input: `len(reverse([]))`, expected: "0"},
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
func TestJsonBuiltins(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		// parseJson
		{desc: "parseJson string value", input: `let t = parseJson("{\"name\":\"john\"}") t["name"]`, expected: "john"},
		{desc: "parseJson int value", input: `let t = parseJson("{\"age\":42}") t["age"]`, expected: "42"},
		{desc: "parseJson bool value", input: `let t = parseJson("{\"ok\":true}") t["ok"]`, expected: "true"},
		{desc: "parseJson invalid returns null", input: `parseJson("notjson")`, expected: "null"},
		{desc: "parseJson array", input: `let a = parseJson("[1,2,3]") len(a)`, expected: "3"},

		// toJson
		{desc: "toJson int", input: `toJson(42)`, expected: "42"},
		{desc: "toJson string", input: `toJson("hello")`, expected: `"hello"`},
		{desc: "toJson bool", input: `toJson(true)`, expected: "true"},
		{desc: "toJson array", input: `toJson([1, 2, 3])`, expected: "[1,2,3]"},
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
func TestOsSystemBuiltins(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{desc: "pwd returns string", input: `type(pwd())`, expected: "STRING"},
		{desc: "osname returns string", input: `type(osname())`, expected: "STRING"},
		{desc: "osname is valid os", input: `contains(["linux", "darwin", "windows"], osname())`, expected: "true"},
		{desc: "env returns null for missing key", input: `env("TOTALLY_FAKE_KEY_XYZ")`, expected: "null"},
		{desc: "setenv and env roundtrip", input: `setenv("TEST_KEY", "hello") env("TEST_KEY")`, expected: "hello"},
		{desc: "args returns array", input: `type(args())`, expected: "ARRAY"},
		{desc: "sleep returns null", input: `sleep(1)`, expected: "null"},
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

func TestTimeBuiltins(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{desc: "timeNow returns integer", input: `type(timeNow())`, expected: "INTEGER"},
		{desc: "timeMs returns integer", input: `type(timeMs())`, expected: "INTEGER"},
		{desc: "timeMs greater than timeNow", input: `timeMs() > timeNow()`, expected: "true"},
		{desc: "timeStr returns string", input: `type(timeStr())`, expected: "STRING"},
		{desc: "dateStr returns string", input: `type(dateStr())`, expected: "STRING"},
		{desc: "dateTimeStr returns string", input: `type(dateTimeStr())`, expected: "STRING"},
		{desc: "timeFormat returns string", input: `type(timeFormat(timeNow(), "2006-01-02"))`, expected: "STRING"},
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
