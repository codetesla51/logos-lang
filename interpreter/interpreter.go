package interpreter

import (
	"fmt"
	"os"
	"strings"

	"github.com/codetesla51/golexer/golexer"
	"github.com/codetesla51/logos/parser"
)

type ObjectType string
type BuiltinFunc func(args ...Object) Object

const (
	INTEGER_OBJ      = "Integer"
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
type Interpreter struct {
	Env *Environment
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

func newErrorAt(line, col int, format string, a ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, a...),
		Line:    line,
		Column:  col,
	}
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
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newErrorAt(node.Token.Line, node.Token.Column, "identifier not found: %s", node.Value)
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
	if node.Operator == "=" {
		ident, ok := node.Left.(*parser.Identifier)
		if !ok {
			return newErrorAt(node.Token.Line, node.Token.Column, "cannot assign to non-identifier")
		}
		val := i.Eval(node.Right, env)
		if isError(val) {
			return val
		}
		result := env.Update(ident.Value, val)
		if result == nil {
			return newErrorAt(node.Token.Line, node.Token.Column, "cannot assign to undeclared variable: %s", ident.Value)
		}
		return val
	}
	if node.Operator == "+=" {
		ident, ok := node.Left.(*parser.Identifier)
		if !ok {
			return newErrorAt(node.Token.Line, node.Token.Column, "cannot assign to non-identifier")
		}
		val := i.Eval(node.Right, env)
		if isError(val) {
			return val
		}
		if !ok {
			return newErrorAt(node.Token.Line, node.Token.Column, "cannot assign to undeclared variable: %s", ident.String())
		}
		result := i.evalInfixExpression(&parser.InfixExpression{
			Token:    node.Token,
			Left:     node.Left,
			Operator: "+",
			Right:    node.Right,
		}, env)
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
	if node.Operator == "-=" {
		ident, ok := node.Left.(*parser.Identifier)
		if !ok {
			return newErrorAt(node.Token.Line, node.Token.Column, "cannot assign to non-identifier")
		}
		val := i.Eval(node.Right, env)
		if isError(val) {
			return val
		}
		if !ok {
			return newErrorAt(node.Token.Line, node.Token.Column, "cannot assign to undeclared variable: %s", ident.String())
		}
		result := i.evalInfixExpression(&parser.InfixExpression{
			Token:    node.Token,
			Left:     node.Left,
			Operator: "-",
			Right:    node.Right,
		}, env)
		if isError(result) {
			return result
		}
		env.Update(ident.Value, result)
		return result
	}
	switch {

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
		return newErrorAt(node.Token.Line, node.Token.Column,
			"type mismatch: %s %s %s", left.Type(), node.Operator, right.Type())
	default:
		return newErrorAt(node.Token.Line, node.Token.Column,
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
		return newErrorAt(line, col, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
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
		return newErrorAt(line, col, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
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
		return newErrorAt(line, col, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
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
		return newErrorAt(node.Token.Line, node.Token.Column, "unknown operator: %s%s", node.Operator, right.Type())
	default:
		return newErrorAt(node.Token.Line, node.Token.Column, "unknown operator: %s%s", node.Operator, right.Type())
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
			return newErrorAt(node.Token.Line, node.Token.Column, "index operator not supported: %s", object.Type())
		}

		idx := index.(*Integer).Value
		if idx < 0 || idx >= int64(len(object.(*Array).Elements)) {
			return newErrorAt(node.Token.Line, node.Token.Column,
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
		return newErrorAt(node.Token.Line, node.Token.Column, "index operator not supported: %s", object.Type())
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
	if cond.String() == caseVal.String() {
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
func (i *Interpreter) evalForInStatement(node *parser.ForInStatement, env *Environment) Object {
	iterable := i.Eval(node.Collection, env)
	if isError(iterable) {
		return iterable
	}

	switch obj := iterable.(type) {
	case *Array:
		for _, e := range obj.Elements {
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
		return newErrorAt(node.Token.Line, node.Token.Column, "cannot iterate over non-iterable type: %s", iterable.Type())
	}

	return NULL
}

func (i *Interpreter) evalTableLiteral(node *parser.TableLiteral, env *Environment) Object {
	table := &Table{Pairs: make(map[string]Object)}
	for _, p := range node.Pairs {
		evaluatedKey := i.Eval(p.Key, env)
		if isError(evaluatedKey) {
			return evaluatedKey
		}
		evaluatedValue := i.Eval(p.Value, env)
		if isError(evaluatedValue) {
			return evaluatedValue
		}
		key := fmt.Sprintf("%s:%s", evaluatedKey.Type(), evaluatedKey.String())
		table.Pairs[key] = evaluatedValue
	}

	return table
}
func (i *Interpreter) evalUseStatement(node *parser.UseStatement, env *Environment) Object {
	fileName := node.FileName.Value + ".lgs"
	data, err := os.ReadFile(fileName)
	if err != nil {
		return newErrorAt(node.FileName.Token.Line, node.FileName.Token.Column,
			"module not found: %s", fileName)
	}
	lexer := golexer.NewLexerWithConfig(string(data), "../tokens.json")
	p := parser.NewParser(lexer)
	program := p.Parse()
	if len(p.Errors()) != 0 {
		for _, e := range p.Errors() {
			fmt.Println(e)
		}
		return newErrorAt(node.FileName.Token.Line, node.FileName.Token.Column,
			"failed to parse module: %s", fileName)
	}
	return i.Eval(program, env)
}
