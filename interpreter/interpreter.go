package interpreter

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"

	"github.com/codetesla51/golexer/golexer"
	"github.com/codetesla51/logos/parser"
)

type ObjectType string
type BuiltinFunc func(args ...Object) Object

const (
	INTEGER_OBJ      = "INTEGER"
	FLOAT_OBJ        = "FLOAT"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
	TABLE_OBJ        = "TABLE"
	CONTINUE_OBJ     = "CONTINUE"
	BREAK_OBJ        = "BREAK"
)

// singletons
var (
	TRUE  = &Bool{Value: true}
	FALSE = &Bool{Value: false}
	NULL  = &Null{}
)

type Object interface {
	Type() ObjectType
	String() string // testing and debugging
}
type Environment struct {
	store map[string]Object
	outer *Environment
}

// SandboxConfig controls which built-in capabilities are available to scripts.
type SandboxConfig struct {
	AllowFileIO  bool
	AllowNetwork bool
	AllowShell   bool
	AllowExit    bool
}

type Interpreter struct {
	Env              *Environment
	ModuleCache      map[string]*Environment
	StdFs            fs.FS
	CurrentFile      string // currently executing file for error reporting
	Config           SandboxConfig
	instanceBuiltins map[string]*Builtin // per-instance builtins, checked before global
}
type Integer struct {
	Value int64
}
type Float struct {
	Value float64
}
type Bool struct {
	Value bool
}
type Array struct {
	Elements []Object
}
type Null struct{}
type ReturnValue struct {
	Value Object
}
type ContinueSignal struct{}
type BreakSignal struct{}
type Error struct {
	Message string
	Line    int
	Column  int
	File    string
}

type Function struct {
	Parameters []*parser.Identifier
	Body       *parser.BlockStatement
	Env        *Environment
}

type String struct {
	Value string
}
type Builtin struct {
	Fn BuiltinFunc
}
type Table struct {
	Pairs map[string]Object
}

func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ

}
func (i *Integer) String() string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%d", i.Value))
	return out.String()
}
func (f *Float) Type() ObjectType {
	return FLOAT_OBJ
}
func (f *Float) String() string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%g", f.Value))
	return out.String()
}
func (b *Bool) Type() ObjectType {
	return BOOLEAN_OBJ
}
func (b *Bool) String() string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%t", b.Value))
	return out.String()
}
func (n *Null) Type() ObjectType {
	return NULL_OBJ
}
func (n *Null) String() string {
	return "null"
}
func (rv *ReturnValue) Type() ObjectType {
	return RETURN_VALUE_OBJ
}
func (rv *ReturnValue) String() string {
	return rv.Value.String()
}
func (e *Error) Type() ObjectType {
	return ERROR_OBJ
}

func (c *ContinueSignal) Type() ObjectType { return CONTINUE_OBJ }
func (b *BreakSignal) Type() ObjectType    { return BREAK_OBJ }
func (c *ContinueSignal) String() string   { return "continue" }
func (b *BreakSignal) String() string      { return "break" }
func (e *Error) String() string {
	// if file is known, prefer concise file:line:column: message format
	if e.File != "" && e.Line > 0 {
		return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
	}
	if e.Line > 0 {
		return fmt.Sprintf("ERROR [%d:%d]: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("ERROR: %s", e.Message)
}
func (f *Function) Type() ObjectType {
	return FUNCTION_OBJ
}
func (f *Function) String() string {
	var out strings.Builder
	out.WriteString("fn(")
	for i, p := range f.Parameters {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(p.String())
		out.WriteString(") {\n")
		out.WriteString(f.Body.String())
		out.WriteString("\n}")
	}
	return out.String()
}

func (s *String) Type() ObjectType {
	return STRING_OBJ
}
func (s *String) String() string {
	return s.Value
}
func (a *Array) Type() ObjectType {
	return ARRAY_OBJ
}
func (a *Array) String() string {
	var out strings.Builder
	out.WriteString("[")
	for i, e := range a.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(e.String())
	}
	out.WriteString("]")
	return out.String()
}
func (bi *Builtin) Type() ObjectType {
	return BUILTIN_OBJ
}
func (bi *Builtin) String() string {
	return "builtin function"
}
func (t *Table) Type() ObjectType { return TABLE_OBJ }
func (t *Table) String() string {
	var out strings.Builder
	out.WriteString("table{")
	for k, v := range t.Pairs {
		out.WriteString(k + ": " + v.String() + ", ")
	}
	out.WriteString("}")
	return out.String()
}

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Object), outer: nil}
}
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return obj, ok
}
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
func (e *Environment) Update(name string, val Object) Object {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return val
	}
	if e.outer != nil {
		return e.outer.Update(name, val)
	}
	return nil
}

// NewInterpreter creates a new interpreter and sets the default token path.
// Optional arguments: first fs.FS for StdFs, second SandboxConfig for sandboxing.
// If no SandboxConfig is provided, all capabilities are enabled by default.
func NewInterpreter(args ...interface{}) *Interpreter {
	// Default config allows everything
	config := SandboxConfig{
		AllowFileIO:  true,
		AllowNetwork: true,
		AllowShell:   true,
		AllowExit:    true,
	}

	var stdFs fs.FS

	// Parse variadic arguments
	for _, arg := range args {
		switch v := arg.(type) {
		case fs.FS:
			stdFs = v
		case SandboxConfig:
			config = v
		}
	}

	i := &Interpreter{
		Env:              NewEnvironment(),
		ModuleCache:      make(map[string]*Environment),
		Config:           config,
		instanceBuiltins: make(map[string]*Builtin),
	}

	if stdFs != nil {
		i.StdFs = stdFs
	}

	return i
}

// isBuiltinAllowed checks if a builtin function is allowed by the sandbox config.
func (i *Interpreter) isBuiltinAllowed(name string) bool {
	// File I/O functions
	fileIOFuncs := map[string]bool{
		"fileRead": true, "fileWrite": true, "fileAppend": true,
		"fileDelete": true, "fileDeleteAll": true, "fileCopy": true,
		"fileMove": true, "fileMkdir": true, "fileRmdir": true,
		"fileRename": true, "fileReadDir": true, "fileGlob": true,
		"fileChmod": true, "fileExt": true, "fileExists": true,
	}
	if !i.Config.AllowFileIO && fileIOFuncs[name] {
		return false
	}

	// Network functions
	networkFuncs := map[string]bool{
		"httpGet": true, "httpPost": true, "httpPatch": true, "httpDelete": true, "httpPut": true,
	}
	if !i.Config.AllowNetwork && networkFuncs[name] {
		return false
	}

	// Shell functions
	shellFuncs := map[string]bool{
		"shell": true, "run": true,
	}
	if !i.Config.AllowShell && shellFuncs[name] {
		return false
	}

	// Exit function
	if !i.Config.AllowExit && name == "exit" {
		return false
	}

	return true
}

// Register registers a Go function callable from Logos scripts.
// The function is added to this interpreter instance's local builtins map,
// so different interpreter instances can have different functions registered.
func (i *Interpreter) Register(name string, fn BuiltinFunc) {
	i.instanceBuiltins[name] = &Builtin{Fn: fn}
}

// SetVar converts a native Go value to a Logos Object and sets it in the interpreter environment.
// Supported Go types: int, int64, float64, string, bool, []interface{}, map[string]interface{}, nil
func (i *Interpreter) SetVar(name string, val interface{}) {
	obj := GoToObject(val)
	i.Env.Set(name, obj)
}

// GetVar gets a variable from the interpreter environment and converts it back to a native Go value.
// Returns nil if not found.
func (i *Interpreter) GetVar(name string) interface{} {
	obj, ok := i.Env.Get(name)
	if !ok {
		return nil
	}
	return ObjectToGo(obj)
}

// Call looks up a function defined in the Logos script by name, converts the Go args to Logos Objects,
// calls it, converts the result back to a Go value and returns it.
// Returns an error if the function is not found or if evaluation returns an Error object.
func (i *Interpreter) Call(name string, args ...interface{}) (interface{}, error) {
	// Look up the function in the environment
	obj, ok := i.Env.Get(name)
	if !ok {
		// Also check instance builtins
		if builtin, ok := i.instanceBuiltins[name]; ok {
			obj = builtin
		} else if builtin, ok := builtins[name]; ok {
			if !i.isBuiltinAllowed(name) {
				return nil, fmt.Errorf("function '%s' is not available in sandbox mode", name)
			}
			obj = builtin
		} else {
			return nil, fmt.Errorf("function not found: %s", name)
		}
	}

	// Convert Go args to Logos Objects
	logoArgs := make([]Object, len(args))
	for i, arg := range args {
		logoArgs[i] = GoToObject(arg)
	}

	// Call the function
	var result Object
	switch fn := obj.(type) {
	case *Function:
		extendedEnv := i.extendFunctionEnv(fn, logoArgs)
		result = i.Eval(fn.Body, extendedEnv)
		result = i.unwrapReturnValue(result)
	case *Builtin:
		result = fn.Fn(logoArgs...)
	default:
		return nil, fmt.Errorf("'%s' is not a function", name)
	}

	// Check for errors
	if err, ok := result.(*Error); ok {
		return nil, fmt.Errorf("%s", err.String())
	}

	return ObjectToGo(result), nil
}

// Run evaluates a Logos source string.
// Uses defer recover() to catch any panics and return them as errors instead of crashing the host.
// Returns an error if the script produces an Error object.
func (i *Interpreter) Run(source string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during evaluation: %v", r)
		}
	}()

	lexer := golexer.NewLexer(source)
	p := parser.NewParser(lexer)
	program := p.Parse()

	if len(p.Errors()) != 0 {
		return fmt.Errorf("parse errors: %v", p.Errors())
	}

	result := i.Eval(program, i.Env)
	if result != nil {
		if errObj, ok := result.(*Error); ok {
			return fmt.Errorf("%s", errObj.String())
		}
	}

	return nil
}

// isTruthy returns whether an object counts as true in conditionals
func isTruthy(obj Object) bool {
	switch obj {
	case TRUE:
		return true
	case FALSE:
		return false
	case NULL:
		return false
	default:
		return true
	}
}
func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func newErrorAt(line, col int, format string, a ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, a...),
		Line:    line,
		Column:  col,
	}
}

// fileError creates an error that includes the interpreter's current file context.
func (i *Interpreter) fileError(line, col int, format string, a ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, a...),
		Line:    line,
		Column:  col,
		File:    i.CurrentFile,
	}
}
func isError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}

// Eval dispatches AST nodes to their evaluator implementations
func (i *Interpreter) Eval(node parser.Node, env *Environment) Object {
	switch node := node.(type) {
	case *parser.Program:
		return i.evalProgram(node, env)
	case *parser.ExpressionStatement:
		return i.Eval(node.Expression, env)
	case *parser.IntegerLiteral:
		return &Integer{Value: node.Value}
	case *parser.FloatLiteral:
		return &Float{Value: node.Value}
	case *parser.BooleanLiteral:
		if node.Value {
			return TRUE
		}
		return FALSE
	case *parser.StringLiteral:
		return &String{Value: node.Value}
	case *parser.NullExpression:
		return NULL
	case *parser.Identifier:
		return i.evalIdentifier(node, env)
	case *parser.LetStatement:
		return i.evalLetStatement(node, env)
	case *parser.InfixExpression:
		return i.evalInfixExpression(node, env)
	case *parser.ReturnStatement:
		return i.evalReturnStatement(node, env)
	case *parser.BlockStatement:
		return i.evalBlockStatment(node, env)
	case *parser.IfExpression:
		return i.evalIfExpression(node, env)
	case *parser.FunctionLiteral:
		return i.evalFunctionLiteral(node, env)
	case *parser.CallExpression:
		return i.evalFunctionCalls(node, env)
	case *parser.PrefixExpression:
		return i.evalPrefixExpression(node, env)
	case *parser.ArrayLiteral:
		return i.evalArrayLiteral(node, env)
	case *parser.ArrayIndexExpression:
		return i.evalArrayIndexExpression(node, env)
	case *parser.ForStatement:
		return i.evalForStatement(node, env)
	case *parser.BreakStatement:
		return &BreakSignal{}
	case *parser.ContinueStatement:
		return &ContinueSignal{}
	case *parser.SwitchStatement:
		return i.evalSwitchStatement(node, env)
	case *parser.ForInStatement:
		return i.evalForInStatement(node, env)
	case *parser.TableLiteral:
		return i.evalTableLiteral(node, env)
	case *parser.UseStatement:
		return i.evalUseStatement(node, env)
	case *parser.DotExpression:
		return i.evalDotExpression(node, env)
	case *parser.SpawnStatment:
		return i.evalSpawnStatment(node, env)
	case *parser.SpawnForInStatement:
		return i.evalSpawnForInStatement(node, env)
	case *parser.TenaryExpression:
		return i.evalTernary(node, env)
	case *parser.InterpolatedString:
		return i.evalInterpol(node, env)
	case *parser.TryExpression:
		return i.evalTry(node, env)
	default:
		fmt.Printf("unknown node: %T %+v\n", node, node)
		return newError("unknown node type %T", node)
	}
}
func (i *Interpreter) evalProgram(program *parser.Program, env *Environment) Object {
	var result Object
	for _, statement := range program.Statements {
		result = i.Eval(statement, env)
		switch result := result.(type) {
		case *ReturnValue:
			return result.Value
		case *Error:
			return result
		}
	}
	return result
}
func (i *Interpreter) evalIdentifier(node *parser.Identifier, env *Environment) Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	// Check instance-specific builtins first
	if builtin, ok := i.instanceBuiltins[node.Value]; ok {
		return builtin
	}
	// Check global builtins, but respect sandbox config
	if builtin, ok := builtins[node.Value]; ok {
		if !i.isBuiltinAllowed(node.Value) {
			return i.fileError(node.Token.Line, node.Token.Column,
				"function '%s' is not available in sandbox mode", node.Value)
		}
		return builtin
	}
	return i.fileError(node.Token.Line, node.Token.Column, "identifier not found: %s", node.Value)
}
func (i *Interpreter) evalLetStatement(node *parser.LetStatement, env *Environment) Object {
	val := i.Eval(node.Value, env)
	if isError(val) {
		return val
	}
	env.Set(node.Name.Value, val)
	return val
}

// evalInfixExpression evaluates binary and assignment operators, including compound assignments
func (i *Interpreter) evalInfixExpression(node *parser.InfixExpression, env *Environment) Object {
	if node.Operator == "=" {
		switch left := node.Left.(type) {
		case *parser.Identifier:
			val := i.Eval(node.Right, env)
			if isError(val) {
				return val
			}
			result := env.Update(left.Value, val)
			if result == nil {
				return i.fileError(node.Token.Line, node.Token.Column, "cannot assign to undeclared variable: %s", left.Value)
			}
			return val
		case *parser.DotExpression:
			val := i.Eval(node.Right, env)
			if isError(val) {
				return val
			}
			obj := i.Eval(left.Left, env)
			if isError(obj) {
				return obj
			}
			table, ok := obj.(*Table)
			if !ok {
				return i.fileError(node.Token.Line, node.Token.Column,
					"dot assignment only supported on tables, got: %s", obj.Type())
			}
			key := fmt.Sprintf("%s:%s", STRING_OBJ, left.Right.Value)
			table.Pairs[key] = val
			return val
		case *parser.ArrayIndexExpression:
			val := i.Eval(node.Right, env)
			if isError(val) {
				return val
			}
			obj := i.Eval(left.Array, env)
			if isError(obj) {
				return obj
			}
			index := i.Eval(left.Index, env)
			if isError(index) {
				return index
			}

			switch collection := obj.(type) {
			case *Array:
				idx, ok := index.(*Integer)
				if !ok {
					return newErrorAt(node.Token.Line, node.Token.Column, "array index must be an integer")
				}
				if idx.Value < 0 || idx.Value >= int64(len(collection.Elements)) {
					return newErrorAt(node.Token.Line, node.Token.Column, "index out of bounds: %d", idx.Value)
				}
				collection.Elements[idx.Value] = val
				return val
			case *Table:
				key := fmt.Sprintf("%s:%s", index.Type(), index.String())
				collection.Pairs[key] = val
				return val

			default:
				return i.fileError(node.Token.Line, node.Token.Column, "cannot index assign on type: %s", obj.Type())
			}

		default:
			return i.fileError(node.Token.Line, node.Token.Column, "cannot assign to non-identifier")
		}
	}
	if node.Operator == "+=" {
		ident, ok := node.Left.(*parser.Identifier)
		if !ok {
			return newErrorAt(node.Token.Line, node.Token.Column, "cannot assign to non-identifier")
		}
		currentVal, exists := env.Get(ident.Value)
		if !exists {
			return i.fileError(node.Token.Line, node.Token.Column, "cannot assign to undeclared variable: %s", ident.String())
		}
		rightVal := i.Eval(node.Right, env)
		if isError(rightVal) {
			return rightVal
		}
		// Compute the result based on types
		var result Object
		switch {
		case currentVal.Type() == INTEGER_OBJ && rightVal.Type() == FLOAT_OBJ:
			l := float64(currentVal.(*Integer).Value)
			result = i.evalFloatInfixExpression("+", &Float{Value: l}, rightVal, node.Token.Line, node.Token.Column)
		case currentVal.Type() == FLOAT_OBJ && rightVal.Type() == INTEGER_OBJ:
			r := float64(rightVal.(*Integer).Value)
			result = i.evalFloatInfixExpression("+", currentVal, &Float{Value: r}, node.Token.Line, node.Token.Column)
		case currentVal.Type() == INTEGER_OBJ && rightVal.Type() == INTEGER_OBJ:
			result = i.evalIntegerInfixExpression("+", currentVal, rightVal, node.Token.Line, node.Token.Column)
		case currentVal.Type() == FLOAT_OBJ && rightVal.Type() == FLOAT_OBJ:
			result = i.evalFloatInfixExpression("+", currentVal, rightVal, node.Token.Line, node.Token.Column)
		case currentVal.Type() == STRING_OBJ && rightVal.Type() == STRING_OBJ:
			result = i.evalStringInfixExpression("+", currentVal, rightVal, node.Token.Line, node.Token.Column)
		default:
			return i.fileError(node.Token.Line, node.Token.Column, "cannot use += with types %s and %s", currentVal.Type(), rightVal.Type())
		}
		if isError(result) {
			return result
		}
		env.Update(ident.Value, result)
		return result
	}
	if node.Operator == "-=" {
		ident, ok := node.Left.(*parser.Identifier)
		if !ok {
			return newErrorAt(node.Token.Line, node.Token.Column, "cannot assign to non-identifier")
		}
		currentVal, exists := env.Get(ident.Value)
		if !exists {
			return i.fileError(node.Token.Line, node.Token.Column, "cannot assign to undeclared variable: %s", ident.String())
		}
		rightVal := i.Eval(node.Right, env)
		if isError(rightVal) {
			return rightVal
		}
		// Compute the result based on types
		var result Object
		switch {
		case currentVal.Type() == INTEGER_OBJ && rightVal.Type() == FLOAT_OBJ:
			l := float64(currentVal.(*Integer).Value)
			result = i.evalFloatInfixExpression("-", &Float{Value: l}, rightVal, node.Token.Line, node.Token.Column)
		case currentVal.Type() == FLOAT_OBJ && rightVal.Type() == INTEGER_OBJ:
			r := float64(rightVal.(*Integer).Value)
			result = i.evalFloatInfixExpression("-", currentVal, &Float{Value: r}, node.Token.Line, node.Token.Column)
		case currentVal.Type() == INTEGER_OBJ && rightVal.Type() == INTEGER_OBJ:
			result = i.evalIntegerInfixExpression("-", currentVal, rightVal, node.Token.Line, node.Token.Column)
		case currentVal.Type() == FLOAT_OBJ && rightVal.Type() == FLOAT_OBJ:
			result = i.evalFloatInfixExpression("-", currentVal, rightVal, node.Token.Line, node.Token.Column)
		default:
			return i.fileError(node.Token.Line, node.Token.Column, "cannot use -= with types %s and %s", currentVal.Type(), rightVal.Type())
		}
		if isError(result) {
			return result
		}
		env.Update(ident.Value, result)
		return result
	}
	left := i.Eval(node.Left, env)
	if isError(left) {
		return left
	}
	right := i.Eval(node.Right, env)
	if isError(right) {
		return right
	}
	switch {
	case left.Type() == INTEGER_OBJ && right.Type() == FLOAT_OBJ:
		l := float64(left.(*Integer).Value)
		return i.evalFloatInfixExpression(node.Operator, &Float{Value: l}, right, node.Token.Line, node.Token.Column)
	case left.Type() == FLOAT_OBJ && right.Type() == INTEGER_OBJ:
		r := float64(right.(*Integer).Value)
		return i.evalFloatInfixExpression(node.Operator, left, &Float{Value: r}, node.Token.Line, node.Token.Column)
	case left.Type() == INTEGER_OBJ && right.Type() == INTEGER_OBJ:
		return i.evalIntegerInfixExpression(node.Operator, left, right, node.Token.Line, node.Token.Column)
	case left.Type() == FLOAT_OBJ && right.Type() == FLOAT_OBJ:
		return i.evalFloatInfixExpression(node.Operator, left, right, node.Token.Line, node.Token.Column)
	case left.Type() == STRING_OBJ && right.Type() == STRING_OBJ:
		return i.evalStringInfixExpression(node.Operator, left, right, node.Token.Line, node.Token.Column)
	case node.Operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case node.Operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case node.Operator == "&&":
		if !isTruthy(left) {
			return FALSE
		}
		right = i.Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return nativeBoolToBooleanObject(isTruthy(right))
	case node.Operator == "||":
		if isTruthy(left) {
			return TRUE
		}
		right = i.Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return nativeBoolToBooleanObject(isTruthy(right))
	case left.Type() != right.Type():
		return i.fileError(node.Token.Line, node.Token.Column,
			"type mismatch: %s %s %s", left.Type(), node.Operator, right.Type())
	default:
		return i.fileError(node.Token.Line, node.Token.Column,
			"unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
	}
}
func (i *Interpreter) evalIntegerInfixExpression(operator string, left, right Object, line, col int) Object {
	rightVal := right.(*Integer).Value
	leftVal := left.(*Integer).Value
	switch operator {
	case "+":
		return &Integer{Value: leftVal + rightVal}
	case "-":
		return &Integer{Value: leftVal - rightVal}
	case "*":
		return &Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newErrorAt(line, col, "division by zero")
		}
		return &Integer{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newErrorAt(line, col, "modulo by zero")
		}
		return &Integer{Value: leftVal % rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	default:
		return i.fileError(line, col, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}
func (i *Interpreter) evalFloatInfixExpression(operator string, left, right Object, line, col int) Object {
	rightVal := right.(*Float).Value
	leftVal := left.(*Float).Value
	switch operator {
	case "+":
		return &Float{Value: leftVal + rightVal}
	case "-":
		return &Float{Value: leftVal - rightVal}
	case "*":
		return &Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newErrorAt(line, col, "division by zero")
		}
		return &Float{Value: leftVal / rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)

	default:
		return i.fileError(line, col, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func nativeBoolToBooleanObject(input bool) *Bool {
	if input {
		return TRUE
	}
	return FALSE
}
func (i *Interpreter) evalReturnStatement(node *parser.ReturnStatement, env *Environment) Object {
	val := i.Eval(node.ReturnValue, env)
	if isError(val) {
		return val
	}
	return &ReturnValue{Value: val}
}
func (i *Interpreter) evalStringInfixExpression(operator string, left, right Object, line, col int) Object {
	leftVal := left.(*String).Value
	rightVal := right.(*String).Value
	switch operator {
	case "+":
		return &String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return i.fileError(line, col, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}
func (i *Interpreter) evalBlockStatment(block *parser.BlockStatement, env *Environment) Object {
	var result Object
	for _, statement := range block.Statements {
		result = i.Eval(statement, env)
		if result != nil {
			rt := result.Type()
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ || rt == CONTINUE_OBJ || rt == BREAK_OBJ {
				return result
			}
		}
	}
	return result
}
func (i *Interpreter) evalIfExpression(ie *parser.IfExpression, env *Environment) Object {
	condition := i.Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return i.Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return i.Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}
func (i *Interpreter) evalFunctionLiteral(node *parser.FunctionLiteral, env *Environment) Object {
	params := node.Parameters
	body := node.Body
	return &Function{Parameters: params, Body: body, Env: env}
}
func (i *Interpreter) evalFunctionCalls(node *parser.CallExpression, env *Environment) Object {
	function := i.Eval(node.Function, env)
	if isError(function) {
		return function
	}
	args := i.evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
	return i.applyFunction(function, args)
}
func (i *Interpreter) evalExpressions(exps []parser.Expression, env *Environment) []Object {
	var result []Object
	for _, e := range exps {
		evaluated := i.Eval(e, env)
		if isError(evaluated) {
			return []Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

// applyFunction invokes user-defined or builtin functions with the provided arguments
func (i *Interpreter) applyFunction(fn Object, args []Object) Object {
	switch function := fn.(type) {
	case *Function:
		// create new function environment
		extendedEnv := i.extendFunctionEnv(function, args)
		evaluated := i.Eval(function.Body, extendedEnv)
		return i.unwrapReturnValue(evaluated)
	case *Builtin:
		return function.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}
func (i *Interpreter) extendFunctionEnv(function *Function, args []Object) *Environment {
	env := NewEnclosedEnvironment(function.Env)
	for paramIdx, param := range function.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}
func (i *Interpreter) unwrapReturnValue(obj Object) Object {
	if returnValue, ok := obj.(*ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}
func (i *Interpreter) evalPrefixExpression(node *parser.PrefixExpression, env *Environment) Object {
	right := i.Eval(node.Right, env)
	if isError(right) {
		return right
	}
	switch node.Operator {
	case "!":
		return i.evalBangOperatorExpression(right)
	case "-":
		if right.Type() == FLOAT_OBJ {
			return i.evalMinusPrefixOperatorExpressionFloat(right)
		}
		if right.Type() == INTEGER_OBJ {
			return i.evalMinusPrefixOperatorExpression(right)
		}
		return i.fileError(node.Token.Line, node.Token.Column, "unknown operator: %s%s", node.Operator, right.Type())
	default:
		return i.fileError(node.Token.Line, node.Token.Column, "unknown operator: %s%s", node.Operator, right.Type())
	}
}
func (i *Interpreter) evalBangOperatorExpression(right Object) Object {
	return nativeBoolToBooleanObject(!isTruthy(right))
}
func (i *Interpreter) evalMinusPrefixOperatorExpression(right Object) Object {
	rightVal := right.(*Integer).Value
	return &Integer{Value: -rightVal}
}
func (i *Interpreter) evalMinusPrefixOperatorExpressionFloat(right Object) Object {
	rightVal := right.(*Float).Value
	return &Float{Value: -rightVal}
}
func (i *Interpreter) evalArrayLiteral(node *parser.ArrayLiteral, env *Environment) Object {
	elements := i.evalExpressions(node.Elements, env)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}
	return &Array{Elements: elements}
}
func (i *Interpreter) evalArrayIndexExpression(node *parser.ArrayIndexExpression, env *Environment) Object {
	object := i.Eval(node.Array, env)
	if isError(object) {
		return object
	}
	index := i.Eval(node.Index, env)
	if isError(index) {
		return index
	}
	switch obj := object.(type) {
	case *Array:
		if object.Type() != ARRAY_OBJ || index.Type() != INTEGER_OBJ {
			return i.fileError(node.Token.Line, node.Token.Column, "index operator not supported: %s", object.Type())
		}

		idx := index.(*Integer).Value
		if idx < 0 || idx >= int64(len(object.(*Array).Elements)) {
			return i.fileError(node.Token.Line, node.Token.Column,
				"index out of bounds: index %d, length %d", idx, len(object.(*Array).Elements))
		}

		return object.(*Array).Elements[idx]
	case *Table:
		key := fmt.Sprintf("%s:%s", index.Type(), index.String())
		val, ok := obj.Pairs[key]
		if !ok {
			return NULL
		}
		return val
	default:
		return i.fileError(node.Token.Line, node.Token.Column, "index operator not supported: %s", object.Type())
	}
}

func (i *Interpreter) evalForStatement(node *parser.ForStatement, env *Environment) Object {
	for {
		if node.Condition != nil {
			con := i.Eval(node.Condition, env)
			if isError(con) {
				return con
			}
			if !isTruthy(con) {
				break
			}
		}

		result := i.Eval(node.Body, env)
		if isError(result) {
			return result
		}
		switch result.(type) {
		case *ContinueSignal:
			continue
		case *BreakSignal:
			return NULL
		case *ReturnValue:
			return result
		case *Error:
			return result
		}
	}
	return NULL
}

func (i *Interpreter) evalSwitchStatement(node *parser.SwitchStatement, env *Environment) Object {
	cond := i.Eval(node.Expression, env)
	if isError(cond) {
		return cond
	}
	if node.Cases != nil {
		for _, cs := range node.Cases {
			result := i.evalSwitchCase(&cs, env, cond)
			if result != nil {
				return result
			}
		}
	}

	if node.DefaultCase != nil {
		result := i.Eval(node.DefaultCase, env)
		if isError(result) {
			return result
		}
		return result
	}
	return NULL
}
func (i *Interpreter) evalSwitchCase(node *parser.SwitchCase, env *Environment, cond Object) Object {
	caseVal := i.Eval(node.Condition, env)
	if isError(caseVal) {
		return caseVal
	}
	if cond.String() == caseVal.String() && cond.Type() == caseVal.Type() {
		result := i.Eval(node.Body, env)
		if isError(result) {
			return result
		}
		switch result.(type) {
		case *ContinueSignal:
			return result
		case *BreakSignal:
			return NULL
		case *ReturnValue:
			return result
		case *Error:
			return result
		}
		return result
	}
	return nil
}

// evalForInStatement iterates over arrays, tables, or strings and executes the loop body
func (i *Interpreter) evalForInStatement(node *parser.ForInStatement, env *Environment) Object {
	iterable := i.Eval(node.Collection, env)
	if isError(iterable) {
		return iterable
	}

	switch obj := iterable.(type) {
	case *Array:
		for idx, e := range obj.Elements {
			if node.Index != nil {
				env.Set(node.Index.Value, &Integer{Value: int64(idx)})
			}
			env.Set(node.Item.Value, e)
			result := i.Eval(node.Body, env)
			if isError(result) {
				return result
			}
			switch result.(type) {
			case *ContinueSignal:
				continue
			case *BreakSignal:
				return NULL
			case *ReturnValue:
				return result
			}
		}
	case *Table:
		for k, v := range obj.Pairs {
			parts := strings.SplitN(k, ":", 2)
			displayKey := parts[1]
			env.Set(node.Item.Value, &Array{Elements: []Object{&String{Value: displayKey}, v}})
			result := i.Eval(node.Body, env)
			if isError(result) {
				return result
			}
			switch result.(type) {
			case *ContinueSignal:
				continue
			case *BreakSignal:
				return NULL
			case *ReturnValue:
				return result
			}
		}
	case *String:
		for _, ch := range obj.Value {
			env.Set(node.Item.Value, &String{Value: string(ch)})
			result := i.Eval(node.Body, env)
			if isError(result) {
				return result
			}
			switch result.(type) {
			case *ContinueSignal:
				continue
			case *BreakSignal:
				return NULL
			case *ReturnValue:
				return result
			}
		}
	default:
		return i.fileError(node.Token.Line, node.Token.Column, "cannot iterate over non-iterable type: %s", iterable.Type())
	}

	return NULL
}

func (i *Interpreter) evalTableLiteral(node *parser.TableLiteral, env *Environment) Object {
	table := &Table{Pairs: make(map[string]Object)}
	for _, p := range node.Pairs {
		var key string

		if ident, ok := p.Key.(*parser.Identifier); ok {
			key = fmt.Sprintf("%s:%s", STRING_OBJ, ident.Value)
		} else {
			evaluatedKey := i.Eval(p.Key, env)
			if isError(evaluatedKey) {
				return evaluatedKey
			}
			key = fmt.Sprintf("%s:%s", evaluatedKey.Type(), evaluatedKey.String())
		}

		evaluatedValue := i.Eval(p.Value, env)
		if isError(evaluatedValue) {
			return evaluatedValue
		}
		table.Pairs[key] = evaluatedValue
	}
	return table
}

// evalUseStatement loads and executes a module file into the current environment
func (i *Interpreter) evalUseStatement(node *parser.UseStatement, env *Environment) Object {
	fileName := node.FileName.Value + ".lgs"

	if cached, ok := i.ModuleCache[fileName]; ok {
		for k, v := range cached.store {
			env.Set(k, v)
		}
		return NULL
	}

	data, err := os.ReadFile(fileName)
	if err != nil && i.StdFs != nil {
		// fall back to embedded FS - try std/ subdirectory first, then root
		data, err = fs.ReadFile(i.StdFs, "std/"+fileName)
		if err != nil {
			data, err = fs.ReadFile(i.StdFs, fileName)
		}
	}
	if err != nil {
		return i.fileError(node.FileName.Token.Line, node.FileName.Token.Column,
			"module not found: %s", fileName)
	}

	lexer := golexer.NewLexer(string(data))
	p := parser.NewParser(lexer, fileName)
	program := p.Parse()
	if len(p.Errors()) != 0 {
		for _, e := range p.Errors() {
			fmt.Println(e)
		}
		return i.fileError(node.FileName.Token.Line, node.FileName.Token.Column,
			"failed to parse module: %s", fileName)
	}
	modulEnv := NewEnvironment()
	i.Eval(program, modulEnv)
	i.ModuleCache[fileName] = modulEnv

	for k, v := range modulEnv.store {
		env.Set(k, v)
	}
	return NULL
}

// evalDotExpression returns a table member by identifier when left is a table
func (i *Interpreter) evalDotExpression(node *parser.DotExpression, env *Environment) Object {
	left := i.Eval(node.Left, env)
	if isError(left) {
		return left
	}
	switch obj := left.(type) {
	case *Table:
		key := fmt.Sprintf("%s:%s", STRING_OBJ, node.Right.Value)
		val, ok := obj.Pairs[key]
		if !ok {
			return NULL
		}
		return val
	default:
		return i.fileError(node.Token.Line, node.Token.Column,
			"dot operator not supported on type: %s", left.Type())
	}
}
func (i *Interpreter) evalSpawnStatment(node *parser.SpawnStatment, env *Environment) Object {
	var wg sync.WaitGroup
	for _, stmt := range node.Block.Statements {
		wg.Add(1)
		go func(s parser.Statement) {
			defer wg.Done()
			localEnv := NewEnclosedEnvironment(env)
			result := i.Eval(s, localEnv)
			if isError(result) {
				fmt.Printf("error in spawned goroutine: %s\n", result.String())
			}
		}(stmt)
	}
	wg.Wait()
	return NULL
}
func (i *Interpreter) evalSpawnForInStatement(node *parser.SpawnForInStatement, env *Environment) Object {
	iterable := i.Eval(node.Collection, env)
	if isError(iterable) {
		return iterable
	}
	arr, ok := iterable.(*Array)
	if !ok {
		return i.fileError(node.Token.Line, node.Token.Column, "spawn for-in only supports arrays, got: %s", iterable.Type())
	}
	var wg sync.WaitGroup
	for _, e := range arr.Elements {
		wg.Add(1)
		go func(elem Object) {
			defer wg.Done()
			localEnv := NewEnclosedEnvironment(env)
			localEnv.Set(node.Item.Value, elem)
			result := i.Eval(node.Body, localEnv)
			if isError(result) {
				fmt.Printf("error in spawned goroutine: %s\n", result.String())
			}
		}(e)
	}
	wg.Wait()
	return NULL
}
func (i *Interpreter) evalTernary(node *parser.TenaryExpression, env *Environment) Object {
	condition := i.Eval(node.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return i.Eval(node.TrueBranch, env)
	} else {
		return i.Eval(node.FalseBranch, env)
	}
}

func (i *Interpreter) evalInterpol(node *parser.InterpolatedString, env *Environment) Object {
	var result strings.Builder
	for _, part := range node.Parts {
		val := i.Eval(part, env)
		if isError(val) {
			return val
		}
		result.WriteString(val.String())
	}
	return &String{Value: result.String()}
}
func (i *Interpreter) evalTry(node *parser.TryExpression, env *Environment) Object {
	res := i.Eval(node.Right, env)
	if isError(res) {
		return res
	}
	table, ok := res.(*Table)
	if !ok {
		return res
	}
	okVal, exists := table.Pairs["STRING:ok"]
	if !exists {
		return res
	}
	if okVal == FALSE {
		return &ReturnValue{Value: table}
	}
	val, exists := table.Pairs["STRING:value"]
	if !exists {
		return NULL
	}
	return val
}
