package main

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/codetesla51/golexer/golexer"
	"github.com/codetesla51/logos/formatter"
	"github.com/codetesla51/logos/interpreter"
	"github.com/codetesla51/logos/parser"
)

const (
	PROMPT  = ">> "
	VERSION = "0.0.1"
)

var stdFiles embed.FS

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
func buildFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read file %q: %s\n", path, err)
		os.Exit(1)
	}

	source := stripShebang(string(data))

	// verify it parses
	lexer := golexer.NewLexerWithConfig(source, "tokens.json")
	p := parser.NewParser(lexer, path)
	p.Parse()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		os.Exit(1)
	}

	outputName := strings.TrimSuffix(filepath.Base(path), ".lgs")
	goFile := outputName + "_build.go"

	goSource := fmt.Sprintf(`package main

import (
    "embed"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "github.com/codetesla51/golexer/golexer"
    "github.com/codetesla51/logos/interpreter"
    "github.com/codetesla51/logos/parser"
)

//go:embed tokens.json
var tokensJson []byte

//go:embed std
var stdFiles embed.FS

const script = %s%s%s
const filename = %s

func main() {
    tmpDir, err := os.MkdirTemp("", "logos")
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %%s\n", err)
        os.Exit(1)
    }
    defer os.RemoveAll(tmpDir)

    tokensPath := filepath.Join(tmpDir, "tokens.json")
    err = os.WriteFile(tokensPath, tokensJson, 0644)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %%s\n", err)
        os.Exit(1)
    }

    lexer := golexer.NewLexerWithConfig(script, tokensPath)
    p := parser.NewParser(lexer, filename)
    program := p.Parse()
    if len(p.Errors()) != 0 {
        for _, err := range p.Errors() {
            fmt.Fprintf(os.Stderr, "%%s\n", err)
        }
        os.Exit(1)
    }

    inter := interpreter.NewInterpreter(fs.FS(stdFiles))
    inter.ToeknPath  = tokensPath
    result := inter.Eval(program, inter.Env)
    if result != nil && result.Type() == interpreter.ERROR_OBJ {
        fmt.Fprintln(os.Stderr, result.String())
        os.Exit(1)
    }
}
`, "`", source, "`", path)
	err = os.WriteFile(goFile, []byte(goSource), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing build file: %s\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", outputName, goFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	os.Remove(goFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed\n")
		os.Exit(1)
	}

	fmt.Printf("Built: %s\n", outputName)
}
func formatFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read file %q: %s\n", path, err)
		os.Exit(1)
	}
	source := stripShebang(string(data))
	lexer := golexer.NewLexerWithConfig(source, "tokens.json")
	p := parser.NewParser(lexer, path)
	program := p.Parse()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		os.Exit(1)
	}
	f := formatter.New()
	result := f.Format(program)
	err = os.WriteFile(path, []byte(result), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not write file %q: %s\n", path, err)
		os.Exit(1)
	}
	fmt.Printf("formatted %s\n", path)
}
func stripShebang(source string) string {
	if strings.HasPrefix(source, "#!") {
		if idx := strings.Index(source, "\n"); idx != -1 {
			return source[idx+1:]
		}
		return ""
	}
	return source
}

func runFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read file %q: %s\n", path, err)
		os.Exit(1)
	}
	source := stripShebang(string(data))
	inter := interpreter.NewInterpreter(stdFiles)
	inter.CurrentFile = path
	lexer := golexer.NewLexerWithConfig(source, "tokens.json")
	p := parser.NewParser(lexer, path)
	program := p.Parse()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		os.Exit(1)
	}
	result := inter.Eval(program, inter.Env)
	if result != nil && result.Type() == interpreter.ERROR_OBJ {
		fmt.Fprintln(os.Stderr, result.String())
		os.Exit(1)
	}
}

func runREPL() {
	scanner := bufio.NewScanner(os.Stdin)
	inter := interpreter.NewInterpreter(stdFiles)
	fmt.Printf("Logos v%s — REPL (ctrl+c to exit)\n", VERSION)
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

func formatAll(dir string) {
	files, err := filepath.Glob(filepath.Join(dir, "**/*.lgs"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".lgs") {
			formatFile(path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error walking directory: %s\n", err)
		os.Exit(1)
	}
	_ = files
}

func printHelp() {
	fmt.Printf(`Logos v%s - A scripting language

Usage:
  lgs               Start the REPL
  lgs <file.lgs>    Run a .lgs file
  lgs --version     Print version
  lgs --help        Print this help message
`, VERSION)
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		runREPL()
		return
	}

	switch args[0] {
	case "--version", "-v":
		fmt.Printf("Logos v%s\n", VERSION)
	case "--help", "-h":
		printHelp()
	case "fmt":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "error: fmt requires a file or directory argument\n")
			os.Exit(1)
		}
		target := args[1]
		if target == "." {
			formatAll(".")
		} else {
			info, err := os.Stat(target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
				os.Exit(1)
			}
			if info.IsDir() {
				formatAll(target)
			} else {
				formatFile(target)
			}
		}
	case "build":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "error: build requires a file argument\n")
			os.Exit(1)
		}
		buildFile(args[1])
	default:
		path := args[0]
		if !strings.HasSuffix(path, ".lgs") {
			fmt.Fprintf(os.Stderr, "error: expected .lgs file, got %q\n", path)
			os.Exit(1)
		}
		runFile(path)
	}
}
