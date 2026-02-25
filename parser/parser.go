package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codetesla51/golexer/golexer"
)

// Operator precedence levels for expression parsing.
const (
	_           int = iota
	LOWEST          // 1 - Empty statement, assignment
	LOGICAL_OR      // 2 - ||
	LOGICAL_AND     // 3 - &&
	EQUALS          // 4 - ==, !=
	LESSGREATER     // 5 - <, >, <=, >=
	SUM             // 6 - +, -
	PRODUCT         // 7 - *, /
	PREFIX          // 9 - unary -, !
	CALL            // 11 - function calls
)

// extra tokens
const (
	SWITCH  = golexer.TokenType("SWITCH")
	CASE    = golexer.TokenType("CASE")
	DEFAULT = golexer.TokenType("DEFAULT")
	ARROW   = golexer.TokenType("ARROW")
	IN      = golexer.TokenType("IN")
	TABLE   = golexer.TokenType("TABLE")
)

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement in the AST.
type Statement interface {
	Node
	statmentNode()
}

// Expression represents an expression in the AST.
type Expression interface {
	Node
	expressionNode()
}
type Program struct {
	Statements []Statement
}
type Identifier struct {
	Token golexer.Token
	Value string
}
type IntegerLiteral struct {
	Token golexer.Token
	Value int64
}
type FloatLiteral struct {
	Token golexer.Token
	Value float64
}
type InfixExpression struct {
	Token    golexer.Token
	Left     Expression
	Operator string
	Right    Expression
}
type PrefixExpression struct {
	Token    golexer.Token
	Operator string
	Right    Expression
}
type BooleanLiteral struct {
	Token golexer.Token
	Value bool
}
type LetStatement struct {
	Token golexer.Token
	Name  *Identifier
	Value Expression
}
type ReturnStatement struct {
	Token       golexer.Token
	ReturnValue Expression
}
type ExpressionStatement struct {
	Token      golexer.Token
	Expression Expression
}
type IfExpression struct {
	Token       golexer.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

type ForStatement struct {
	Token     golexer.Token
	Condition Expression // Optional: condition expression
	Body      *BlockStatement
}
type BlockStatement struct {
	Token      golexer.Token
	Statements []Statement
}
type FunctionLiteral struct {
	Token      golexer.Token
	Parameters []*Identifier
	Body       *BlockStatement
}
type CallExpression struct {
	Token     golexer.Token
	Function  Expression
	Arguments []Expression
}
type StringLiteral struct {
	Token golexer.Token
	Value string
	IsRaw bool
}
type ArrayLiteral struct {
	Token    golexer.Token
	Elements []Expression
}
type ArrayIndexExpression struct {
	Token golexer.Token
	Array Expression
	Index Expression
}
type BreakStatement struct {
	Token golexer.Token
}
type ContinueStatement struct {
	Token golexer.Token
}
type SwitchStatement struct {
	Token       golexer.Token
	Expression  Expression
	Cases       []SwitchCase
	DefaultCase *BlockStatement
}
type SwitchCase struct {
	Token     golexer.Token
	Condition Expression
	Body      *BlockStatement
}
type NullExpression struct {
	Token golexer.Token
}
type ForInStatement struct {
	Token      golexer.Token
	item       *Identifier
	collection Expression
	Body       *BlockStatement
}
type TableLiteral struct {
	Token golexer.Token
	Pairs []TablePair
}
type TablePair struct {
	Key   Expression
	Value Expression
}
type PrefixParsefn func() Expression
type InfixParsefn func(Expression) Expression

var precedences = map[golexer.TokenType]int{
	golexer.ASSIGN:           LOWEST,
	golexer.PLUS_ASSIGN:      LOWEST,
	golexer.MINUS_ASSIGN:     LOWEST,
	golexer.MULTIPLY_ASSIGN:  LOWEST,
	golexer.DIVIDE_ASSIGN:    LOWEST,
	golexer.MODULUS_ASSIGN:   LOWEST,
	golexer.EQL:              EQUALS,
	golexer.NOT_EQL:          EQUALS,
	golexer.LESS_THAN:        LESSGREATER,
	golexer.GREATER_THAN:     LESSGREATER,
	golexer.PLUS:             SUM,
	golexer.MINUS:            SUM,
	golexer.MULTIPLY:         PRODUCT,
	golexer.DIVIDE:           PRODUCT,
	golexer.AND:              LOGICAL_AND,
	golexer.OR:               LOGICAL_OR,
	golexer.GREATER_THAN_EQL: LESSGREATER,
	golexer.LESS_THAN_EQL:    LESSGREATER,
	golexer.LPAREN:           CALL,
	golexer.LBRACKET:         CALL,
	golexer.MODULUS:          PRODUCT,
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}
func (p *Program) String() string {
	var out strings.Builder
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}
func (i *Identifier) expressionNode()      {}
func (i *Identifier) statmentNode()        {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", ie.Left.String(), ie.Operator, ie.Right.String())
}
func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String())
}

func (ls *LetStatement) statmentNode()        {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out strings.Builder
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

func (rs *ReturnStatement) statmentNode()        {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out strings.Builder
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}
func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out strings.Builder
	out.WriteString("if")
	out.WriteString("(")
	out.WriteString(ie.Condition.String())
	out.WriteString(")")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

func (fs *ForStatement) statmentNode()        {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out strings.Builder
	out.WriteString(fs.TokenLiteral())
	out.WriteString("(")
	out.WriteString(fs.Condition.String())
	out.WriteString(")")

	out.WriteString(fs.Body.String())
	return out.String()
}

func (bs *BlockStatement) statmentNode()        {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out strings.Builder
	out.WriteString("{")
	for i, s := range bs.Statements {
		str := s.String()
		out.WriteString(str)
		if i < len(bs.Statements)-1 && !strings.HasSuffix(str, ";") {
			out.WriteString(";")
		}
	}
	out.WriteString("}")
	return out.String()
}
func (es *ExpressionStatement) statmentNode()        {}
func (es *ExpressionStatement) expressionNode()      {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}
func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) statmentNode()        {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out strings.Builder
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	for i, p := range fl.Parameters {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(p.String())
	}
	out.WriteString(")")
	out.WriteString(fl.Body.String())
	return out.String()
}
func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out strings.Builder
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	for i, arg := range ce.Arguments {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(arg.String())
	}
	out.WriteString(")")
	return out.String()
}
func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string {
	var out strings.Builder
	if sl.IsRaw {
		out.WriteString("`")
		out.WriteString(sl.Value)
		out.WriteString("`")
		return out.String()
	}
	out.WriteString("\"")
	out.WriteString(sl.Value)
	out.WriteString("\"")
	return out.String()
}
func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out strings.Builder
	out.WriteString("[")
	for i, elem := range al.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(elem.String())
	}
	out.WriteString("]")
	return out.String()
}
func (ae *ArrayIndexExpression) expressionNode()      {}
func (ae *ArrayIndexExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *ArrayIndexExpression) String() string {
	var out strings.Builder
	out.WriteString("(")
	out.WriteString(ae.Array.String())
	out.WriteString("[")
	out.WriteString(ae.Index.String())
	out.WriteString("])")
	return out.String()
}
func (bs *BreakStatement) statmentNode()        {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string {
	return bs.TokenLiteral() + ";"
}
func (cs *ContinueStatement) statmentNode()        {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string {
	return cs.TokenLiteral() + ";"
}
func (ss *SwitchStatement) statmentNode()        {}
func (ss *SwitchStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SwitchStatement) String() string {
	var out strings.Builder
	out.WriteString("switch")
	out.WriteString("(")
	out.WriteString(ss.Expression.String())
	out.WriteString(")")
	out.WriteString("{")
	for _, c := range ss.Cases {
		out.WriteString("case ")
		out.WriteString(c.Condition.String())
		out.WriteString(" ")
		out.WriteString(c.Body.String())
	}
	if ss.DefaultCase != nil {
		out.WriteString("default ")
		out.WriteString(ss.DefaultCase.String())
	}
	out.WriteString("}")
	return out.String()
}
func (sc *SwitchCase) TokenLiteral() string { return sc.Token.Literal }
func (sc *SwitchCase) String() string {
	var out strings.Builder
	out.WriteString(sc.TokenLiteral())
	out.WriteString(" ")
	out.WriteString(sc.Condition.String())
	out.WriteString(" {")
	out.WriteString(sc.Body.String())
	out.WriteString("}")
	return out.String()
}
func (nl *NullExpression) expressionNode()      {}
func (nl *NullExpression) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullExpression) String() string {
	return nl.TokenLiteral()
}
func (fi *ForInStatement) statmentNode()        {}
func (fi *ForInStatement) TokenLiteral() string { return fi.Token.Literal }
func (fi *ForInStatement) String() string {
	var out strings.Builder
	out.WriteString("for")
	out.WriteString("(")
	out.WriteString(fi.item.String())
	out.WriteString(" in ")
	out.WriteString(fi.collection.String())
	out.WriteString(")")
	out.WriteString(fi.Body.String())
	return out.String()
}
func (tl *TableLiteral) expressionNode()      {}
func (tl *TableLiteral) TokenLiteral() string { return tl.Token.Literal }
func (tl *TableLiteral) String() string {
	var out strings.Builder
	out.WriteString(tl.TokenLiteral())
	out.WriteString("{")
	if tl.Pairs != nil {
		for i, p := range tl.Pairs {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(p.Key.String())
			out.WriteString(":")
			out.WriteString(p.Value.String())

		}
	}
	out.WriteString("}")
	return out.String()

}

// Parser implements a Pratt parser for parsing tokens into an AST.
type Parser struct {
	lexer  *golexer.Lexer
	errors []string

	curToken       golexer.Token
	peekToken      golexer.Token
	prefixParseFns map[golexer.TokenType]PrefixParsefn
	infixParseFns  map[golexer.TokenType]InfixParsefn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}
func (p *Parser) curTokenIs(t golexer.TokenType) bool {
	return p.curToken.Type == t
}
func (p *Parser) peekTokenIs(t golexer.TokenType) bool {
	return p.peekToken.Type == t
}
func (p *Parser) expectPeek(t golexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}
func (p *Parser) peekError(t golexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead on line %d col %d", t, p.peekToken.Type, p.peekToken.Line, p.peekToken.Column)
	p.errors = append(p.errors, msg)
}
func (p *Parser) noPrefixParseFnError(t golexer.TokenType) {
	msg := fmt.Sprintf(
		"no prefix parse function for %s found on line %d col %d",
		t,
		p.curToken.Line,
		p.curToken.Column,
	)
	p.errors = append(p.errors, msg)
}
func (p *Parser) synchronize() {
	for !p.curTokenIs(golexer.EOF) {
		if p.curTokenIs(golexer.SEMICOLON) {
			p.nextToken()
			return
		}
		switch p.peekToken.Type {
		case golexer.LET, golexer.RETURN, golexer.IF, golexer.FN:
			return
		}
		p.nextToken()
	}

}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}
func (p *Parser) currentPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) Errors() []string {
	return p.errors
}

// NewParser creates and initializes a new Parser with the given lexer.
func NewParser(lexer *golexer.Lexer) *Parser {
	p := &Parser{
		lexer:  lexer,
		errors: []string{},
	}
	p.prefixParseFns = make(map[golexer.TokenType]PrefixParsefn)
	p.infixParseFns = make(map[golexer.TokenType]InfixParsefn)
	p.registerPrefix(golexer.IDENT, p.parseIdentifier)
	p.registerPrefix(golexer.NUMBER, p.parseIntegerLiteral)
	p.registerPrefix(golexer.LPAREN, p.parseGroupedExpression)

	// infix expressions
	p.registerInfix(golexer.PLUS, p.parseInfixExpression)
	p.registerInfix(golexer.MINUS, p.parseInfixExpression)
	p.registerInfix(golexer.MULTIPLY, p.parseInfixExpression)
	p.registerInfix(golexer.DIVIDE, p.parseInfixExpression)
	p.registerInfix(golexer.MODULUS, p.parseInfixExpression)
	p.registerInfix(golexer.EQL, p.parseInfixExpression)
	p.registerInfix(golexer.NOT_EQL, p.parseInfixExpression)
	p.registerInfix(golexer.LESS_THAN, p.parseInfixExpression)
	p.registerInfix(golexer.GREATER_THAN, p.parseInfixExpression)
	p.registerInfix(golexer.GREATER_THAN_EQL, p.parseInfixExpression)
	p.registerInfix(golexer.LESS_THAN_EQL, p.parseInfixExpression)
	p.registerInfix(golexer.AND, p.parseInfixExpression)
	p.registerInfix(golexer.OR, p.parseInfixExpression)
	p.registerInfix(golexer.LBRACKET, p.parseArrayIndexExpression)
	p.registerInfix(golexer.ASSIGN, p.parseInfixExpression)
	p.registerInfix(golexer.PLUS_ASSIGN, p.parseInfixExpression)
	p.registerInfix(golexer.MINUS_ASSIGN, p.parseInfixExpression)
	p.registerInfix(golexer.MULTIPLY_ASSIGN, p.parseInfixExpression)
	p.registerInfix(golexer.DIVIDE_ASSIGN, p.parseInfixExpression)
	p.registerInfix(golexer.MODULUS_ASSIGN, p.parseInfixExpression)
	// prefix expressions
	p.registerPrefix(golexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(golexer.BANG, p.parsePrefixExpression)
	p.registerPrefix(golexer.STRING, p.parserStringLiteral)
	p.registerPrefix(golexer.BACKTICK_STRING, p.parserStringLiteral)
	p.registerPrefix(golexer.NULL, p.parseNullExpression)
	// boolean literals
	p.registerPrefix(golexer.TRUE, p.parseBoolean)
	p.registerPrefix(golexer.FALSE, p.parseBoolean)

	// if expressions
	p.registerPrefix(golexer.IF, func() Expression { return p.parseIfExpression() })
	p.registerPrefix(golexer.FN, func() Expression { return p.parseFunctionLiteral() })
	// function calls
	p.registerInfix(golexer.LPAREN, p.parseFunctionCall)
	//arrays
	p.registerPrefix(golexer.LBRACKET, func() Expression {
		return p.parseArrayLiteral()
	})
	//tables (hash maps)
	p.registerPrefix(TABLE, p.parseTableLiteral)
	p.nextToken()
	p.nextToken()
	return p
}
func (p *Parser) registerPrefix(tokenType golexer.TokenType, fn PrefixParsefn) {
	p.prefixParseFns[tokenType] = fn
}
func (p *Parser) registerInfix(tokenType golexer.TokenType, fn InfixParsefn) {
	p.infixParseFns[tokenType] = fn
}

// Parse parses the input tokens into a Program AST node.
// Returns nil if parsing encounters any errors.
func (p *Parser) Parse() *Program {
	program := &Program{}
	program.Statements = []Statement{}
	for !p.curTokenIs(golexer.EOF) {
		stmt := p.parseStatment()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}
func (p *Parser) parseStatment() Statement {
	switch p.curToken.Type {
	case golexer.LET:
		return p.parseLetStatment()
	case golexer.RETURN:
		return p.parseReturnStatment()
	case golexer.FOR:
		return p.parseForStatement()
	case golexer.IDENT:
		return p.parseExpressionStatment()
	case golexer.NUMBER:
		return p.parseExpressionStatment()
	case golexer.IF:
		return &ExpressionStatement{
			Token:      p.curToken,
			Expression: p.parseIfExpression(),
		}
	case golexer.FN:
		return &ExpressionStatement{
			Token:      p.curToken,
			Expression: p.parseFunctionLiteral(),
		}
	case golexer.CONTINUE:
		return &ContinueStatement{
			Token: p.curToken,
		}
	case golexer.BREAK:
		return &BreakStatement{
			Token: p.curToken,
		}
	case SWITCH:
		return p.parseSwitchStatement()

	default:
		return p.parseExpressionStatment()
	}

}
func (p *Parser) parseLetStatment() *LetStatement {
	stmt := &LetStatement{Token: p.curToken}
	if !p.expectPeek(golexer.IDENT) {
		p.synchronize()
		return nil
	}
	stmt.Name = &Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	if !p.expectPeek(golexer.ASSIGN) {
		p.synchronize()
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(golexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
func (p *Parser) parseReturnStatment() *ReturnStatement {
	stmt := &ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	if p.peekTokenIs(golexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
func (p *Parser) parseExpressionStatment() *ExpressionStatement {
	stmt := &ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(golexer.ASSIGN) || p.peekTokenIs(golexer.PLUS_ASSIGN) || p.peekTokenIs(golexer.MINUS_ASSIGN) || p.peekTokenIs(golexer.MULTIPLY_ASSIGN) || p.peekTokenIs(golexer.DIVIDE_ASSIGN) || p.peekTokenIs(golexer.MODULUS_ASSIGN) {
		p.nextToken()
		operator := p.curToken.Literal
		p.nextToken()
		right := p.parseExpression(LOWEST)
		stmt.Expression = &InfixExpression{
			Token:    p.curToken,
			Left:     stmt.Expression,
			Operator: operator,
			Right:    right,
		}
	}
	if p.peekTokenIs(golexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
func (p *Parser) parseIfExpression() *IfExpression {
	stmt := &IfExpression{
		Token: p.curToken,
	}
	if !p.expectPeek(golexer.LPAREN) {
		p.synchronize()
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(golexer.RPAREN) {
		p.synchronize()
		return nil
	}
	if !p.expectPeek(golexer.LBRACE) {
		p.synchronize()
		return nil
	}
	stmt.Consequence = p.parseBlockStatement()
	if p.peekTokenIs(golexer.ELSE) {
		p.nextToken()
		if !p.expectPeek(golexer.LBRACE) {
			p.synchronize()
			return nil
		}
		stmt.Alternative = p.parseBlockStatement()
	}
	return stmt
}

func (p *Parser) parseForStatement() Statement {
	stmt := &ForStatement{Token: p.curToken}
	if !p.expectPeek(golexer.LPAREN) {
		p.synchronize()
		return nil
	}
	p.nextToken()
	if p.curTokenIs(golexer.IDENT) && p.peekTokenIs(IN) {
		return p.parseForInStatement()
	}
	if p.curTokenIs(golexer.RPAREN) {

		p.nextToken()
	} else {
		stmt.Condition = p.parseExpression(LOWEST)
		if !p.expectPeek(golexer.RPAREN) {
			p.synchronize()
			return nil
		}
		p.nextToken()
	}
	if !p.curTokenIs(golexer.LBRACE) {
		p.synchronize()
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}
func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{
		Token: p.curToken,
	}
	block.Statements = []Statement{}
	p.nextToken()
	for !p.curTokenIs(golexer.RBRACE) && !p.curTokenIs(golexer.EOF) {
		stmt := p.parseStatment()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.nextToken()

	}
	return block
}
func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		p.synchronize()
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(golexer.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}
func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}
func (p *Parser) parseIntegerLiteral() Expression {
	// Check if the literal contains a decimal point (float)
	if strings.Contains(p.curToken.Literal, ".") {
		lit := &FloatLiteral{Token: p.curToken}
		value, err := strconv.ParseFloat(p.curToken.Literal, 64)
		if err != nil {
			p.synchronize()
			msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
			p.errors = append(p.errors, msg)
			return nil
		}
		lit.Value = value
		return lit
	}

	lit := &IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.synchronize()
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}
func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.currentPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression

}
func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(golexer.RPAREN) {
		p.synchronize()
		return nil
	}
	return exp
}
func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}
func (p *Parser) parseBoolean() Expression {
	return &BooleanLiteral{
		Token: p.curToken,
		Value: p.curTokenIs(golexer.TRUE),
	}
}
func (p *Parser) parseFunctionLiteral() Expression {
	lit := &FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(golexer.LPAREN) {
		p.synchronize()
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()
	if p.peekTokenIs(ARROW) {
		p.nextToken()

		p.nextToken()

		body := p.parseExpression(LOWEST)
		lit.Body = &BlockStatement{
			Token: p.curToken,
			Statements: []Statement{&ReturnStatement{
				Token:       golexer.Token{Type: golexer.RETURN, Literal: "return"},
				ReturnValue: body,
			}},
		}
	} else {
		if !p.expectPeek(golexer.LBRACE) {
			p.synchronize()
			return nil
		}
		lit.Body = p.parseBlockStatement()
	}
	return lit
}

func (p *Parser) parseFunctionParameters() []*Identifier {
	identifiers := []*Identifier{}
	if p.peekTokenIs(golexer.RPAREN) {
		p.nextToken()
		return identifiers
	}
	p.nextToken()
	ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)
	for p.peekTokenIs(golexer.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}
	if !p.expectPeek(golexer.RPAREN) {
		p.synchronize()
		return nil
	}

	return identifiers
}
func (p *Parser) parseFunctionCall(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}
func (p *Parser) parseCallArguments() []Expression {
	args := []Expression{}
	if p.peekTokenIs(golexer.RPAREN) {
		p.nextToken()
		return args // no args
	}
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))
	for p.peekTokenIs(golexer.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(golexer.RPAREN) {
		p.synchronize()
		return nil
	}
	return args
}
func (p *Parser) parserStringLiteral() Expression {
	slit := &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	if p.curToken.Type == golexer.BACKTICK_STRING {
		slit.IsRaw = true
	}
	return slit
}
func (p *Parser) parseArrayLiteral() Expression {
	arr := &ArrayLiteral{Token: p.curToken}
	arr.Elements = p.parseArrayElements()
	return arr
}

func (p *Parser) parseArrayElements() []Expression {
	arrElements := []Expression{}
	if p.peekTokenIs(golexer.RBRACKET) {
		p.nextToken()
		return arrElements
	}
	p.nextToken()
	arrElements = append(arrElements, p.parseExpression(LOWEST))
	for p.peekTokenIs(golexer.COMMA) {
		p.nextToken()
		p.nextToken()
		arrElements = append(arrElements, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(golexer.RBRACKET) {
		p.synchronize()
		return nil
	}
	return arrElements
}
func (p *Parser) parseArrayIndexExpression(array Expression) Expression {
	exp := &ArrayIndexExpression{Token: p.curToken, Array: array}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(golexer.RBRACKET) {
		p.synchronize()
		return nil
	}
	return exp
}
func (p *Parser) parseSwitchStatement() Statement {
	stmt := &SwitchStatement{Token: p.curToken}
	if !p.expectPeek(golexer.LPAREN) {
		p.synchronize()
		return nil
	}
	p.nextToken()
	stmt.Expression = p.parseExpression(LOWEST)
	if !p.expectPeek(golexer.RPAREN) {
		p.synchronize()
		return nil
	}
	if !p.expectPeek(golexer.LBRACE) {
		p.synchronize()
		return nil
	}
	p.nextToken()
	stmt.Cases = []SwitchCase{}
	for p.curTokenIs(CASE) {
		p.nextToken()
		caseStmt := SwitchCase{Token: p.curToken}
		caseStmt.Condition = p.parseExpression(LOWEST)
		if !p.expectPeek(golexer.LBRACE) {
			p.synchronize()
			return nil
		}
		caseStmt.Body = p.parseBlockStatement()
		p.nextToken()
		stmt.Cases = append(stmt.Cases, caseStmt)
	}
	if p.curTokenIs(DEFAULT) {
		if !p.expectPeek(golexer.LBRACE) {
			p.synchronize()
			return nil
		}
		stmt.DefaultCase = p.parseBlockStatement()
		p.nextToken() // consume outer }
	}

	return stmt

}
func (p *Parser) parseNullExpression() Expression {
	return &NullExpression{Token: p.curToken}
}
func (p *Parser) parseForInStatement() Statement {
	stmt := &ForInStatement{Token: p.curToken}
	stmt.item = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()
	p.nextToken()
	stmt.collection = p.parseExpression(LOWEST)
	if !p.expectPeek(golexer.RPAREN) {
		p.synchronize()
		return nil
	}

	if !p.expectPeek(golexer.LBRACE) {
		p.synchronize()
		return nil
	}
	stmt.Body = p.parseBlockStatement()

	return stmt
}
func (p *Parser) parseTableLiteral() Expression {
	exp := &TableLiteral{Token: p.curToken}
	if !p.expectPeek(golexer.LBRACE) {
		p.synchronize()
		return nil
	}
	p.nextToken()
	for !p.curTokenIs(golexer.RBRACE) && golexer.TokenType("EOF") != p.curToken.Type {
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(golexer.COLON) {
			p.synchronize()
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		exp.Pairs = append(exp.Pairs, TablePair{Key: key, Value: value})
		if p.peekTokenIs(golexer.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}
	return exp
}
