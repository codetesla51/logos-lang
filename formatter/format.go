package formatter

import (
	"strings"

	"github.com/codetesla51/logos/parser"
)

type Formatter struct {
	indent int
}

func New() *Formatter {
	return &Formatter{indent: 0}
}

func (f *Formatter) tab() string {
	return strings.Repeat("    ", f.indent)
}

func (f *Formatter) Format(program *parser.Program) string {
	var out strings.Builder
	for _, stmt := range program.Statements {
		out.WriteString(f.formatStatement(stmt))
		out.WriteString("\n")
	}
	return out.String()
}

func (f *Formatter) formatStatement(stmt parser.Statement) string {
	switch s := stmt.(type) {
	case *parser.LetStatement:
		return f.tab() + "let " + s.Name.String() + " = " + f.formatExpression(s.Value)
	case *parser.ReturnStatement:
		return f.tab() + "return " + f.formatExpression(s.ReturnValue)
	case *parser.ExpressionStatement:
		return f.tab() + f.formatExpression(s.Expression)
	case *parser.ForStatement:
		return f.formatFor(s)
	case *parser.ForInStatement:
		return f.formatForIn(s)
	case *parser.BreakStatement:
		return f.tab() + "break"
	case *parser.ContinueStatement:
		return f.tab() + "continue"
	case *parser.SwitchStatement:
		return f.formatSwitch(s)
	case *parser.UseStatement:
		return f.tab() + "use " + s.FileName.String()
	case *parser.SpawnStatment:
		return f.formatSpawn(s)
	case *parser.SpawnForInStatement:
		return f.formatSpawnForIn(s)
	default:
		return ""
	}
}

func (f *Formatter) formatExpression(expr parser.Expression) string {
	if expr == nil {
		return ""
	}
	switch e := expr.(type) {
	case *parser.Identifier:
		return e.Value
	case *parser.IntegerLiteral:
		return e.Token.Literal
	case *parser.FloatLiteral:
		return e.Token.Literal
	case *parser.BooleanLiteral:
		return e.Token.Literal
	case *parser.NullExpression:
		return "null"
	case *parser.StringLiteral:
		if e.IsRaw {
			return "`" + e.Value + "`"
		}
		return "\"" + e.Value + "\""
	case *parser.PrefixExpression:
		return e.Operator + f.formatExpression(e.Right)
	case *parser.InfixExpression:
		return f.formatExpression(e.Left) + " " + e.Operator + " " + f.formatExpression(e.Right)
	case *parser.CallExpression:
		return f.formatCall(e)
	case *parser.FunctionLiteral:
		return f.formatFunction(e)
	case *parser.IfExpression:
		return f.formatIf(e)
	case *parser.ArrayLiteral:
		return f.formatArray(e)
	case *parser.ArrayIndexExpression:
		return f.formatExpression(e.Array) + "[" + f.formatExpression(e.Index) + "]"
	case *parser.TableLiteral:
		return f.formatTable(e)
	case *parser.DotExpression:
		return f.formatExpression(e.Left) + "." + e.Right.String()

	default:
		return ""
	}
}

func (f *Formatter) formatBlock(block *parser.BlockStatement) string {
	var out strings.Builder
	out.WriteString("{\n")
	f.indent++
	for _, stmt := range block.Statements {
		out.WriteString(f.formatStatement(stmt))
		out.WriteString("\n")
	}
	f.indent--
	out.WriteString(f.tab() + "}")
	return out.String()
}

func (f *Formatter) formatCall(e *parser.CallExpression) string {
	var out strings.Builder
	out.WriteString(f.formatExpression(e.Function))
	out.WriteString("(")
	for i, arg := range e.Arguments {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(f.formatExpression(arg))
	}
	out.WriteString(")")
	return out.String()
}

func (f *Formatter) formatFunction(e *parser.FunctionLiteral) string {
	var out strings.Builder
	out.WriteString("fn(")
	for i, p := range e.Parameters {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(p.String())
	}
	out.WriteString(") ")
	if e.IsArrow {
		returnStmt := e.Body.Statements[0].(*parser.ReturnStatement)
		out.WriteString("-> " + f.formatExpression(returnStmt.ReturnValue))
		return out.String()
	}
	out.WriteString(f.formatBlock(e.Body))
	return out.String()
}

func (f *Formatter) formatIf(e *parser.IfExpression) string {
	var out strings.Builder
	out.WriteString("if " + f.formatExpression(e.Condition) + " ")
	out.WriteString(f.formatBlock(e.Consequence))
	if e.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(f.formatBlock(e.Alternative))
	}
	return out.String()
}

func (f *Formatter) formatFor(s *parser.ForStatement) string {
	var out strings.Builder
	out.WriteString(f.tab() + "for")
	if s.Condition != nil {
		out.WriteString(" " + f.formatExpression(s.Condition))
	}
	out.WriteString(" ")
	out.WriteString(f.formatBlock(s.Body))
	return out.String()
}

func (f *Formatter) formatForIn(s *parser.ForInStatement) string {
	var out strings.Builder
	out.WriteString(f.tab() + "for " + s.Item.String() + " in " + f.formatExpression(s.Collection) + " ")
	out.WriteString(f.formatBlock(s.Body))
	return out.String()
}

func (f *Formatter) formatSwitch(s *parser.SwitchStatement) string {
	var out strings.Builder
	out.WriteString(f.tab() + "switch " + f.formatExpression(s.Expression) + " {\n")
	f.indent++
	for _, c := range s.Cases {
		out.WriteString(f.tab() + "case " + f.formatExpression(c.Condition) + " ")
		out.WriteString(f.formatBlock(c.Body))
		out.WriteString("\n")
	}
	if s.DefaultCase != nil {
		out.WriteString(f.tab() + "default ")
		out.WriteString(f.formatBlock(s.DefaultCase))
		out.WriteString("\n")
	}
	f.indent--
	out.WriteString(f.tab() + "}")
	return out.String()
}

func (f *Formatter) formatArray(e *parser.ArrayLiteral) string {
	if len(e.Elements) == 0 {
		return "[]"
	}
	var out strings.Builder
	out.WriteString("[")
	for i, el := range e.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(f.formatExpression(el))
	}
	out.WriteString("]")
	return out.String()
}

func (f *Formatter) formatTable(e *parser.TableLiteral) string {
	if len(e.Pairs) == 0 {
		return "table{}"
	}
	var out strings.Builder
	out.WriteString("table{\n")
	f.indent++
	for _, p := range e.Pairs {
		out.WriteString(f.tab() + f.formatExpression(p.Key) + ": " + f.formatExpression(p.Value) + ",\n")
	}
	f.indent--
	out.WriteString(f.tab() + "}")
	return out.String()
}
func (f *Formatter) formatSpawn(e *parser.SpawnStatment) string {
	var out strings.Builder
	out.WriteString("spawn ")
	out.WriteString(f.formatBlock(e.Block))
	return out.String()
}
func (f *Formatter) formatSpawnForIn(s *parser.SpawnForInStatement) string {
	var out strings.Builder
	out.WriteString(f.tab() + "spawn " + "for " + s.Item.String() + " in " + f.formatExpression(s.Collection) + " ")
	out.WriteString(f.formatBlock(s.Body))
	return out.String()
}
