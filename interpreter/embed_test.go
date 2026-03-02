package interpreter

import (
	"strings"
	"testing"

	"github.com/codetesla51/golexer/golexer"
	"github.com/codetesla51/logos/parser"
)

// TestGoToObject tests conversion from Go values to Logos Objects
func TestGoToObject(t *testing.T) {
	testCases := []struct {
		input    interface{}
		expected string
		desc     string
	}{
		{nil, "null", "nil converts to null"},
		{42, "42", "int converts to integer"},
		{int64(42), "42", "int64 converts to integer"},
		{3.14, "3.14", "float64 converts to float"},
		{"hello", "hello", "string converts to string"},
		{true, "true", "true converts to bool"},
		{false, "false", "false converts to bool"},
		{[]interface{}{1, 2, 3}, "[1, 2, 3]", "slice converts to array"},
		{[]interface{}{"a", "b"}, "[a, b]", "string slice converts to array"},
		{[]interface{}{}, "[]", "empty slice converts to empty array"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			obj := GoToObject(tc.input)
			if obj.String() != tc.expected {
				t.Errorf("[%s] expected %q, got %q", tc.desc, tc.expected, obj.String())
			}
		})
	}
}

// TestObjectToGo tests conversion from Logos Objects to Go values
func TestObjectToGo(t *testing.T) {
	t.Run("integer to int64", func(t *testing.T) {
		obj := &Integer{Value: 42}
		result := ObjectToGo(obj)
		if v, ok := result.(int64); !ok || v != 42 {
			t.Errorf("expected int64(42), got %v (%T)", result, result)
		}
	})

	t.Run("float to float64", func(t *testing.T) {
		obj := &Float{Value: 3.14}
		result := ObjectToGo(obj)
		if v, ok := result.(float64); !ok || v != 3.14 {
			t.Errorf("expected float64(3.14), got %v (%T)", result, result)
		}
	})

	t.Run("string to string", func(t *testing.T) {
		obj := &String{Value: "hello"}
		result := ObjectToGo(obj)
		if v, ok := result.(string); !ok || v != "hello" {
			t.Errorf("expected string(hello), got %v (%T)", result, result)
		}
	})

	t.Run("bool true to bool", func(t *testing.T) {
		result := ObjectToGo(TRUE)
		if v, ok := result.(bool); !ok || v != true {
			t.Errorf("expected bool(true), got %v (%T)", result, result)
		}
	})

	t.Run("bool false to bool", func(t *testing.T) {
		result := ObjectToGo(FALSE)
		if v, ok := result.(bool); !ok || v != false {
			t.Errorf("expected bool(false), got %v (%T)", result, result)
		}
	})

	t.Run("null to nil", func(t *testing.T) {
		result := ObjectToGo(NULL)
		if result != nil {
			t.Errorf("expected nil, got %v (%T)", result, result)
		}
	})

	t.Run("array to slice", func(t *testing.T) {
		obj := &Array{Elements: []Object{
			&Integer{Value: 1},
			&Integer{Value: 2},
			&Integer{Value: 3},
		}}
		result := ObjectToGo(obj)
		arr, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected []interface{}, got %T", result)
		}
		if len(arr) != 3 {
			t.Errorf("expected length 3, got %d", len(arr))
		}
		if arr[0].(int64) != 1 || arr[1].(int64) != 2 || arr[2].(int64) != 3 {
			t.Errorf("unexpected array values: %v", arr)
		}
	})

	t.Run("table to map", func(t *testing.T) {
		obj := &Table{Pairs: map[string]Object{
			"STRING:name": &String{Value: "logos"},
			"STRING:age":  &Integer{Value: 1},
		}}
		result := ObjectToGo(obj)
		m, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", result)
		}
		if m["name"] != "logos" {
			t.Errorf("expected name=logos, got %v", m["name"])
		}
		if m["age"].(int64) != 1 {
			t.Errorf("expected age=1, got %v", m["age"])
		}
	})
}

// TestRegister tests registering Go functions callable from Logos scripts
func TestRegister(t *testing.T) {
	t.Run("basic registered function", func(t *testing.T) {
		inter := NewInterpreter()

		// Register a Go function that doubles a number
		inter.Register("double", func(args ...Object) Object {
			if len(args) != 1 {
				return newError("double() takes 1 argument")
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newError("double() argument must be an integer")
			}
			return &Integer{Value: n.Value * 2}
		})

		// Run a Logos script that calls the function
		lexer := golexer.NewLexer("double(21)")
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if result.String() != "42" {
			t.Errorf("expected 42, got %s", result.String())
		}
	})

	t.Run("registered function with multiple args", func(t *testing.T) {
		inter := NewInterpreter()

		// Register a function that adds two numbers
		inter.Register("add", func(args ...Object) Object {
			if len(args) != 2 {
				return newError("add() takes 2 arguments")
			}
			a, ok1 := args[0].(*Integer)
			b, ok2 := args[1].(*Integer)
			if !ok1 || !ok2 {
				return newError("add() arguments must be integers")
			}
			return &Integer{Value: a.Value + b.Value}
		})

		lexer := golexer.NewLexer("add(10, 32)")
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if result.String() != "42" {
			t.Errorf("expected 42, got %s", result.String())
		}
	})

	t.Run("instance builtins are isolated", func(t *testing.T) {
		inter1 := NewInterpreter()
		inter2 := NewInterpreter()

		// Register function only on inter1
		inter1.Register("myFunc", func(args ...Object) Object {
			return &String{Value: "from inter1"}
		})

		// inter1 should have the function
		lexer1 := golexer.NewLexer("myFunc()")
		p1 := parser.NewParser(lexer1)
		program1 := p1.Parse()
		result1 := inter1.Eval(program1, inter1.Env)

		if result1.String() != "from inter1" {
			t.Errorf("inter1 expected 'from inter1', got %s", result1.String())
		}

		// inter2 should NOT have the function
		lexer2 := golexer.NewLexer("myFunc()")
		p2 := parser.NewParser(lexer2)
		program2 := p2.Parse()
		result2 := inter2.Eval(program2, inter2.Env)

		if !strings.Contains(result2.String(), "identifier not found") {
			t.Errorf("inter2 should not have myFunc, got %s", result2.String())
		}
	})
}

// TestSetVar tests setting Go values in the Logos environment
func TestSetVar(t *testing.T) {
	t.Run("set and use integer", func(t *testing.T) {
		inter := NewInterpreter()
		inter.SetVar("x", 42)

		lexer := golexer.NewLexer("x + 1")
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if result.String() != "43" {
			t.Errorf("expected 43, got %s", result.String())
		}
	})

	t.Run("set and use string", func(t *testing.T) {
		inter := NewInterpreter()
		inter.SetVar("name", "logos")

		lexer := golexer.NewLexer(`"Hello, " + name`)
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if result.String() != "Hello, logos" {
			t.Errorf("expected 'Hello, logos', got %s", result.String())
		}
	})

	t.Run("set and use array", func(t *testing.T) {
		inter := NewInterpreter()
		inter.SetVar("arr", []interface{}{1, 2, 3})

		lexer := golexer.NewLexer("arr[1]")
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if result.String() != "2" {
			t.Errorf("expected 2, got %s", result.String())
		}
	})

	t.Run("set and use map", func(t *testing.T) {
		inter := NewInterpreter()
		inter.SetVar("config", map[string]interface{}{
			"port": 8080,
			"host": "localhost",
		})

		lexer := golexer.NewLexer(`config["port"]`)
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if result.String() != "8080" {
			t.Errorf("expected 8080, got %s", result.String())
		}
	})

	t.Run("set boolean", func(t *testing.T) {
		inter := NewInterpreter()
		inter.SetVar("enabled", true)

		lexer := golexer.NewLexer("if enabled { 1 } else { 0 }")
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if result.String() != "1" {
			t.Errorf("expected 1, got %s", result.String())
		}
	})
}

// TestGetVar tests retrieving values from the Logos environment
func TestGetVar(t *testing.T) {
	t.Run("get integer set by script", func(t *testing.T) {
		inter := NewInterpreter()

		lexer := golexer.NewLexer("let x = 42")
		p := parser.NewParser(lexer)
		program := p.Parse()
		inter.Eval(program, inter.Env)

		result := inter.GetVar("x")
		if v, ok := result.(int64); !ok || v != 42 {
			t.Errorf("expected int64(42), got %v (%T)", result, result)
		}
	})

	t.Run("get string set by script", func(t *testing.T) {
		inter := NewInterpreter()

		lexer := golexer.NewLexer(`let name = "logos"`)
		p := parser.NewParser(lexer)
		program := p.Parse()
		inter.Eval(program, inter.Env)

		result := inter.GetVar("name")
		if v, ok := result.(string); !ok || v != "logos" {
			t.Errorf("expected string(logos), got %v (%T)", result, result)
		}
	})

	t.Run("get undefined returns nil", func(t *testing.T) {
		inter := NewInterpreter()

		result := inter.GetVar("undefined")
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("get array set by script", func(t *testing.T) {
		inter := NewInterpreter()

		lexer := golexer.NewLexer("let arr = [1, 2, 3]")
		p := parser.NewParser(lexer)
		program := p.Parse()
		inter.Eval(program, inter.Env)

		result := inter.GetVar("arr")
		arr, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected []interface{}, got %T", result)
		}
		if len(arr) != 3 {
			t.Errorf("expected length 3, got %d", len(arr))
		}
	})

	t.Run("get computed value", func(t *testing.T) {
		inter := NewInterpreter()

		lexer := golexer.NewLexer("let sum = 10 + 20 + 12")
		p := parser.NewParser(lexer)
		program := p.Parse()
		inter.Eval(program, inter.Env)

		result := inter.GetVar("sum")
		if v, ok := result.(int64); !ok || v != 42 {
			t.Errorf("expected int64(42), got %v (%T)", result, result)
		}
	})
}

// TestCall tests calling Logos functions from Go
func TestCall(t *testing.T) {
	t.Run("call simple function", func(t *testing.T) {
		inter := NewInterpreter()

		// Define a function in Logos
		lexer := golexer.NewLexer("let double = fn(x) { x * 2 }")
		p := parser.NewParser(lexer)
		program := p.Parse()
		inter.Eval(program, inter.Env)

		// Call it from Go
		result, err := inter.Call("double", 21)
		if err != nil {
			t.Fatalf("Call error: %v", err)
		}
		if v, ok := result.(int64); !ok || v != 42 {
			t.Errorf("expected int64(42), got %v (%T)", result, result)
		}
	})

	t.Run("call function with multiple args", func(t *testing.T) {
		inter := NewInterpreter()

		lexer := golexer.NewLexer("let add = fn(a, b) { a + b }")
		p := parser.NewParser(lexer)
		program := p.Parse()
		inter.Eval(program, inter.Env)

		result, err := inter.Call("add", 10, 32)
		if err != nil {
			t.Fatalf("Call error: %v", err)
		}
		if v, ok := result.(int64); !ok || v != 42 {
			t.Errorf("expected int64(42), got %v (%T)", result, result)
		}
	})

	t.Run("call function with string args", func(t *testing.T) {
		inter := NewInterpreter()

		lexer := golexer.NewLexer(`let greet = fn(name) { "Hello, " + name }`)
		p := parser.NewParser(lexer)
		program := p.Parse()
		inter.Eval(program, inter.Env)

		result, err := inter.Call("greet", "World")
		if err != nil {
			t.Fatalf("Call error: %v", err)
		}
		if v, ok := result.(string); !ok || v != "Hello, World" {
			t.Errorf("expected 'Hello, World', got %v (%T)", result, result)
		}
	})

	t.Run("call undefined function returns error", func(t *testing.T) {
		inter := NewInterpreter()

		_, err := inter.Call("nonexistent", 1)
		if err == nil {
			t.Error("expected error for undefined function")
		}
		if !strings.Contains(err.Error(), "function not found") {
			t.Errorf("expected 'function not found' error, got: %v", err)
		}
	})

	t.Run("call builtin function", func(t *testing.T) {
		inter := NewInterpreter()

		result, err := inter.Call("len", "hello")
		if err != nil {
			t.Fatalf("Call error: %v", err)
		}
		if v, ok := result.(int64); !ok || v != 5 {
			t.Errorf("expected int64(5), got %v (%T)", result, result)
		}
	})

	t.Run("call registered function", func(t *testing.T) {
		inter := NewInterpreter()

		inter.Register("triple", func(args ...Object) Object {
			n := args[0].(*Integer)
			return &Integer{Value: n.Value * 3}
		})

		result, err := inter.Call("triple", 14)
		if err != nil {
			t.Fatalf("Call error: %v", err)
		}
		if v, ok := result.(int64); !ok || v != 42 {
			t.Errorf("expected int64(42), got %v (%T)", result, result)
		}
	})

	t.Run("call function that returns error", func(t *testing.T) {
		inter := NewInterpreter()

		lexer := golexer.NewLexer("let bad = fn(x) { x + y }")
		p := parser.NewParser(lexer)
		program := p.Parse()
		inter.Eval(program, inter.Env)

		_, err := inter.Call("bad", 1)
		if err == nil {
			t.Error("expected error from function")
		}
	})
}

// TestRun tests safe evaluation that catches panics
func TestRun(t *testing.T) {
	t.Run("successful eval", func(t *testing.T) {
		inter := NewInterpreter()

		err := inter.Run("let x = 42")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		result := inter.GetVar("x")
		if v, ok := result.(int64); !ok || v != 42 {
			t.Errorf("expected int64(42), got %v (%T)", result, result)
		}
	})

	t.Run("eval with parse error", func(t *testing.T) {
		inter := NewInterpreter()

		err := inter.Run("let x = ")
		if err == nil {
			t.Error("expected parse error")
		}
		if !strings.Contains(err.Error(), "parse error") {
			t.Errorf("expected parse error, got: %v", err)
		}
	})

	t.Run("eval with runtime error", func(t *testing.T) {
		inter := NewInterpreter()

		err := inter.Run("x + 1")
		if err == nil {
			t.Error("expected runtime error")
		}
		if !strings.Contains(err.Error(), "identifier not found") {
			t.Errorf("expected 'identifier not found' error, got: %v", err)
		}
	})

	t.Run("multiple evals accumulate state", func(t *testing.T) {
		inter := NewInterpreter()

		err := inter.Run("let x = 10")
		if err != nil {
			t.Fatalf("first eval error: %v", err)
		}

		err = inter.Run("let y = x * 2")
		if err != nil {
			t.Fatalf("second eval error: %v", err)
		}

		result := inter.GetVar("y")
		if v, ok := result.(int64); !ok || v != 20 {
			t.Errorf("expected int64(20), got %v (%T)", result, result)
		}
	})
}

// TestSandboxConfig tests the sandbox configuration
func TestSandboxConfig(t *testing.T) {
	t.Run("shell disabled", func(t *testing.T) {
		config := SandboxConfig{
			AllowFileIO:  true,
			AllowNetwork: true,
			AllowShell:   false,
			AllowExit:    true,
		}
		inter := NewInterpreter(config)

		lexer := golexer.NewLexer(`shell("echo hello")`)
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if !strings.Contains(result.String(), "not available in sandbox mode") {
			t.Errorf("expected sandbox error for shell, got: %s", result.String())
		}
	})

	t.Run("run disabled", func(t *testing.T) {
		config := SandboxConfig{
			AllowFileIO:  true,
			AllowNetwork: true,
			AllowShell:   false,
			AllowExit:    true,
		}
		inter := NewInterpreter(config)

		lexer := golexer.NewLexer(`run("echo", "hello")`)
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if !strings.Contains(result.String(), "not available in sandbox mode") {
			t.Errorf("expected sandbox error for run, got: %s", result.String())
		}
	})

	t.Run("network disabled", func(t *testing.T) {
		config := SandboxConfig{
			AllowFileIO:  true,
			AllowNetwork: false,
			AllowShell:   true,
			AllowExit:    true,
		}
		inter := NewInterpreter(config)

		lexer := golexer.NewLexer(`httpGet("http://example.com")`)
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if !strings.Contains(result.String(), "not available in sandbox mode") {
			t.Errorf("expected sandbox error for httpGet, got: %s", result.String())
		}
	})

	t.Run("file IO disabled", func(t *testing.T) {
		config := SandboxConfig{
			AllowFileIO:  false,
			AllowNetwork: true,
			AllowShell:   true,
			AllowExit:    true,
		}
		inter := NewInterpreter(config)

		lexer := golexer.NewLexer(`fileRead("test.txt")`)
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if !strings.Contains(result.String(), "not available in sandbox mode") {
			t.Errorf("expected sandbox error for fileRead, got: %s", result.String())
		}
	})

	t.Run("exit disabled", func(t *testing.T) {
		config := SandboxConfig{
			AllowFileIO:  true,
			AllowNetwork: true,
			AllowShell:   true,
			AllowExit:    false,
		}
		inter := NewInterpreter(config)

		lexer := golexer.NewLexer(`exit(0)`)
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if !strings.Contains(result.String(), "not available in sandbox mode") {
			t.Errorf("expected sandbox error for exit, got: %s", result.String())
		}
	})

	t.Run("all enabled by default", func(t *testing.T) {
		inter := NewInterpreter()

		// Just verify the default config allows things
		// We'll test with len() which should always work
		lexer := golexer.NewLexer(`len("hello")`)
		p := parser.NewParser(lexer)
		program := p.Parse()
		result := inter.Eval(program, inter.Env)

		if result.String() != "5" {
			t.Errorf("expected 5, got %s", result.String())
		}
	})

	t.Run("Call respects sandbox", func(t *testing.T) {
		config := SandboxConfig{
			AllowFileIO:  true,
			AllowNetwork: true,
			AllowShell:   false,
			AllowExit:    true,
		}
		inter := NewInterpreter(config)

		_, err := inter.Call("shell", "echo hello")
		if err == nil {
			t.Error("expected error when calling sandboxed function")
		}
		if !strings.Contains(err.Error(), "not available in sandbox mode") {
			t.Errorf("expected sandbox error, got: %v", err)
		}
	})
}

// TestSandboxConfigAllFileIOFuncs tests all file IO functions are blocked
func TestSandboxConfigAllFileIOFuncs(t *testing.T) {
	config := SandboxConfig{
		AllowFileIO:  false,
		AllowNetwork: true,
		AllowShell:   true,
		AllowExit:    true,
	}

	funcs := []string{
		`fileRead("test.txt")`,
		`fileWrite("test.txt", "content")`,
		`fileAppend("test.txt", "content")`,
		`fileDelete("test.txt")`,
		`fileDeleteAll("testdir")`,
		`fileCopy("src.txt", "dst.txt")`,
		`fileMove("src.txt", "dst.txt")`,
		`fileMkdir("testdir")`,
		`fileRmdir("testdir")`,
		`fileRename("old.txt", "new.txt")`,
		`fileReadDir(".")`,
		`fileGlob("*.txt")`,
		`fileChmod("test.txt", "0755")`,
		`fileExt("test.txt")`,
		`fileExists("test.txt")`,
	}

	for _, code := range funcs {
		t.Run(code, func(t *testing.T) {
			inter := NewInterpreter(config)
			lexer := golexer.NewLexer(code)
			p := parser.NewParser(lexer)
			program := p.Parse()
			result := inter.Eval(program, inter.Env)

			if !strings.Contains(result.String(), "not available in sandbox mode") {
				t.Errorf("expected sandbox error for %s, got: %s", code, result.String())
			}
		})
	}
}

// TestSandboxConfigAllNetworkFuncs tests all network functions are blocked
func TestSandboxConfigAllNetworkFuncs(t *testing.T) {
	config := SandboxConfig{
		AllowFileIO:  true,
		AllowNetwork: false,
		AllowShell:   true,
		AllowExit:    true,
	}

	funcs := []string{
		`httpGet("http://example.com")`,
		`httpPost("http://example.com", "{}")`,
		`httpPatch("http://example.com", "{}")`,
		`httpDelete("http://example.com")`,
	}

	for _, code := range funcs {
		t.Run(code, func(t *testing.T) {
			inter := NewInterpreter(config)
			lexer := golexer.NewLexer(code)
			p := parser.NewParser(lexer)
			program := p.Parse()
			result := inter.Eval(program, inter.Env)

			if !strings.Contains(result.String(), "not available in sandbox mode") {
				t.Errorf("expected sandbox error for %s, got: %s", code, result.String())
			}
		})
	}
}

// TestInterpreterIsolation tests that multiple interpreters don't interfere
func TestInterpreterIsolation(t *testing.T) {
	t.Run("globals are isolated", func(t *testing.T) {
		inter1 := NewInterpreter()
		inter2 := NewInterpreter()

		inter1.SetVar("x", 42)
		inter2.SetVar("x", 100)

		if inter1.GetVar("x").(int64) != 42 {
			t.Error("inter1 x should be 42")
		}
		if inter2.GetVar("x").(int64) != 100 {
			t.Error("inter2 x should be 100")
		}
	})

	t.Run("registered funcs are isolated", func(t *testing.T) {
		inter1 := NewInterpreter()
		inter2 := NewInterpreter()

		inter1.Register("getValue", func(args ...Object) Object {
			return &Integer{Value: 1}
		})
		inter2.Register("getValue", func(args ...Object) Object {
			return &Integer{Value: 2}
		})

		result1, _ := inter1.Call("getValue")
		result2, _ := inter2.Call("getValue")

		if result1.(int64) != 1 {
			t.Error("inter1 getValue should return 1")
		}
		if result2.(int64) != 2 {
			t.Error("inter2 getValue should return 2")
		}
	})

	t.Run("sandbox configs are isolated", func(t *testing.T) {
		sandboxed := NewInterpreter(SandboxConfig{AllowShell: false})
		normal := NewInterpreter()

		// Sandboxed should block shell
		lexer1 := golexer.NewLexer(`shell("echo test")`)
		p1 := parser.NewParser(lexer1)
		program1 := p1.Parse()
		result1 := sandboxed.Eval(program1, sandboxed.Env)
		if !strings.Contains(result1.String(), "not available in sandbox mode") {
			t.Error("sandboxed interpreter should block shell")
		}

		// Normal should allow shell (returns null because we're not actually running it)
		lexer2 := golexer.NewLexer(`shell("echo test")`)
		p2 := parser.NewParser(lexer2)
		program2 := p2.Parse()
		result2 := normal.Eval(program2, normal.Env)
		// shell() returns the output or null, not an error
		if strings.Contains(result2.String(), "not available in sandbox mode") {
			t.Error("normal interpreter should allow shell")
		}
	})
}
