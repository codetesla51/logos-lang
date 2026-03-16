package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/codetesla51/golexer/golexer"
	"github.com/codetesla51/logos/formatter"
	"github.com/codetesla51/logos/interpreter"
	"github.com/codetesla51/logos/parser"
)

const (
	PROMPT  = ">>> "
	VERSION = "0.2.4"
)

//go:embed std
var stdFiles embed.FS

func eval(input string, inter *interpreter.Interpreter) interpreter.Object {
	lexer := golexer.NewLexer(input)
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

// stdModules contains the names of standard library modules
var stdModules = map[string]bool{
	"array": true, "log": true, "math": true, "path": true,
	"string": true, "testing": true, "time": true, "type": true,
}

// findUserModules scans source code for use statements and returns non-std module names
func findUserModules(source, baseDir string, visited map[string]bool) []string {
	var modules []string
	useRegex := regexp.MustCompile(`use\s+"([^"]+)"`)
	matches := useRegex.FindAllStringSubmatch(source, -1)

	for _, match := range matches {
		modName := match[1]
		if stdModules[modName] || visited[modName] {
			continue
		}
		visited[modName] = true
		modules = append(modules, modName)

		// recursively scan this module for its dependencies
		modPath := filepath.Join(baseDir, modName+".lgs")
		if data, err := os.ReadFile(modPath); err == nil {
			subModules := findUserModules(string(data), baseDir, visited)
			modules = append(modules, subModules...)
		}
	}
	return modules
}

func buildFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read file %q: %s\n", path, err)
		os.Exit(1)
	}

	source := stripShebang(string(data))

	// verify it parses
	lexer := golexer.NewLexer(source)
	p := parser.NewParser(lexer, path)
	p.Parse()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		os.Exit(1)
	}

	// find user modules referenced by this script
	baseDir := filepath.Dir(path)
	if baseDir == "" {
		baseDir = "."
	}
	userModules := findUserModules(source, baseDir, make(map[string]bool))

	outputName := strings.TrimSuffix(filepath.Base(path), ".lgs")

	// create temp build directory
	buildDir, err := os.MkdirTemp("", "logos-build-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating build directory: %s\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(buildDir)

	// extract embedded std files to build directory
	stdDstDir := filepath.Join(buildDir, "std")
	if err := os.MkdirAll(stdDstDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating std directory: %s\n", err)
		os.Exit(1)
	}
	if err := extractEmbeddedDir(stdFiles, "std", stdDstDir); err != nil {
		fmt.Fprintf(os.Stderr, "error extracting std files: %s\n", err)
		os.Exit(1)
	}

	// create go.mod that uses the published logos module
	goModContent := fmt.Sprintf(`module logos-build

go 1.21

require (
    github.com/codetesla51/golexer v1.0.7
    github.com/codetesla51/logos v%s
)
`, VERSION)
	if err := os.WriteFile(filepath.Join(buildDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing go.mod: %s\n", err)
		os.Exit(1)
	}

	// copy user modules to std directory (so they're embedded and found via StdFs)
	for _, mod := range userModules {
		srcPath := filepath.Join(baseDir, mod+".lgs")
		dstPath := filepath.Join(stdDstDir, mod+".lgs")
		modData, err := os.ReadFile(srcPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading module %q: %s\n", mod, err)
			os.Exit(1)
		}
		if err := os.WriteFile(dstPath, modData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing module %q: %s\n", mod, err)
			os.Exit(1)
		}
	}

	goFile := filepath.Join(buildDir, "main.go")
	goSource := fmt.Sprintf(`package main

import (
    "embed"
    "fmt"
    "io/fs"
    "os"
    "github.com/codetesla51/golexer/golexer"
    "github.com/codetesla51/logos/interpreter"
    "github.com/codetesla51/logos/parser"
)

//go:embed std
var stdFiles embed.FS

const script = %s%s%s
const filename = %s%s%s

func main() {
    lexer := golexer.NewLexer(script)
    p := parser.NewParser(lexer, filename)
    program := p.Parse()
    if len(p.Errors()) != 0 {
        for _, err := range p.Errors() {
            fmt.Fprintf(os.Stderr, "%%s\n", err)
        }
        os.Exit(1)
    }

    inter := interpreter.NewInterpreter(fs.FS(stdFiles))
    result := inter.Eval(program, inter.Env)
    if result != nil && result.Type() == interpreter.ERROR_OBJ {
        fmt.Fprintln(os.Stderr, result.String())
        os.Exit(1)
    }
}
`, "`", source, "`", "`", path, "`")

	if err := os.WriteFile(goFile, []byte(goSource), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing build file: %s\n", err)
		os.Exit(1)
	}

	// get absolute path for output since we're building from temp dir
	outputPath := outputName
	if !filepath.IsAbs(outputName) {
		cwd, _ := os.Getwd()
		outputPath = filepath.Join(cwd, outputName)
	}

	// run go mod tidy to resolve dependencies
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = buildDir
	if out, err := tidyCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "go mod tidy failed: %s\n%s\n", err, out)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", outputPath, ".")
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed\n")
		os.Exit(1)
	}

	fmt.Printf("Built: %s\n", outputName)
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// extractEmbeddedDir extracts files from an embedded FS to a destination directory
func extractEmbeddedDir(efs embed.FS, srcDir, dstDir string) error {
	return fs.WalkDir(efs, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from srcDir
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		data, err := efs.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0644)
	})
}

func formatFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read file %q: %s\n", path, err)
		os.Exit(1)
	}
	source := stripShebang(string(data))
	lexer := golexer.NewLexer(source)
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
	lexer := golexer.NewLexer(source)
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
	fmt.Printf(">_ Logos v%s (%s) on %s\n", VERSION, time.Now().Format("Jan 02 2006, 15:04:05"), runtime.GOOS)
	fmt.Printf("Type \"help\" for more information\n")
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
	fmt.Printf("\033[36m>_ Logos v%s\033[0m\n\n", VERSION)
	fmt.Printf("\033[1mUSAGE:\033[0m\n")
	fmt.Printf("  lgs                    Start the REPL\n")
	fmt.Printf("  lgs <file.lgs>         Run a .lgs file\n")
	fmt.Printf("  lgs fmt <file.lgs>     Format a file\n")
	fmt.Printf("  lgs fmt .              Format all .lgs files\n")
	fmt.Printf("  lgs build <file.lgs>   Compile to binary\n")
	fmt.Printf("  lgs --version          Print version\n")
	fmt.Printf("  lgs --help             Print this message\n")
	fmt.Printf("\n\033[1mEXAMPLES:\033[0m\n")
	fmt.Printf("  lgs script.lgs\n")
	fmt.Printf("  lgs fmt .\n")
	fmt.Printf("  lgs build app.lgs\n")
	fmt.Printf("\n\033[1mDOCS:\033[0m\n")
	fmt.Printf(" https://github.com/codetesla51/logos-lang.git\n")
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
