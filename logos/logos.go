// Package logos provides a clean embedding API for the Logos scripting language.
//
// Example usage:
//
//	vm := logos.New()
//	vm.Register("greet", func(args ...logos.Object) logos.Object {
//	    name := args[0].(*logos.String).Value
//	    return &logos.String{Value: "Hello, " + name}
//	})
//	vm.SetVar("count", 42)
//	err := vm.Run(`print(greet("World"))`)
package logos

import (
	"github.com/codetesla51/logos/interpreter"
)

// Re-export types from interpreter for convenience
type (
	Object      = interpreter.Object
	Integer     = interpreter.Integer
	Float       = interpreter.Float
	String      = interpreter.String
	Bool        = interpreter.Bool
	Array       = interpreter.Array
	Table       = interpreter.Table
	Null        = interpreter.Null
	Function    = interpreter.Function
	Builtin     = interpreter.Builtin
	BuiltinFunc = interpreter.BuiltinFunc
)

// SandboxConfig controls which built-in capabilities are available to scripts.
type SandboxConfig struct {
	AllowFileIO  bool
	AllowNetwork bool
	AllowShell   bool
	AllowExit    bool
}

// VM represents a Logos virtual machine instance.
type VM struct {
	interp *interpreter.Interpreter
}

// New creates a new Logos VM with all capabilities enabled.
func New() *VM {
	return &VM{
		interp: interpreter.NewInterpreter(),
	}
}

// NewWithConfig creates a new Logos VM with the specified sandbox configuration.
func NewWithConfig(config SandboxConfig) *VM {
	return &VM{
		interp: interpreter.NewInterpreter(interpreter.SandboxConfig{
			AllowFileIO:  config.AllowFileIO,
			AllowNetwork: config.AllowNetwork,
			AllowShell:   config.AllowShell,
			AllowExit:    config.AllowExit,
		}),
	}
}

// Register registers a Go function callable from Logos scripts.
func (vm *VM) Register(name string, fn BuiltinFunc) {
	vm.interp.Register(name, fn)
}

// SetVar sets a variable in the Logos environment.
// Supported Go types: int, int64, float64, string, bool, []interface{}, map[string]interface{}, nil
func (vm *VM) SetVar(name string, val interface{}) {
	vm.interp.SetVar(name, val)
}

// GetVar gets a variable from the Logos environment.
// Returns nil if not found.
func (vm *VM) GetVar(name string) interface{} {
	return vm.interp.GetVar(name)
}

// Call calls a Logos function by name with the given arguments.
// Returns the result converted to a Go value, or an error.
func (vm *VM) Call(name string, args ...interface{}) (interface{}, error) {
	return vm.interp.Call(name, args...)
}

// Run evaluates a Logos script string.
// Returns an error if parsing fails or the script produces an error.
func (vm *VM) Run(source string) error {
	return vm.interp.Run(source)
}
