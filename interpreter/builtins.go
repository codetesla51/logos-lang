package interpreter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var builtins = map[string]*Builtin{}

// -------------------------
// RESULT HELPERS
// -------------------------

// okResult wraps a successful value in a result table {ok: true, value: v, error: ""}
func okResult(value Object) Object {
	return &Table{Pairs: map[string]Object{
		"STRING:ok":    TRUE,
		"STRING:value": value,
		"STRING:error": &String{Value: ""},
	}}
}

// errResult wraps a failure in a result table {ok: false, value: null, error: msg}
func errResult(msg string, args ...interface{}) Object {
	return &Table{Pairs: map[string]Object{
		"STRING:ok":    FALSE,
		"STRING:value": NULL,
		"STRING:error": &String{Value: fmt.Sprintf(msg, args...)},
	}}
}
func formatTable(t *Table, indent int) string {
	var out strings.Builder
	prefix := strings.Repeat("  ", indent)
	out.WriteString("{\n")
	for k, val := range t.Pairs {
		cleanKey := strings.TrimPrefix(k, "STRING:")
		out.WriteString(prefix + "  " + cleanKey + ": ")
		switch v := val.(type) {
		case *Table:
			out.WriteString(formatTable(v, indent+1))
		default:
			out.WriteString(val.String() + "\n")
		}
	}
	out.WriteString(prefix + "}\n")
	return out.String()
}

func init() {

	// -------------------------
	// I/O
	// -------------------------

	// print(args...) - prints values to stdout separated by spaces
	builtins["print"] = &Builtin{
		Fn: func(args ...Object) Object {
			var out strings.Builder
			for i, arg := range args {
				if i > 0 {
					out.WriteString(" ")
				}
				switch v := arg.(type) {
				case *Table:
					out.WriteString(formatTable(v, 0))
				default:
					out.WriteString(v.String())
				}
			}

			fmt.Println(out.String())
			return NULL
		},
	}

	// input(prompt?) - reads a line from stdin, optional prompt string
	builtins["input"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) > 1 {
				return newError("input() takes at most 1 argument, got %d", len(args))
			}
			if len(args) == 1 {
				fmt.Print(args[0].String())
			}
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if err := scanner.Err(); err != nil {
				return newError("input() error: %s", err.Error())
			}
			return &String{Value: scanner.Text()}
		},
	}

	// prompt(message) - prints message then reads a line from stdin
	builtins["prompt"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("prompt() takes 1 argument, got %d", len(args))
			}
			fmt.Print(args[0].String())
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if err := scanner.Err(); err != nil {
				return newError("prompt() error: %s", err.Error())
			}
			return &String{Value: scanner.Text()}
		},
	}

	// clear() - clears the terminal screen
	builtins["clear"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("clear() takes no arguments, got %d", len(args))
			}
			fmt.Print("\033[H\033[2J")
			return NULL
		},
	}

	builtins["confirm"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("confirm() takes 1 argument, got %d", len(args))
			}
			msg, ok := args[0].(*String)
			if !ok {
				return newError("confirm() argument must be a string")
			}
			fmt.Print(msg.Value + " (y/n): ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if err := scanner.Err(); err != nil {
				return FALSE
			}
			input := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if input == "y" || input == "yes" {
				return TRUE
			}
			return FALSE
		},
	}

	builtins["select"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("select() takes 2 arguments, got %d", len(args))
			}
			msg, ok := args[0].(*String)
			if !ok {
				return newError("select() first argument must be a string")
			}
			options, ok := args[1].(*Array)
			if !ok {
				return newError("select() second argument must be an array")
			}
			if len(options.Elements) == 0 {
				return newError("select() options array cannot be empty")
			}
			fmt.Println(msg.Value)
			for i, opt := range options.Elements {
				fmt.Printf("  %d) %s\n", i+1, opt.String())
			}
			fmt.Print("Enter number: ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if err := scanner.Err(); err != nil {
				return NULL
			}
			n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || n < 1 || n > len(options.Elements) {
				return newError("select() invalid choice")
			}
			return options.Elements[n-1]
		},
	}

	// -------------------------
	// COLOR OUTPUT
	// -------------------------

	// colorRed(str) - returns string wrapped in red ANSI color codes
	builtins["colorRed"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("colorRed() takes 1 argument, got %d", len(args))
			}
			return &String{Value: "\033[31m" + args[0].String() + "\033[0m"}
		},
	}

	// colorGreen(str) - returns string wrapped in green ANSI color codes
	builtins["colorGreen"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("colorGreen() takes 1 argument, got %d", len(args))
			}
			return &String{Value: "\033[32m" + args[0].String() + "\033[0m"}
		},
	}

	// colorYellow(str) - returns string wrapped in yellow ANSI color codes
	builtins["colorYellow"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("colorYellow() takes 1 argument, got %d", len(args))
			}
			return &String{Value: "\033[33m" + args[0].String() + "\033[0m"}
		},
	}

	// colorBlue(str) - returns string wrapped in blue ANSI color codes
	builtins["colorBlue"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("colorBlue() takes 1 argument, got %d", len(args))
			}
			return &String{Value: "\033[34m" + args[0].String() + "\033[0m"}
		},
	}

	// colorMagenta(str) - returns string wrapped in magenta ANSI color codes
	builtins["colorMagenta"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("colorMagenta() takes 1 argument, got %d", len(args))
			}
			return &String{Value: "\033[35m" + args[0].String() + "\033[0m"}
		},
	}

	// colorCyan(str) - returns string wrapped in cyan ANSI color codes
	builtins["colorCyan"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("colorCyan() takes 1 argument, got %d", len(args))
			}
			return &String{Value: "\033[36m" + args[0].String() + "\033[0m"}
		},
	}

	// colorWhite(str) - returns string wrapped in white ANSI color codes
	builtins["colorWhite"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("colorWhite() takes 1 argument, got %d", len(args))
			}
			return &String{Value: "\033[37m" + args[0].String() + "\033[0m"}
		},
	}

	// colorBold(str) - returns string wrapped in bold ANSI codes
	builtins["colorBold"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("colorBold() takes 1 argument, got %d", len(args))
			}
			return &String{Value: "\033[1m" + args[0].String() + "\033[0m"}
		},
	}

	// -------------------------
	// TYPE
	// -------------------------

	// type(value) - returns the type name of a value as a string
	builtins["type"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("type() takes exactly 1 argument, got %d", len(args))
			}
			return &String{Value: string(args[0].Type())}
		},
	}

	// -------------------------
	// LENGTH
	// -------------------------

	// len(value) - returns the length of a string or array
	builtins["len"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("len() takes exactly 1 argument, got %d", len(args))
			}
			switch arg := args[0].(type) {
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("len() not supported for %s", args[0].Type())
			}
		},
	}

	// -------------------------
	// FILE OPS
	// -------------------------

	// fileRead(path) - reads a file and returns {ok, value, error}
	builtins["fileRead"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("fileRead() takes exactly 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("fileRead() argument must be a string (filepath)")
			}
			data, err := os.ReadFile(path.Value)
			if err != nil {
				return errResult("fileRead() failed: %s", err.Error())
			}
			return okResult(&String{Value: string(data)})
		},
	}

	// fileWrite(path, content) - overwrites a file, returns {ok, value, error}
	builtins["fileWrite"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("fileWrite() takes 2 arguments, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("fileWrite() first argument must be a string (filepath)")
			}
			content, ok := args[1].(*String)
			if !ok {
				return newError("fileWrite() second argument must be a string")
			}
			err := os.WriteFile(path.Value, []byte(content.Value), 0644)
			if err != nil {
				return errResult("fileWrite() failed: %s", err.Error())
			}
			return okResult(NULL)
		},
	}

	// fileAppend(path, content) - appends content to a file, returns {ok, value, error}
	builtins["fileAppend"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("fileAppend() takes 2 arguments, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("fileAppend() first argument must be a string (filepath)")
			}
			content, ok := args[1].(*String)
			if !ok {
				return newError("fileAppend() second argument must be a string")
			}
			f, err := os.OpenFile(path.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return errResult("fileAppend() failed: %s", err.Error())
			}
			defer f.Close()
			_, err = f.WriteString(content.Value)
			if err != nil {
				return errResult("fileAppend() failed: %s", err.Error())
			}
			return okResult(NULL)
		},
	}

	// fileExists(path) - returns true if file or folder exists, false otherwise
	builtins["fileExists"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("fileExists() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("fileExists() argument must be a string (filepath)")
			}
			_, err := os.Stat(path.Value)
			if os.IsNotExist(err) {
				return FALSE
			}
			return TRUE
		},
	}

	// fileDelete(path) - deletes a file or empty folder, returns {ok, value, error}
	builtins["fileDelete"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("fileDelete() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("fileDelete() argument must be a string (filepath)")
			}
			err := os.Remove(path.Value)
			if err != nil {
				return errResult("fileDelete() failed: %s", err.Error())
			}
			return okResult(NULL)
		},
	}

	// fileDeleteAll(path) - deletes a folder and all its contents, returns {ok, value, error}
	builtins["fileDeleteAll"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("fileDeleteAll() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("fileDeleteAll() argument must be a string (folderpath)")
			}
			err := os.RemoveAll(path.Value)
			if err != nil {
				return errResult("fileDeleteAll() failed: %s", err.Error())
			}
			return okResult(NULL)
		},
	}

	// fileRename(oldPath, newPath) - renames or moves a file/folder, returns {ok, value, error}
	builtins["fileRename"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("fileRename() takes 2 arguments, got %d", len(args))
			}
			oldPath, ok := args[0].(*String)
			if !ok {
				return newError("fileRename() first argument must be a string")
			}
			newPath, ok := args[1].(*String)
			if !ok {
				return newError("fileRename() second argument must be a string")
			}
			err := os.Rename(oldPath.Value, newPath.Value)
			if err != nil {
				return errResult("fileRename() failed: %s", err.Error())
			}
			return okResult(NULL)
		},
	}

	// fileMkdir(path) - creates a folder and any missing parents, returns {ok, value, error}
	builtins["fileMkdir"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("fileMkdir() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("fileMkdir() argument must be a string (folderpath)")
			}
			err := os.MkdirAll(path.Value, 0755)
			if err != nil {
				return errResult("fileMkdir() failed: %s", err.Error())
			}
			return okResult(NULL)
		},
	}

	// fileRmdir(path) - removes an empty directory, returns {ok, value, error}
	builtins["fileRmdir"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("fileRmdir() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("fileRmdir() argument must be a string (folderpath)")
			}
			err := os.Remove(path.Value)
			if err != nil {
				return errResult("fileRmdir() failed: %s", err.Error())
			}
			return okResult(NULL)
		},
	}

	// fileReadDir(path) - returns array of file/folder names in directory, returns {ok, value, error}
	builtins["fileReadDir"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("fileReadDir() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("fileReadDir() argument must be a string (folderpath)")
			}
			entries, err := os.ReadDir(path.Value)
			if err != nil {
				return errResult("fileReadDir() failed: %s", err.Error())
			}
			items := make([]Object, len(entries))
			for i, entry := range entries {
				items[i] = &String{Value: entry.Name()}
			}
			return okResult(&Array{Elements: items})
		},
	}

	// fileCopy(src, dst) - copies a file from src to dst, returns {ok, value, error}
	builtins["fileCopy"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("fileCopy() takes 2 arguments, got %d", len(args))
			}
			src, ok := args[0].(*String)
			if !ok {
				return newError("fileCopy() first argument must be a string (source path)")
			}
			dst, ok := args[1].(*String)
			if !ok {
				return newError("fileCopy() second argument must be a string (destination path)")
			}
			data, err := os.ReadFile(src.Value)
			if err != nil {
				return errResult("fileCopy() failed to read source: %s", err.Error())
			}
			err = os.WriteFile(dst.Value, data, 0644)
			if err != nil {
				return errResult("fileCopy() failed to write destination: %s", err.Error())
			}
			return okResult(NULL)
		},
	}

	// fileMove(src, dst) - moves a file or folder from src to dst, returns {ok, value, error}
	builtins["fileMove"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("fileMove() takes 2 arguments, got %d", len(args))
			}
			src, ok := args[0].(*String)
			if !ok {
				return newError("fileMove() first argument must be a string (source path)")
			}
			dst, ok := args[1].(*String)
			if !ok {
				return newError("fileMove() second argument must be a string (destination path)")
			}
			err := os.Rename(src.Value, dst.Value)
			if err != nil {
				return errResult("fileMove() failed: %s", err.Error())
			}
			return okResult(NULL)
		},
	}

	// fileChmod(path, mode) - changes file permissions, mode is octal string e.g "0755"
	builtins["fileChmod"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("fileChmod() takes 2 arguments, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("fileChmod() first argument must be a string (filepath)")
			}
			modeStr, ok := args[1].(*String)
			if !ok {
				return newError("fileChmod() second argument must be a string (e.g \"0755\")")
			}
			mode, err := strconv.ParseUint(modeStr.Value, 8, 32)
			if err != nil {
				return newError("fileChmod() invalid mode \"%s\", use octal like \"0755\"", modeStr.Value)
			}
			err = os.Chmod(path.Value, os.FileMode(mode))
			if err != nil {
				return errResult("fileChmod() failed: %s", err.Error())
			}
			return okResult(NULL)
		},
	}

	// fileGlob(pattern) - returns array of paths matching a glob pattern e.g "*.txt"
	builtins["fileGlob"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("fileGlob() takes 1 argument, got %d", len(args))
			}
			pattern, ok := args[0].(*String)
			if !ok {
				return newError("fileGlob() argument must be a string (pattern)")
			}
			matches, err := filepath.Glob(pattern.Value)
			if err != nil {
				return errResult("fileGlob() failed: %s", err.Error())
			}
			items := make([]Object, len(matches))
			for i, m := range matches {
				items[i] = &String{Value: m}
			}
			return okResult(&Array{Elements: items})
		},
	}

	// fileExt(path) - returns the file extension of a path e.g ".txt"
	builtins["fileExt"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("fileExt() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("fileExt() argument must be a string (filepath)")
			}
			return &String{Value: filepath.Ext(path.Value)}
		},
	}

	// -------------------------
	// STRING OPS
	// -------------------------

	// upper(str) - returns string converted to uppercase
	builtins["upper"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("upper() takes 1 argument, got %d", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("upper() argument must be a string")
			}
			return &String{Value: strings.ToUpper(str.Value)}
		},
	}

	// lower(str) - returns string converted to lowercase
	builtins["lower"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("lower() takes 1 argument, got %d", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("lower() argument must be a string")
			}
			return &String{Value: strings.ToLower(str.Value)}
		},
	}

	// trim(str) - removes leading and trailing whitespace
	builtins["trim"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("trim() takes 1 argument, got %d", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("trim() argument must be a string")
			}
			return &String{Value: strings.TrimSpace(str.Value)}
		},
	}

	// replace(str, old, new) - replaces all occurrences of old with new
	builtins["replace"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("replace() takes 3 arguments, got %d", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("replace() first argument must be a string")
			}
			old, ok := args[1].(*String)
			if !ok {
				return newError("replace() second argument must be a string")
			}
			newStr, ok := args[2].(*String)
			if !ok {
				return newError("replace() third argument must be a string")
			}
			return &String{Value: strings.ReplaceAll(str.Value, old.Value, newStr.Value)}
		},
	}

	// split(str, delimiter) - splits a string into an array by delimiter
	builtins["split"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("split() takes 2 arguments, got %d", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("split() first argument must be a string")
			}
			delim, ok := args[1].(*String)
			if !ok {
				return newError("split() second argument must be a string")
			}
			parts := strings.Split(str.Value, delim.Value)
			items := make([]Object, len(parts))
			for i, p := range parts {
				items[i] = &String{Value: p}
			}
			return &Array{Elements: items}
		},
	}

	// join(array, delimiter) - joins an array of strings into one string
	builtins["join"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("join() takes 2 arguments, got %d", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("join() first argument must be an array")
			}
			delim, ok := args[1].(*String)
			if !ok {
				return newError("join() second argument must be a string")
			}
			parts := make([]string, len(arr.Elements))
			for i, el := range arr.Elements {
				parts[i] = el.String()
			}
			return &String{Value: strings.Join(parts, delim.Value)}
		},
	}

	// contains(str|array, value) - returns true if string contains substring or array contains value
	builtins["contains"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("contains() takes 2 arguments, got %d", len(args))
			}
			switch arg := args[0].(type) {
			case *Array:
				for _, el := range arg.Elements {
					if el.String() == args[1].String() {
						return TRUE
					}
				}
				return FALSE
			case *String:
				sub, ok := args[1].(*String)
				if !ok {
					return newError("contains() second argument must be a string")
				}
				if strings.Contains(arg.Value, sub.Value) {
					return TRUE
				}
				return FALSE
			default:
				return newError("contains() first argument must be an array or string")
			}
		},
	}

	// startsWith(str, prefix) - returns true if string starts with prefix
	builtins["startsWith"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("startsWith() takes 2 arguments, got %d", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("startsWith() first argument must be a string")
			}
			prefix, ok := args[1].(*String)
			if !ok {
				return newError("startsWith() second argument must be a string")
			}
			if strings.HasPrefix(str.Value, prefix.Value) {
				return TRUE
			}
			return FALSE
		},
	}

	// endsWith(str, suffix) - returns true if string ends with suffix
	builtins["endsWith"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("endsWith() takes 2 arguments, got %d", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("endsWith() first argument must be a string")
			}
			suffix, ok := args[1].(*String)
			if !ok {
				return newError("endsWith() second argument must be a string")
			}
			if strings.HasSuffix(str.Value, suffix.Value) {
				return TRUE
			}
			return FALSE
		},
	}

	// indexOf(str, substr) - returns index of first occurrence, or -1 if not found
	builtins["indexOf"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("indexOf() takes 2 arguments, got %d", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("indexOf() first argument must be a string")
			}
			sub, ok := args[1].(*String)
			if !ok {
				return newError("indexOf() second argument must be a string")
			}
			return &Integer{Value: int64(strings.Index(str.Value, sub.Value))}
		},
	}

	// repeat(str, n) - repeats a string n times
	builtins["repeat"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("repeat() takes 2 arguments, got %d", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("repeat() first argument must be a string")
			}
			n, ok := args[1].(*Integer)
			if !ok {
				return newError("repeat() second argument must be an integer")
			}
			return &String{Value: strings.Repeat(str.Value, int(n.Value))}
		},
	}

	// slice(str|array, start, end) - returns substring or sub-array from start to end (exclusive)
	builtins["slice"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("slice() takes 3 arguments, got %d", len(args))
			}
			start, ok := args[1].(*Integer)
			if !ok {
				return newError("slice() second argument must be an integer (start)")
			}
			end, ok := args[2].(*Integer)
			if !ok {
				return newError("slice() third argument must be an integer (end)")
			}
			switch arg := args[0].(type) {
			case *String:
				s := int(start.Value)
				e := int(end.Value)
				if s < 0 || e > len(arg.Value) || s > e {
					return newError("slice() index out of bounds")
				}
				return &String{Value: arg.Value[s:e]}
			case *Array:
				s := int(start.Value)
				e := int(end.Value)
				if s < 0 || e > len(arg.Elements) || s > e {
					return newError("slice() index out of bounds")
				}
				newElements := make([]Object, e-s)
				copy(newElements, arg.Elements[s:e])
				return &Array{Elements: newElements}
			default:
				return newError("slice() first argument must be a string or array")
			}
		},
	}

	// format(template, args...) - formats a string using %s %d %f placeholders
	builtins["format"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("format() takes at least 1 argument, got %d", len(args))
			}
			template, ok := args[0].(*String)
			if !ok {
				return newError("format() first argument must be a string (template)")
			}
			fmtArgs := make([]interface{}, len(args)-1)
			for i, arg := range args[1:] {
				switch v := arg.(type) {
				case *Integer:
					fmtArgs[i] = v.Value
				case *Float:
					fmtArgs[i] = v.Value
				case *Bool:
					fmtArgs[i] = v.Value
				default:
					fmtArgs[i] = arg.String()
				}
			}
			return &String{Value: fmt.Sprintf(template.Value, fmtArgs...)}
		},
	}

	// -------------------------
	// TYPE CONVERSION
	// -------------------------

	// toInt(value) - converts a value to an integer, returns null on failure
	builtins["toInt"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("toInt() takes 1 argument, got %d", len(args))
			}
			switch arg := args[0].(type) {
			case *Integer:
				return arg
			case *Float:
				return &Integer{Value: int64(arg.Value)}
			case *Bool:
				if arg.Value {
					return &Integer{Value: 1}
				}
				return &Integer{Value: 0}
			case *String:
				n, err := strconv.ParseInt(arg.Value, 10, 64)
				if err != nil {
					return NULL
				}
				return &Integer{Value: n}
			default:
				return NULL
			}
		},
	}

	// toFloat(value) - converts a value to a float, returns null on failure
	builtins["toFloat"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("toFloat() takes 1 argument, got %d", len(args))
			}
			switch arg := args[0].(type) {
			case *Float:
				return arg
			case *Integer:
				return &Float{Value: float64(arg.Value)}
			case *Bool:
				if arg.Value {
					return &Float{Value: 1.0}
				}
				return &Float{Value: 0.0}
			case *String:
				n, err := strconv.ParseFloat(arg.Value, 64)
				if err != nil {
					return NULL
				}
				return &Float{Value: n}
			default:
				return NULL
			}
		},
	}

	// toBool(value) - converts a value to a boolean, returns null on failure
	builtins["toBool"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("toBool() takes 1 argument, got %d", len(args))
			}
			switch arg := args[0].(type) {
			case *Bool:
				return arg
			case *Integer:
				if arg.Value != 0 {
					return TRUE
				}
				return FALSE
			case *Float:
				if arg.Value != 0.0 {
					return TRUE
				}
				return FALSE
			case *String:
				if arg.Value == "true" {
					return TRUE
				} else if arg.Value == "false" {
					return FALSE
				}
				return NULL
			default:
				return NULL
			}
		},
	}

	// toStr(value) - converts any value to its string representation
	builtins["toStr"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("toStr() takes 1 argument, got %d", len(args))
			}
			return &String{Value: args[0].String()}
		},
	}

	// -------------------------
	// ARRAY OPS
	// -------------------------

	// push(array, value) - returns new array with value added to the end
	builtins["push"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("push() takes 2 arguments, got %d", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("push() first argument must be an array")
			}
			newElements := make([]Object, len(arr.Elements)+1)
			copy(newElements, arr.Elements)
			newElements[len(arr.Elements)] = args[1]
			return &Array{Elements: newElements}
		},
	}

	// pop(array) - returns the last element of an array, null if empty
	builtins["pop"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("pop() takes 1 argument, got %d", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("pop() argument must be an array")
			}
			if len(arr.Elements) == 0 {
				return NULL
			}
			return arr.Elements[len(arr.Elements)-1]
		},
	}

	// first(array) - returns the first element of an array, null if empty
	builtins["first"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("first() takes 1 argument, got %d", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first() argument must be an array")
			}
			if len(arr.Elements) == 0 {
				return NULL
			}
			return arr.Elements[0]
		},
	}

	// last(array) - returns the last element of an array, null if empty
	builtins["last"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("last() takes 1 argument, got %d", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("last() argument must be an array")
			}
			if len(arr.Elements) == 0 {
				return NULL
			}
			return arr.Elements[len(arr.Elements)-1]
		},
	}

	// tail(array) - returns all elements except the first, null if empty
	builtins["tail"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("tail() takes 1 argument, got %d", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("tail() argument must be an array")
			}
			if len(arr.Elements) == 0 {
				return NULL
			}
			newElements := make([]Object, len(arr.Elements)-1)
			copy(newElements, arr.Elements[1:])
			return &Array{Elements: newElements}
		},
	}

	// prepend(array, value) - returns new array with value added to the front
	builtins["prepend"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("prepend() takes 2 arguments, got %d", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("prepend() first argument must be an array")
			}
			newElements := make([]Object, len(arr.Elements)+1)
			newElements[0] = args[1]
			copy(newElements[1:], arr.Elements)
			return &Array{Elements: newElements}
		},
	}

	// reverse(array) - returns a new array with elements in reverse order
	builtins["reverse"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("reverse() takes 1 argument, got %d", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("reverse() argument must be an array")
			}
			newElements := make([]Object, len(arr.Elements))
			for i, el := range arr.Elements {
				newElements[len(arr.Elements)-1-i] = el
			}
			return &Array{Elements: newElements}
		},
	}

	builtins["sort"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("sort() takes 1 argument, got %d", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("sort() argument must be an array")
			}
			newElements := make([]Object, len(arr.Elements))
			copy(newElements, arr.Elements)
			sort.Slice(newElements, func(i, j int) bool {
				a := newElements[i]
				b := newElements[j]

				// both numbers
				af := toFloat64(a)
				bf := toFloat64(b)
				if af != nil && bf != nil {
					return *af < *bf
				}

				// both strings
				as, aok := a.(*String)
				bs, bok := b.(*String)
				if aok && bok {
					return as.Value < bs.Value
				}

				return false
			})
			return &Array{Elements: newElements}
		},
	}

	// -------------------------
	// TABLE OPS
	// -------------------------

	// keys(table) - returns all keys of a table as an array
	builtins["keys"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("keys() takes 1 argument, got %d", len(args))
			}
			table, ok := args[0].(*Table)
			if !ok {
				return newError("keys() argument must be a table")
			}
			items := make([]Object, 0, len(table.Pairs))
			for k := range table.Pairs {
				parts := strings.SplitN(k, ":", 2)
				if len(parts) == 2 {
					items = append(items, &String{Value: parts[1]})
				} else {
					items = append(items, &String{Value: k})
				}
			}
			return &Array{Elements: items}
		},
	}

	// values(table) - returns all values of a table as an array
	builtins["values"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("values() takes 1 argument, got %d", len(args))
			}
			table, ok := args[0].(*Table)
			if !ok {
				return newError("values() argument must be a table")
			}
			items := make([]Object, 0, len(table.Pairs))
			for _, v := range table.Pairs {
				items = append(items, v)
			}
			return &Array{Elements: items}
		},
	}

	// has(table, key) - returns true if table has the given key
	builtins["has"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("has() takes 2 arguments, got %d", len(args))
			}
			table, ok := args[0].(*Table)
			if !ok {
				return newError("has() first argument must be a table")
			}
			key, ok := args[1].(*String)
			if !ok {
				return newError("has() second argument must be a string")
			}
			_, exists := table.Pairs["STRING:"+key.Value]
			if exists {
				return TRUE
			}
			return FALSE
		},
	}

	// tableDelete(table, key) - returns new table without the specified key
	builtins["tableDelete"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("tableDelete() takes 2 arguments, got %d", len(args))
			}
			table, ok := args[0].(*Table)
			if !ok {
				return newError("tableDelete() first argument must be a table")
			}
			key, ok := args[1].(*String)
			if !ok {
				return newError("tableDelete() second argument must be a string")
			}
			newPairs := map[string]Object{}
			for k, v := range table.Pairs {
				if k != "STRING:"+key.Value {
					newPairs[k] = v
				}
			}
			return &Table{Pairs: newPairs}
		},
	}

	// merge(table1, table2) - merges two tables, table2 values overwrite table1 on conflict
	builtins["merge"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("merge() takes 2 arguments, got %d", len(args))
			}
			table1, ok := args[0].(*Table)
			if !ok {
				return newError("merge() first argument must be a table")
			}
			table2, ok := args[1].(*Table)
			if !ok {
				return newError("merge() second argument must be a table")
			}
			newPairs := map[string]Object{}
			for k, v := range table1.Pairs {
				newPairs[k] = v
			}
			for k, v := range table2.Pairs {
				newPairs[k] = v
			}
			return &Table{Pairs: newPairs}
		},
	}

	// -------------------------
	// JSON
	// -------------------------

	// parseJson(str) - parses a JSON string into a table/array/value, returns null on failure
	builtins["parseJson"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("parseJson() takes 1 argument, got %d", len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("parseJson() argument must be a string")
			}
			var raw interface{}
			if err := json.Unmarshal([]byte(str.Value), &raw); err != nil {
				return errResult("parseJson Error: %s", err)
			}
			return okResult(jsonToObject(raw))
		},
	}

	// toJson(value) - converts a value to a JSON string, returns {ok, value, error}
	builtins["toJson"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("toJson() takes 1 argument, got %d", len(args))
			}
			raw := objectToJson(args[0])
			data, err := json.Marshal(raw)
			if err != nil {
				return errResult("toJson() failed: %s", err.Error())
			}
			return okResult(&String{Value: string(data)})
		},
	}
	// prettyJson(value) - converts a value to pretty-printed JSON, returns {ok, value, error}
	builtins["prettyJson"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("prettyJson() takes 1 argument, got %d", len(args))
			}
			raw := objectToJson(args[0])
			data, err := json.MarshalIndent(raw, "", "  ")
			if err != nil {
				return errResult("prettyJson() failed: %s", err.Error())
			}
			lines := strings.Split(string(data), "\n")
			for i, line := range lines {
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					key := parts[0]
					value := parts[1]
					key = "\033[32m" + key + "\033[0m"
					value = "\033[34m" + value + "\033[0m"
					lines[i] = key + ":" + value
				}
			}
			return okResult(&String{Value: strings.Join(lines, "\n")})
		},
	}

	// -------------------------
	// MATH
	// -------------------------

	// mathAbs(n) - returns the absolute value of a number
	builtins["mathAbs"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("mathAbs() takes 1 argument, got %d", len(args))
			}
			switch arg := args[0].(type) {
			case *Integer:
				if arg.Value < 0 {
					return &Integer{Value: -arg.Value}
				}
				return arg
			case *Float:
				return &Float{Value: math.Abs(arg.Value)}
			default:
				return newError("mathAbs() argument must be a number")
			}
		},
	}

	// mathPow(base, exp) - returns base raised to the power of exp
	builtins["mathPow"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("mathPow() takes 2 arguments, got %d", len(args))
			}
			base := toFloat64(args[0])
			if base == nil {
				return newError("mathPow() first argument must be a number")
			}
			exp := toFloat64(args[1])
			if exp == nil {
				return newError("mathPow() second argument must be a number")
			}
			return &Float{Value: math.Pow(*base, *exp)}
		},
	}

	// mathSqrt(n) - returns the square root of a number
	builtins["mathSqrt"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("mathSqrt() takes 1 argument, got %d", len(args))
			}
			n := toFloat64(args[0])
			if n == nil {
				return newError("mathSqrt() argument must be a number")
			}
			return &Float{Value: math.Sqrt(*n)}
		},
	}

	// mathFloor(n) - rounds a number down to the nearest integer
	builtins["mathFloor"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("mathFloor() takes 1 argument, got %d", len(args))
			}
			n := toFloat64(args[0])
			if n == nil {
				return newError("mathFloor() argument must be a number")
			}
			return &Integer{Value: int64(math.Floor(*n))}
		},
	}

	// mathCeil(n) - rounds a number up to the nearest integer
	builtins["mathCeil"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("mathCeil() takes 1 argument, got %d", len(args))
			}
			n := toFloat64(args[0])
			if n == nil {
				return newError("mathCeil() argument must be a number")
			}
			return &Integer{Value: int64(math.Ceil(*n))}
		},
	}

	// mathRound(n) - rounds a number to the nearest integer
	builtins["mathRound"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("mathRound() takes 1 argument, got %d", len(args))
			}
			n := toFloat64(args[0])
			if n == nil {
				return newError("mathRound() argument must be a number")
			}
			return &Integer{Value: int64(math.Round(*n))}
		},
	}

	// mathMin(a, b) - returns the smaller of two numbers
	builtins["mathMin"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("mathMin() takes 2 arguments, got %d", len(args))
			}
			a := toFloat64(args[0])
			b := toFloat64(args[1])
			if a == nil || b == nil {
				return newError("mathMin() arguments must be numbers")
			}
			return &Float{Value: math.Min(*a, *b)}
		},
	}

	// mathMax(a, b) - returns the larger of two numbers
	builtins["mathMax"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("mathMax() takes 2 arguments, got %d", len(args))
			}
			a := toFloat64(args[0])
			b := toFloat64(args[1])
			if a == nil || b == nil {
				return newError("mathMax() arguments must be numbers")
			}
			return &Float{Value: math.Max(*a, *b)}
		},
	}

	// mathRandom() - returns a random float between 0.0 and 1.0
	builtins["mathRandom"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("mathRandom() takes no arguments, got %d", len(args))
			}
			return &Float{Value: rand.Float64()}
		},
	}

	// mathRandomInt(min, max) - returns a random integer between min and max (inclusive)
	builtins["mathRandomInt"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("mathRandomInt() takes 2 arguments, got %d", len(args))
			}
			min, ok := args[0].(*Integer)
			if !ok {
				return newError("mathRandomInt() first argument must be an integer")
			}
			max, ok := args[1].(*Integer)
			if !ok {
				return newError("mathRandomInt() second argument must be an integer")
			}
			if min.Value > max.Value {
				return newError("mathRandomInt() min must be less than or equal to max")
			}
			return &Integer{Value: min.Value + rand.Int63n(max.Value-min.Value+1)}
		},
	}

	// mathPi() - returns the value of pi (3.14159...)
	builtins["mathPi"] = &Builtin{
		Fn: func(args ...Object) Object {
			return &Float{Value: math.Pi}
		},
	}

	// -------------------------
	// OS / SYSTEM
	// -------------------------

	// pwd() - returns the current working directory
	builtins["pwd"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("pwd() takes no arguments, got %d", len(args))
			}
			dir, err := os.Getwd()
			if err != nil {
				return NULL
			}
			return &String{Value: dir}
		},
	}

	// cd(path) - changes the current working directory
	builtins["cd"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("cd() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("cd() argument must be a string")
			}
			err := os.Chdir(path.Value)
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// env(key) - returns the value of an environment variable, null if not set
	builtins["env"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("env() takes 1 argument, got %d", len(args))
			}
			key, ok := args[0].(*String)
			if !ok {
				return newError("env() argument must be a string")
			}
			val := os.Getenv(key.Value)
			if val == "" {
				return NULL
			}
			return &String{Value: val}
		},
	}

	// setenv(key, value) - sets an environment variable
	builtins["setenv"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("setenv() takes 2 arguments, got %d", len(args))
			}
			key, ok := args[0].(*String)
			if !ok {
				return newError("setenv() first argument must be a string")
			}
			value, ok := args[1].(*String)
			if !ok {
				return newError("setenv() second argument must be a string")
			}
			err := os.Setenv(key.Value, value.Value)
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// args() - returns command line arguments as an array of strings
	// todo safGaurd slice
	builtins["args"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("args() takes no arguments, got %d", len(args))
			}
			if len(os.Args) <= 2 {
				return &Array{Elements: []Object{}}
			} else {
				osArgs := os.Args[2:] // skip program name and script name
				items := make([]Object, len(osArgs))
				for i, arg := range osArgs {
					items[i] = &String{Value: arg}
				}
				return &Array{Elements: items}
			}
		},
	}

	// exit(code?) - exits the program with optional status code (default 0)
	builtins["exit"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) > 1 {
				return newError("exit() takes at most 1 argument, got %d", len(args))
			}
			code := 0
			if len(args) == 1 {
				n, ok := args[0].(*Integer)
				if !ok {
					return newError("exit() argument must be an integer")
				}
				code = int(n.Value)
			}
			os.Exit(code)
			return NULL
		},
	}

	// sleep(ms) - pauses execution for the given number of milliseconds
	builtins["sleep"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("sleep() takes 1 argument, got %d", len(args))
			}
			ms, ok := args[0].(*Integer)
			if !ok {
				return newError("sleep() argument must be an integer (milliseconds)")
			}
			time.Sleep(time.Duration(ms.Value) * time.Millisecond)
			return NULL
		},
	}

	// osname() - returns the current OS name e.g "linux", "darwin", "windows"
	builtins["osname"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("osname() takes no arguments, got %d", len(args))
			}
			return &String{Value: runtime.GOOS}
		},
	}

	// run(command, args...) - runs a command and returns its output, null on failure
	builtins["run"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("run() takes at least 1 argument, got %d", len(args))
			}
			command, ok := args[0].(*String)
			if !ok {
				return newError("run() first argument must be a string")
			}
			cmdArgs := make([]string, len(args)-1)
			for i, arg := range args[1:] {
				cmdArgs[i] = arg.String()
			}
			cmd := exec.Command(command.Value, cmdArgs...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return errResult("%s", string(output))
			}
			return okResult(&String{Value: string(output)})
		},
	}

	// shell(command) - runs a shell command string and returns output, null on failure
	builtins["shell"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("shell() takes 1 argument, got %d", len(args))
			}
			command, ok := args[0].(*String)
			if !ok {
				return newError("shell() argument must be a string")
			}
			cmd := exec.Command("sh", "-c", command.Value)

			output, err := cmd.CombinedOutput()
			if err != nil {
				return errResult("%s", string(output))
			}

			return okResult(&String{Value: string(output)})
		},
	}

	// -------------------------
	// TIME
	// -------------------------

	// timeNow() - returns current unix timestamp in seconds
	builtins["timeNow"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("timeNow() takes no arguments, got %d", len(args))
			}
			return &Integer{Value: time.Now().Unix()}
		},
	}

	// timeMs() - returns current unix timestamp in milliseconds
	builtins["timeMs"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("timeMs() takes no arguments, got %d", len(args))
			}
			return &Integer{Value: time.Now().UnixMilli()}
		},
	}

	// timeStr() - returns current time as string e.g "15:04:05"
	builtins["timeStr"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("timeStr() takes no arguments, got %d", len(args))
			}
			return &String{Value: time.Now().Format("15:04:05")}
		},
	}

	// dateStr() - returns current date as string e.g "2026-02-28"
	builtins["dateStr"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("dateStr() takes no arguments, got %d", len(args))
			}
			return &String{Value: time.Now().Format("2006-01-02")}
		},
	}

	// dateTimeStr() - returns current date and time as string e.g "2026-02-28 15:04:05"
	builtins["dateTimeStr"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("dateTimeStr() takes no arguments, got %d", len(args))
			}
			return &String{Value: time.Now().Format("2006-01-02 15:04:05")}
		},
	}

	// timeFormat(timestamp, format) - formats a unix timestamp using a Go time format string
	builtins["timeFormat"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("timeFormat() takes 2 arguments, got %d", len(args))
			}
			ts, ok := args[0].(*Integer)
			if !ok {
				return newError("timeFormat() first argument must be an integer (unix timestamp)")
			}
			format, ok := args[1].(*String)
			if !ok {
				return newError("timeFormat() second argument must be a string")
			}
			t := time.Unix(ts.Value, 0)
			return &String{Value: t.Format(format.Value)}
		},
	}

	// -------------------------
	// HTTP
	// -------------------------

	builtins["httpGet"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("httpGet() takes 1 or 2 arguments, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return newError("httpGet() first argument must be a string")
			}
			req, err := http.NewRequest("GET", url.Value, nil)
			if err != nil {
				return errResult("httpGet failed to build request: %s", err.Error())
			}
			if len(args) == 2 {
				headers, ok := args[1].(*Table)
				if !ok {
					return newError("httpGet() second argument must be a table")
				}
				for k, v := range headers.Pairs {
					key := strings.TrimPrefix(k, "STRING:")
					if str, ok := v.(*String); ok {
						req.Header.Set(key, str.Value)
					}
				}
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return errResult("httpGet failed: %s", err.Error())
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return errResult("httpGet failed to read body: %s", err.Error())
			}
			pairs := map[string]Object{}
			pairs["STRING:body"] = &String{Value: string(body)}
			pairs["STRING:status"] = &Integer{Value: int64(resp.StatusCode)}
			return okResult(&Table{Pairs: pairs})
		},
	}

	builtins["httpPost"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("httpPost() takes 2 or 3 arguments, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return newError("httpPost() first argument must be a string")
			}
			body, ok := args[1].(*String)
			if !ok {
				return newError("httpPost() second argument must be a string")
			}
			req, err := http.NewRequest("POST", url.Value, strings.NewReader(body.Value))
			if err != nil {
				return errResult("httpPost failed to build request: %s", err.Error())
			}
			req.Header.Set("Content-Type", "application/json")
			if len(args) == 3 {
				headers, ok := args[2].(*Table)
				if !ok {
					return newError("httpPost() third argument must be a table")
				}
				for k, v := range headers.Pairs {
					key := strings.TrimPrefix(k, "STRING:")
					if str, ok := v.(*String); ok {
						req.Header.Set(key, str.Value)
					}
				}
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return errResult("httpPost failed: %s", err.Error())
			}
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return errResult("httpPost failed to read body: %s", err.Error())
			}
			pairs := map[string]Object{}
			pairs["STRING:body"] = &String{Value: string(respBody)}
			pairs["STRING:status"] = &Integer{Value: int64(resp.StatusCode)}
			return okResult(&Table{Pairs: pairs})
		},
	}

	builtins["httpPut"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("httpPut() takes 2 or 3 arguments, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return newError("httpPut() first argument must be a string")
			}
			body, ok := args[1].(*String)
			if !ok {
				return newError("httpPut() second argument must be a string")
			}
			req, err := http.NewRequest("PUT", url.Value, strings.NewReader(body.Value))
			if err != nil {
				return errResult("httpPut failed to build request: %s", err.Error())
			}
			req.Header.Set("Content-Type", "application/json")
			if len(args) == 3 {
				headers, ok := args[2].(*Table)
				if !ok {
					return newError("httpPut() third argument must be a table")
				}
				for k, v := range headers.Pairs {
					key := strings.TrimPrefix(k, "STRING:")
					if str, ok := v.(*String); ok {
						req.Header.Set(key, str.Value)
					}
				}
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return errResult("httpPut failed: %s", err.Error())
			}
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return errResult("httpPut failed to read body: %s", err.Error())
			}
			pairs := map[string]Object{}
			pairs["STRING:body"] = &String{Value: string(respBody)}
			pairs["STRING:status"] = &Integer{Value: int64(resp.StatusCode)}
			return okResult(&Table{Pairs: pairs})
		},
	}

	builtins["httpPatch"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("httpPatch() takes 2 or 3 arguments, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return newError("httpPatch() first argument must be a string")
			}
			body, ok := args[1].(*String)
			if !ok {
				return newError("httpPatch() second argument must be a string")
			}
			req, err := http.NewRequest("PATCH", url.Value, strings.NewReader(body.Value))
			if err != nil {
				return errResult("httpPatch failed to build request: %s", err.Error())
			}
			req.Header.Set("Content-Type", "application/json")
			if len(args) == 3 {
				headers, ok := args[2].(*Table)
				if !ok {
					return newError("httpPatch() third argument must be a table")
				}
				for k, v := range headers.Pairs {
					key := strings.TrimPrefix(k, "STRING:")
					if str, ok := v.(*String); ok {
						req.Header.Set(key, str.Value)
					}
				}
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return errResult("httpPatch failed: %s", err.Error())
			}
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return errResult("httpPatch failed to read body: %s", err.Error())
			}
			pairs := map[string]Object{}
			pairs["STRING:body"] = &String{Value: string(respBody)}
			pairs["STRING:status"] = &Integer{Value: int64(resp.StatusCode)}
			return okResult(&Table{Pairs: pairs})
		},
	}

	builtins["httpDelete"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("httpDelete() takes 1 or 2 arguments, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return newError("httpDelete() first argument must be a string")
			}
			req, err := http.NewRequest("DELETE", url.Value, nil)
			if err != nil {
				return errResult("httpDelete failed to build request: %s", err.Error())
			}
			if len(args) == 2 {
				headers, ok := args[1].(*Table)
				if !ok {
					return newError("httpDelete() second argument must be a table")
				}
				for k, v := range headers.Pairs {
					key := strings.TrimPrefix(k, "STRING:")
					if str, ok := v.(*String); ok {
						req.Header.Set(key, str.Value)
					}
				}
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return errResult("httpDelete failed: %s", err.Error())
			}
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return errResult("httpDelete failed to read body: %s", err.Error())
			}
			pairs := map[string]Object{}
			pairs["STRING:body"] = &String{Value: string(respBody)}
			pairs["STRING:status"] = &Integer{Value: int64(resp.StatusCode)}
			return okResult(&Table{Pairs: pairs})
		},
	}
}

// -------------------------
// JSON HELPERS
// -------------------------

func jsonToObject(v interface{}) Object {
	switch val := v.(type) {
	case nil:
		return NULL
	case bool:
		if val {
			return TRUE
		}
		return FALSE
	case float64:
		if val == float64(int64(val)) {
			return &Integer{Value: int64(val)}
		}
		return &Float{Value: val}
	case string:
		return &String{Value: val}
	case []interface{}:
		items := make([]Object, len(val))
		for i, el := range val {
			items[i] = jsonToObject(el)
		}
		return &Array{Elements: items}
	case map[string]interface{}:
		pairs := map[string]Object{}
		for k, v2 := range val {
			pairs["STRING:"+k] = jsonToObject(v2)
		}
		return &Table{Pairs: pairs}
	default:
		return &String{Value: fmt.Sprintf("%v", val)}
	}
}

func objectToJson(obj Object) interface{} {
	switch val := obj.(type) {
	case *Integer:
		return val.Value
	case *Float:
		return val.Value
	case *Bool:
		return val.Value
	case *String:
		return val.Value
	case *Array:
		items := make([]interface{}, len(val.Elements))
		for i, el := range val.Elements {
			items[i] = objectToJson(el)
		}
		return items
	case *Table:
		m := map[string]interface{}{}
		for k, v := range val.Pairs {
			parts := strings.SplitN(k, ":", 2)
			key := k
			if len(parts) == 2 {
				key = parts[1]
			}
			m[key] = objectToJson(v)
		}
		return m
	default:
		return nil
	}
}

// -------------------------
// MATH HELPER
// -------------------------

// toFloat64 converts an Object to a *float64, returns nil if not a number
func toFloat64(obj Object) *float64 {
	switch v := obj.(type) {
	case *Integer:
		f := float64(v.Value)
		return &f
	case *Float:
		return &v.Value
	default:
		return nil
	}
}
