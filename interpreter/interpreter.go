package interpreter

import (
	"fmt"
	"strings"

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
type Interpreter struct {
	Env *Environment
}
type Integar struct {
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
type Error struct {
	Message string
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

func (i *Integar) Type() ObjectType {
	return INTEGER_OBJ

}
func (i *Integar) String() string {
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
func (e *Error) String() string {
	var out strings.Builder
	out.WriteString("ERROR: ")
	out.WriteString(e.Message)
	return out.String()
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
func NewInterpreter() *Interpreter {
	return &Interpreter{Env: NewEnvironment()}
}
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
func isError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}
func (i *Interpreter) Eval(node parser.Node, env *Environment) Object {
	switch node := node.(type) {
	case *parser.Program:
		return i.evalProgram(node, env)
	case *parser.ExpressionStatement:
		return i.Eval(node.Expression, env)
	case *parser.IntegerLiteral:
		return &Integar{Value: node.Value}
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
	default:
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
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: %s", node.Value)
}
func (i *Interpreter) evalLetStatement(node *parser.LetStatement, env *Environment) Object {
	val := i.Eval(node.Value, env)
	if isError(val) {
		return val
	}
	env.Set(node.Name.Value, val)
	return val
}
func (i *Interpreter) evalInfixExpression(node *parser.InfixExpression, env *Environment) Object {
	left := i.Eval(node.Left, env)
	if isError(left) {
		return left
	}
	right := i.Eval(node.Right, env)
	if isError(right) {
		return right
	}
	switch {
	case left.Type() == INTEGER_OBJ && right.Type() == INTEGER_OBJ:
		return i.evalIntegerInfixExpression(node.Operator, left, right)
	case left.Type() == FLOAT_OBJ && right.Type() == FLOAT_OBJ:
		return i.evalFloatInfixExpression(node.Operator, left, right)
	case left.Type() == STRING_OBJ && right.Type() == STRING_OBJ:
		return i.evalStringInfixExpression(node.Operator, left, right)
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
		return newError("type mismatch: %s %s %s", left.Type(), node.Operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
	}
}
func (i *Interpreter) evalIntegerInfixExpression(operator string, left, right Object) Object {
	rightVal := right.(*Integar).Value
	leftVal := left.(*Integar).Value
	switch operator {
	case "+":
		return &Integar{Value: leftVal + rightVal}
	case "-":
		return &Integar{Value: leftVal - rightVal}
	case "*":
		return &Integar{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &Integar{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newError("modulo by zero")
		}
		return &Integar{Value: leftVal % rightVal}
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}
func (i *Interpreter) evalFloatInfixExpression(operator string, left, right Object) Object {
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
			return newError("division by zero")
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
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
func (i *Interpreter) evalStringInfixExpression(operator string, left, right Object) Object {
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}
func (i *Interpreter) evalBlockStatment(block *parser.BlockStatement, env *Environment) Object {
	var result Object
	for _, statement := range block.Statements {
		result = i.Eval(statement, env)
		if result != nil {
			rt := result.Type()
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ {
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
func (i *Interpreter) applyFunction(fn Object, args []Object) Object {
	switch function := fn.(type) {
	case *Function:
		// create new function enviroment
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
		return newError("unknown operator: %s%s", node.Operator, right.Type())

	default:
		return newError("unknown operator: %s%s", node.Operator, right.Type())
	}
}
func (i *Interpreter) evalBangOperatorExpression(right Object) Object {
	return nativeBoolToBooleanObject(!isTruthy(right))
}
func (i *Interpreter) evalMinusPrefixOperatorExpression(right Object) Object {
	rightVal := right.(*Integar).Value
	return &Integar{Value: -rightVal}
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
