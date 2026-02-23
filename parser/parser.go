package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codetesla51/golexer/golexer"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

type Node interface {
	TokenLiteral() string
	String() string
}
type Statement interface {
	Node
	statmentNode()
}
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
type PrefixParsefn func() Expression
type InfixParsefn func(Expression) Expression

var precedences = map[golexer.TokenType]int{
	golexer.EQL:          EQUALS,
	golexer.NOT_EQL:      EQUALS,
	golexer.LESS_THAN:    LESSGREATER,
	golexer.GREATER_THAN: LESSGREATER,
	golexer.PLUS:         SUM,
	golexer.MINUS:        SUM,
	golexer.MULTIPLY:     PRODUCT,
	golexer.DIVIDE:       PRODUCT,
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}
func (p *Program) String() string {
	var out strings.Builder
	fmt.Println("all statements")
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
func (rs *ReturnStatement) expressionNode()      {}
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
func (es *ExpressionStatement) statmentNode()        {}
func (es *ExpressionStatement) expressionNode()      {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

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
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}
func (p *Parser) noPrefixParseFnError(t golexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
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
	p.registerInfix(golexer.EQL, p.parseInfixExpression)
	p.registerInfix(golexer.NOT_EQL, p.parseInfixExpression)
	p.registerInfix(golexer.LESS_THAN, p.parseInfixExpression)
	p.registerInfix(golexer.GREATER_THAN, p.parseInfixExpression)
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
		return p.parseLteStatment()
	case golexer.RETURN:
		return p.parseReturnStatment()
	case golexer.IDENT:
		return p.parseExpressionStatment()
	case golexer.NUMBER:
		return p.parseExpressionStatment()
	default:
		return p.parseExpressionStatment()
	}

}
func (p *Parser) parseLteStatment() *LetStatement {
	stmt := &LetStatement{Token: p.curToken}
	if !p.expectPeek(golexer.IDENT) {
		return nil
	}
	stmt.Name = &Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	if !p.expectPeek(golexer.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if !p.peekTokenIs(golexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
func (p *Parser) parseReturnStatment() *ReturnStatement {
	stmt := &ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	if !p.peekTokenIs(golexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
func (p *Parser) parseExpressionStatment() *ExpressionStatement {
	stmt := &ExpressionStatement{Token: p.curToken}
	p.nextToken()
	stmt.Expression = p.parseExpression(LOWEST)
	if !p.peekTokenIs(golexer.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
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
		return nil
	}
	return exp
}
