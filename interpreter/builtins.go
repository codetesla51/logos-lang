package interpreter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var builtins = map[string]*Builtin{}

func init() {
	builtins["print"] = &Builtin{
		Fn: func(args ...Object) Object {
			var out strings.Builder
			for i, arg := range args {
				if i > 0 {
					out.WriteString(" ")
				}
				out.WriteString(arg.String())
			}
			fmt.Println(out.String())
			return NULL
		},
	}

	builtins["type"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("type() takes exactly 1 argument, got %d", len(args))
			}
			return &String{Value: string(args[0].Type())}
		},
	}

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

	// -------------------------
	// FILE OPS
	// -------------------------

	// read(filepath) - reads file, returns null if not found
	builtins["read"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("read() takes exactly 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("read() argument must be a string (filepath)")
			}
			data, err := os.ReadFile(path.Value)
			if err != nil {
				return NULL
			}
			return &String{Value: string(data)}
		},
	}

	// write(filepath, content) - overwrites file with content
	builtins["write"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("write() takes 2 arguments, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("write() first argument must be a string (filepath)")
			}
			content, ok := args[1].(*String)
			if !ok {
				return newError("write() second argument must be a string")
			}
			err := os.WriteFile(path.Value, []byte(content.Value), 0644)
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// append(filepath, content) - appends content to a file
	builtins["append"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("append() takes 2 arguments, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("append() first argument must be a string (filepath)")
			}
			content, ok := args[1].(*String)
			if !ok {
				return newError("append() second argument must be a string")
			}
			f, err := os.OpenFile(path.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return NULL
			}
			defer f.Close()
			_, err = f.WriteString(content.Value)
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// exists(path) - returns true if file or folder exists, false otherwise
	builtins["exists"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("exists() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("exists() argument must be a string (filepath)")
			}
			_, err := os.Stat(path.Value)
			if os.IsNotExist(err) {
				return FALSE
			}
			return TRUE
		},
	}

	// delete(path) - deletes a file or empty folder, returns null on failure
	builtins["delete"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("delete() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("delete() argument must be a string (filepath)")
			}
			err := os.Remove(path.Value)
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// deleteAll(path) - deletes folder and everything inside, returns null on failure
	builtins["deleteAll"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("deleteAll() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("deleteAll() argument must be a string (folderpath)")
			}
			err := os.RemoveAll(path.Value)
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// rename(oldpath, newpath) - renames or moves a file or folder
	builtins["rename"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("rename() takes 2 arguments, got %d", len(args))
			}
			oldPath, ok := args[0].(*String)
			if !ok {
				return newError("rename() first argument must be a string")
			}
			newPath, ok := args[1].(*String)
			if !ok {
				return newError("rename() second argument must be a string")
			}
			err := os.Rename(oldPath.Value, newPath.Value)
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// mkdir(path) - creates folder and any missing parents
	builtins["mkdir"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("mkdir() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("mkdir() argument must be a string (folderpath)")
			}
			err := os.MkdirAll(path.Value, 0755)
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// rmdir(path) - removes an empty directory
	builtins["rmdir"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("rmdir() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("rmdir() argument must be a string (folderpath)")
			}
			err := os.Remove(path.Value)
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// readDir(path) - returns array of file/folder names in directory, null on failure
	builtins["readDir"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("readDir() takes 1 argument, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("readDir() argument must be a string (folderpath)")
			}
			entries, err := os.ReadDir(path.Value)
			if err != nil {
				return NULL
			}
			items := make([]Object, len(entries))
			for i, entry := range entries {
				items[i] = &String{Value: entry.Name()}
			}
			return &Array{Elements: items}
		},
	}

	// cp(src, dst) - copies a file from src to dst
	builtins["cp"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("cp() takes 2 arguments, got %d", len(args))
			}
			src, ok := args[0].(*String)
			if !ok {
				return newError("cp() first argument must be a string (source path)")
			}
			dst, ok := args[1].(*String)
			if !ok {
				return newError("cp() second argument must be a string (destination path)")
			}
			data, err := os.ReadFile(src.Value)
			if err != nil {
				return NULL
			}
			err = os.WriteFile(dst.Value, data, 0644)
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// mv(src, dst) - moves a file or folder from src to dst
	builtins["mv"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("mv() takes 2 arguments, got %d", len(args))
			}
			src, ok := args[0].(*String)
			if !ok {
				return newError("mv() first argument must be a string (source path)")
			}
			dst, ok := args[1].(*String)
			if !ok {
				return newError("mv() second argument must be a string (destination path)")
			}
			err := os.Rename(src.Value, dst.Value)
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// chmod(path, mode) - changes file permissions, mode is octal string e.g "0755"
	builtins["chmod"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("chmod() takes 2 arguments, got %d", len(args))
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("chmod() first argument must be a string (filepath)")
			}
			modeStr, ok := args[1].(*String)
			if !ok {
				return newError("chmod() second argument must be a string (e.g \"0755\")")
			}
			mode, err := strconv.ParseUint(modeStr.Value, 8, 32)
			if err != nil {
				return newError("chmod() invalid mode \"%s\", use octal like \"0755\"", modeStr.Value)
			}
			err = os.Chmod(path.Value, os.FileMode(mode))
			if err != nil {
				return NULL
			}
			return NULL
		},
	}

	// glob(pattern) - returns array of paths matching pattern, null on error
	builtins["glob"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("glob() takes 1 argument, got %d", len(args))
			}
			pattern, ok := args[0].(*String)
			if !ok {
				return newError("glob() argument must be a string (pattern)")
			}
			matches, err := filepath.Glob(pattern.Value)
			if err != nil {
				return NULL
			}
			items := make([]Object, len(matches))
			for i, m := range matches {
				items[i] = &String{Value: m}
			}
			return &Array{Elements: items}
		},
	}

	// -------------------------
	// STRING OPS
	// -------------------------

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

	// -------------------------
	// TYPE CONVERSION
	// -------------------------

	builtins["int"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("int() takes 1 argument, got %d", len(args))
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

	builtins["float"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("float() takes 1 argument, got %d", len(args))
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

	builtins["bool"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("bool() takes 1 argument, got %d", len(args))
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

	builtins["str"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("str() takes 1 argument, got %d", len(args))
			}
			return &String{Value: args[0].String()}
		},
	}

	// -------------------------
	// ARRAY OPS
	// -------------------------

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

	// -------------------------
	// TABLE OPS
	// -------------------------

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
				items = append(items, &String{Value: k})
			}
			return &Array{Elements: items}
		},
	}

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
			_, exists := table.Pairs[key.Value]
			if exists {
				return TRUE
			}
			return FALSE
		},
	}

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
				if k != key.Value {
					newPairs[k] = v
				}
			}
			return &Table{Pairs: newPairs}
		},
	}

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

	// parseJson(string) - parses a JSON string into a table/array/value, null on failure
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
				return NULL
			}
			return jsonToObject(raw)
		},
	}

	// toJson(value) - converts a table/array/value to a JSON string, null on failure
	builtins["toJson"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("toJson() takes 1 argument, got %d", len(args))
			}
			raw := objectToJson(args[0])
			data, err := json.Marshal(raw)
			if err != nil {
				return NULL
			}
			return &String{Value: string(data)}
		},
	}

	// -------------------------
	// OS / SYSTEM
	// -------------------------

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

	builtins["args"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("args() takes no arguments, got %d", len(args))
			}
			osArgs := os.Args
			items := make([]Object, len(osArgs))
			for i, arg := range osArgs {
				items[i] = &String{Value: arg}
			}
			return &Array{Elements: items}
		},
	}

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

	builtins["osname"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("osname() takes no arguments, got %d", len(args))
			}
			return &String{Value: runtime.GOOS}
		},
	}

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
				return NULL
			}
			return &String{Value: string(output)}
		},
	}

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
				return NULL
			}
			return &String{Value: string(output)}
		},
	}

	// -------------------------
	// TIME
	// -------------------------

	builtins["timeNow"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("timeNow() takes no arguments, got %d", len(args))
			}
			return &Integer{Value: time.Now().Unix()}
		},
	}

	builtins["timeMs"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("timeMs() takes no arguments, got %d", len(args))
			}
			return &Integer{Value: time.Now().UnixMilli()}
		},
	}

	builtins["timeStr"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("timeStr() takes no arguments, got %d", len(args))
			}
			return &String{Value: time.Now().Format("15:04:05")}
		},
	}

	builtins["dateStr"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("dateStr() takes no arguments, got %d", len(args))
			}
			return &String{Value: time.Now().Format("2006-01-02")}
		},
	}

	builtins["dateTimeStr"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("dateTimeStr() takes no arguments, got %d", len(args))
			}
			return &String{Value: time.Now().Format("2006-01-02 15:04:05")}
		},
	}

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
			if len(args) != 1 {
				return newError("httpGet() takes 1 argument, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return newError("httpGet() argument must be a string")
			}
			resp, err := http.Get(url.Value)
			if err != nil {
				return NULL
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return NULL
			}
			pairs := map[string]Object{}
			pairs["body"] = &String{Value: string(body)}
			pairs["status"] = &Integer{Value: int64(resp.StatusCode)}
			return &Table{Pairs: pairs}
		},
	}

	builtins["httpPost"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("httpPost() takes 2 arguments, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return newError("httpPost() first argument must be a string")
			}
			body, ok := args[1].(*String)
			if !ok {
				return newError("httpPost() second argument must be a string")
			}
			resp, err := http.Post(url.Value, "application/json", strings.NewReader(body.Value))
			if err != nil {
				return NULL
			}
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return NULL
			}
			pairs := map[string]Object{}
			pairs["body"] = &String{Value: string(respBody)}
			pairs["status"] = &Integer{Value: int64(resp.StatusCode)}
			return &Table{Pairs: pairs}
		},
	}

	builtins["httpPatch"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("httpPatch() takes 2 arguments, got %d", len(args))
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
				return NULL
			}
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return NULL
			}
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return NULL
			}
			pairs := map[string]Object{}
			pairs["body"] = &String{Value: string(respBody)}
			pairs["status"] = &Integer{Value: int64(resp.StatusCode)}
			return &Table{Pairs: pairs}
		},
	}

	builtins["httpDelete"] = &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("httpDelete() takes 1 argument, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return newError("httpDelete() argument must be a string")
			}
			req, err := http.NewRequest("DELETE", url.Value, nil)
			if err != nil {
				return NULL
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return NULL
			}
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return NULL
			}
			pairs := map[string]Object{}
			pairs["body"] = &String{Value: string(respBody)}
			pairs["status"] = &Integer{Value: int64(resp.StatusCode)}
			return &Table{Pairs: pairs}
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
		// if it's a whole number, return as integer
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
			pairs[k] = jsonToObject(v2)
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
			m[k] = objectToJson(v)
		}
		return m
	default:
		return nil
	}
}
