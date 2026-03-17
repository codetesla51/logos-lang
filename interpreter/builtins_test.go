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
	tmpDir, _ := os.MkdirTemp("", "lgs-test-*")
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		desc     string
		input    string
		expected string
		setup    func()
	}{
		// fileWrite
		{
			desc:     "fileWrite creates file",
			input:    fmt.Sprintf(`fileWrite("%s/test.txt", "hello").ok`, tmpDir),
			expected: "true",
		},
		// fileRead
		{
			desc:     "fileRead existing file",
			input:    fmt.Sprintf(`fileRead("%s/test.txt").value`, tmpDir),
			expected: "hello",
			setup: func() {
				os.WriteFile(tmpDir+"/test.txt", []byte("hello"), 0644)
			},
		},
		{
			desc:     "fileRead nonexistent returns ok false",
			input:    fmt.Sprintf(`fileRead("%s/nope.txt").ok`, tmpDir),
			expected: "false",
		},
		// fileAppend
		{
			desc:     "fileAppend to file",
			input:    fmt.Sprintf(`fileAppend("%s/append.txt", "world").ok`, tmpDir),
			expected: "true",
			setup: func() {
				os.WriteFile(tmpDir+"/append.txt", []byte("hello "), 0644)
			},
		},
		{
			desc:     "fileRead appended file",
			input:    fmt.Sprintf(`fileRead("%s/append.txt").value`, tmpDir),
			expected: "hello world",
			setup: func() {
				os.WriteFile(tmpDir+"/append.txt", []byte("hello world"), 0644)
			},
		},
		// fileExists
		{
			desc:     "fileExists returns true",
			input:    fmt.Sprintf(`fileExists("%s/exists.txt")`, tmpDir),
			expected: "true",
			setup: func() {
				os.WriteFile(tmpDir+"/exists.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "fileExists returns false",
			input:    fmt.Sprintf(`fileExists("%s/missing.txt")`, tmpDir),
			expected: "false",
		},
		// fileDelete
		{
			desc:     "fileDelete removes file",
			input:    fmt.Sprintf(`fileDelete("%s/del.txt").ok`, tmpDir),
			expected: "true",
			setup: func() {
				os.WriteFile(tmpDir+"/del.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "fileDelete nonexistent returns ok false",
			input:    fmt.Sprintf(`fileDelete("%s/nope.txt").ok`, tmpDir),
			expected: "false",
		},
		// fileDeleteAll
		{
			desc:     "fileDeleteAll removes folder",
			input:    fmt.Sprintf(`fileDeleteAll("%s/subdir").ok`, tmpDir),
			expected: "true",
			setup: func() {
				os.MkdirAll(tmpDir+"/subdir/nested", 0755)
				os.WriteFile(tmpDir+"/subdir/nested/file.txt", []byte("hi"), 0644)
			},
		},
		// fileRename
		{
			desc:     "fileRename renames file",
			input:    fmt.Sprintf(`fileRename("%s/old.txt", "%s/new.txt").ok`, tmpDir, tmpDir),
			expected: "true",
			setup: func() {
				os.WriteFile(tmpDir+"/old.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "fileRename nonexistent returns ok false",
			input:    fmt.Sprintf(`fileRename("%s/nope.txt", "%s/nope2.txt").ok`, tmpDir, tmpDir),
			expected: "false",
		},
		// fileMkdir
		{
			desc:     "fileMkdir creates folder",
			input:    fmt.Sprintf(`fileMkdir("%s/newdir").ok`, tmpDir),
			expected: "true",
		},
		{
			desc:     "fileMkdir creates nested folders",
			input:    fmt.Sprintf(`fileMkdir("%s/a/b/c").ok`, tmpDir),
			expected: "true",
		},
		// fileRmdir
		{
			desc:     "fileRmdir removes empty folder",
			input:    fmt.Sprintf(`fileRmdir("%s/emptydir").ok`, tmpDir),
			expected: "true",
			setup: func() {
				os.Mkdir(tmpDir+"/emptydir", 0755)
			},
		},
		{
			desc:     "fileRmdir nonexistent returns ok false",
			input:    fmt.Sprintf(`fileRmdir("%s/nope").ok`, tmpDir),
			expected: "false",
		},
		// fileCopy
		{
			desc:     "fileCopy copies file",
			input:    fmt.Sprintf(`fileCopy("%s/src.txt", "%s/dst.txt").ok`, tmpDir, tmpDir),
			expected: "true",
			setup: func() {
				os.WriteFile(tmpDir+"/src.txt", []byte("copied"), 0644)
			},
		},
		{
			desc:     "fileCopy nonexistent returns ok false",
			input:    fmt.Sprintf(`fileCopy("%s/nope.txt", "%s/dst.txt").ok`, tmpDir, tmpDir),
			expected: "false",
		},
		// fileChmod
		{
			desc:     "fileChmod changes permissions",
			input:    fmt.Sprintf(`fileChmod("%s/chmod.txt", "0644").ok`, tmpDir),
			expected: "true",
			setup: func() {
				os.WriteFile(tmpDir+"/chmod.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "fileChmod invalid mode returns error",
			input:    fmt.Sprintf(`fileChmod("%s/chmod.txt", "badmode")`, tmpDir),
			expected: `ERROR: fileChmod() invalid mode "badmode", use octal like "0755"`,
		},
		// fileGlob
		{
			desc:     "fileGlob finds files",
			input:    fmt.Sprintf(`len(fileGlob("%s/*.txt").value)`, tmpDir),
			expected: "1",
			setup: func() {
				os.RemoveAll(tmpDir)
				os.MkdirAll(tmpDir, 0755)
				os.WriteFile(tmpDir+"/only.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "fileGlob no matches returns empty array",
			input:    fmt.Sprintf(`len(fileGlob("%s/*.xyz").value)`, tmpDir),
			expected: "0",
		},
		// fileReadDir
		{
			desc:     "fileReadDir returns file names",
			input:    fmt.Sprintf(`len(fileReadDir("%s").value)`, tmpDir),
			expected: "1",
			setup: func() {
				os.RemoveAll(tmpDir)
				os.MkdirAll(tmpDir, 0755)
				os.WriteFile(tmpDir+"/file.txt", []byte(""), 0644)
			},
		},
		{
			desc:     "fileReadDir nonexistent returns ok false",
			input:    fmt.Sprintf(`fileReadDir("%s/nope").ok`, tmpDir),
			expected: "false",
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
		// parseJson - returns {ok, value, error}, extract .value
		{desc: "parseJson string value", input: `let t = parseJson("{\"name\":\"john\"}")["value"] t["name"]`, expected: "john"},
		{desc: "parseJson int value", input: `let t = parseJson("{\"age\":42}")["value"] t["age"]`, expected: "42"},
		{desc: "parseJson bool value", input: `let t = parseJson("{\"ok\":true}")["value"] t["ok"]`, expected: "true"},
		{desc: "parseJson invalid returns null", input: `parseJson("notjson")["value"]`, expected: "null"},
		{desc: "parseJson array", input: `let a = parseJson("[1,2,3]")["value"] len(a)`, expected: "3"},

		// toJson - returns {ok, value, error}, extract .value
		{desc: "toJson int", input: `toJson(42)["value"]`, expected: "42"},
		{desc: "toJson string", input: `toJson("hello")["value"]`, expected: `"hello"`},
		{desc: "toJson bool", input: `toJson(true)["value"]`, expected: "true"},
		{desc: "toJson array", input: `toJson([1, 2, 3])["value"]`, expected: "[1,2,3]"},
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
