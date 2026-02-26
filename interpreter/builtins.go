package interpreter

import (
	"fmt"
	"strings"
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
}
