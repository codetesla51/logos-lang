package main

import (
	"fmt"

	"github.com/codetesla51/golexer/golexer"
	"github.com/codetesla51/logos/parser"
)

func main() {
	// prog := &parser.Program{
	// 	Statements: []parser.Statement{
	// 		&parser.Identifier{Token: golexer.Token{Type: golexer.IDENT, Literal: "X"}, Value: "X"},
	// 		&parser.LetStatement{
	// 			Token: golexer.Token{Type: golexer.LET, Literal: "let"},
	// 			Name:  &parser.Identifier{Token: golexer.Token{Type: golexer.IDENT, Literal: "y"}, Value: "y"},
	// 			Value: &parser.InfixExpression{
	// 				Token:    golexer.Token{Type: golexer.PLUS, Literal: "+"},
	// 				Operator: "+",
	// 				Left:     &parser.IntegerLiteral{Token: golexer.Token{Type: golexer.NUMBER, Literal: "5"}, Value: 5},
	// 				Right:    &parser.IntegerLiteral{Token: golexer.Token{Type: golexer.NUMBER, Literal: "3"}, Value: 3},
	// 			},
	// 		},
	// 		&parser.ReturnStatement{
	// 			Token:       golexer.Token{Type: golexer.RETURN, Literal: "return"},
	// 			ReturnValue: &parser.Identifier{Token: golexer.Token{Type: golexer.IDENT, Literal: "y"}, Value: "y"},
	// 		},
	// 		&parser.ExpressionStatement{
	// 			Token: golexer.Token{Type: golexer.IDENT, Literal: "foobar"},
	// 			Expression: &parser.InfixExpression{
	// 				Token:    golexer.Token{Type: golexer.PLUS, Literal: "+"},
	// 				Operator: "+",
	// 				Left:     &parser.Identifier{Token: golexer.Token{Type: golexer.IDENT, Literal: "foobar"}, Value: "foobar"},
	// 				Right:    &parser.Identifier{Token: golexer.Token{Type: golexer.IDENT, Literal: "y"}, Value: "y"},
	// 			},
	// 		},
	// 	},
	// }
	// fmt.Println(prog.String())

	input := `let x = 0.5 + 3; return y; foobar + y;`
	lexer := golexer.NewLexer(input)
	parser := parser.NewParser(lexer)
	program := parser.Parse()
	fmt.Println(program.String())
}
