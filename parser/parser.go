package parser

import (
	"fmt"
	"interpreter/ast"
	"interpreter/lexer"
	"interpreter/token"
	"strconv"
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
	INDEX
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	lex          *lexer.Lexer
	currentToken token.Token
	peekToken    token.Token
	errors       []string

	prefixParseFns map[string]prefixParseFn
	infixParseFns  map[string]infixParseFn
}

var precedences = map[string]int{
	token.LPAREN:   CALL,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LTHAN:    LESSGREATER,
	token.GTHAN:    LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LBRACKET: INDEX,
}

func (parser *Parser) PeekPrecedence() int {
	if precedence, ok := precedences[parser.peekToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (parser *Parser) CurrentPrecedence() int {
	if precedence, ok := precedences[parser.currentToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func New(lex *lexer.Lexer) *Parser {
	parser := &Parser{lex: lex, errors: []string{}}
	parser.NextToken()
	parser.NextToken()

	parser.prefixParseFns = make(map[string]prefixParseFn)
	parser.RegisterPrefix(token.IDENTIFIER, parser.ParseIdentifier)
	parser.RegisterPrefix(token.INT, parser.ParseIntegerLiteral)
	parser.RegisterPrefix(token.BANG, parser.ParsePrefixExpression)
	parser.RegisterPrefix(token.MINUS, parser.ParsePrefixExpression)
	parser.RegisterPrefix(token.TRUE, parser.ParseBoolean)
	parser.RegisterPrefix(token.FALSE, parser.ParseBoolean)
	parser.RegisterPrefix(token.LPAREN, parser.ParseGroup)
	parser.RegisterPrefix(token.IF, parser.ParseIfExpression)
	parser.RegisterPrefix(token.FUNCTION, parser.ParseFunctionLiteral)
	parser.RegisterPrefix(token.STRING, parser.ParseStringLiteral)
	parser.RegisterPrefix(token.LBRACKET, parser.ParseArrayLiteral)

	parser.infixParseFns = make(map[string]infixParseFn)
	parser.RegisterInfix(token.PLUS, parser.ParseInfixExpression)
	parser.RegisterInfix(token.MINUS, parser.ParseInfixExpression)
	parser.RegisterInfix(token.SLASH, parser.ParseInfixExpression)
	parser.RegisterInfix(token.ASTERISK, parser.ParseInfixExpression)
	parser.RegisterInfix(token.EQ, parser.ParseInfixExpression)
	parser.RegisterInfix(token.NOT_EQ, parser.ParseInfixExpression)
	parser.RegisterInfix(token.LTHAN, parser.ParseInfixExpression)
	parser.RegisterInfix(token.GTHAN, parser.ParseInfixExpression)
	parser.RegisterInfix(token.LPAREN, parser.ParseCallExpression)
	parser.RegisterInfix(token.LBRACKET, parser.ParseIndexExpression)
	return parser
}

func (parser *Parser) ParseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: parser.currentToken, Left: left}

	parser.NextToken()
	exp.Index = parser.ParseExpression(LOWEST)

	if !parser.ExpectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (parser *Parser) ParseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: parser.currentToken}

	array.Elements = parser.ParseExpressionList(token.RBRACKET)
	return array
}

func (parser *Parser) ParseExpressionList(end string) []ast.Expression {
	list := []ast.Expression{}

	if parser.peekToken.Type == end {
		parser.NextToken()
		return list
	}

	parser.NextToken()
	list = append(list, parser.ParseExpression(LOWEST))

	for parser.peekToken.Type == token.COMMA {
		parser.NextToken()
		parser.NextToken()
		list = append(list, parser.ParseExpression(LOWEST))
	}

	if !parser.ExpectPeek(end) {
		return nil
	}

	return list
}

func (parser *Parser) ParseCallExpression(function ast.Expression) ast.Expression {
	expression := &ast.CallExpression{Token: parser.currentToken, Function: function}
	expression.Arguments = parser.ParseExpressionList(token.RPAREN)
	return expression
}

func (parser *Parser) ParseFunctionLiteral() ast.Expression {
	literal := &ast.FunctionLiteral{Token: parser.currentToken}

	if !parser.ExpectPeek(token.LPAREN) {
		return nil
	}

	literal.Parameters = parser.ParseFunctionParameters()

	if !parser.ExpectPeek(token.LBRACE) {
		return nil
	}

	literal.Body = parser.ParseBlockStatement()

	return literal
}

func (parser *Parser) ParseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if parser.peekToken.Type == token.RPAREN {
		parser.NextToken()
		return identifiers
	}

	parser.NextToken()
	identifier := &ast.Identifier{Token: parser.currentToken, Value: parser.currentToken.Literal}
	identifiers = append(identifiers, identifier)

	for parser.peekToken.Type == token.COMMA {
		parser.NextToken()
		parser.NextToken()
		identifier := &ast.Identifier{Token: parser.currentToken, Value: parser.currentToken.Literal}
		identifiers = append(identifiers, identifier)
	}

	if !parser.ExpectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (parser *Parser) ParseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: parser.currentToken}

	if !parser.ExpectPeek(token.LPAREN) {
		return nil
	}

	parser.NextToken()

	expression.Condition = parser.ParseExpression(LOWEST)

	if !parser.ExpectPeek(token.RPAREN) {
		return nil
	}

	if !parser.ExpectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = parser.ParseBlockStatement()

	if parser.peekToken.Type == token.ELSE {
		parser.NextToken()

		if !parser.ExpectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = parser.ParseBlockStatement()
	}
	return expression
}

func (parser *Parser) ParseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: parser.currentToken}
	block.Statements = []ast.Statement{}

	parser.NextToken()

	for parser.currentToken.Type != token.RBRACE && parser.currentToken.Type != token.EOF {
		statement := parser.ParseStatement()
		if statement != nil {
			block.Statements = append(block.Statements, statement)
		}
		parser.NextToken()
	}

	return block
}

func (parser *Parser) ParseGroup() ast.Expression {
	parser.NextToken()

	expression := parser.ParseExpression(LOWEST)

	if !parser.ExpectPeek(token.RPAREN) {
		return nil
	}
	return expression
}

func (parser *Parser) ParseIdentifier() ast.Expression {
	return &ast.Identifier{Token: parser.currentToken, Value: parser.currentToken.Literal}
}

func (parser *Parser) ParseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: parser.currentToken}

	value, err := strconv.ParseInt(parser.currentToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", parser.currentToken.Literal)
		parser.errors = append(parser.errors, msg)
	}

	literal.Value = value
	return literal
}

func (parser *Parser) ParseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: parser.currentToken, Value: parser.currentToken.Literal}
}

func (parser *Parser) ParseBoolean() ast.Expression {
	boolean := &ast.Boolean{Token: parser.currentToken, Value: parser.currentToken.Type == token.TRUE}
	return boolean
}

func (parser *Parser) ParsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{Token: parser.currentToken, Operator: parser.currentToken.Literal}

	parser.NextToken()
	expression.Right = parser.ParseExpression(PREFIX)

	return expression
}

func (parser *Parser) ParseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    parser.currentToken,
		Operator: parser.currentToken.Literal,
		Left:     left,
	}

	precedence := parser.CurrentPrecedence()
	parser.NextToken()
	expression.Right = parser.ParseExpression(precedence)

	return expression
}

func (parser *Parser) RegisterPrefix(tokenType string, fn prefixParseFn) {
	parser.prefixParseFns[tokenType] = fn
}

func (parser *Parser) RegisterInfix(tokenType string, fn infixParseFn) {
	parser.infixParseFns[tokenType] = fn
}

func (parser *Parser) Errors() []string {
	return parser.errors
}

func (parser *Parser) AddError(tokenType string) {
	error := fmt.Sprintf("expected next token to be %s, got %s instead", tokenType, parser.peekToken.Type)
	parser.errors = append(parser.errors, error)
}

func (parser *Parser) NextToken() {
	parser.currentToken = parser.peekToken
	parser.peekToken = parser.lex.NextToken()
}

func (parser *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for parser.currentToken.Type != token.EOF {
		statement := parser.ParseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		parser.NextToken()
	}
	return program
}

func (parser *Parser) ParseStatement() ast.Statement {
	switch parser.currentToken.Type {
	case token.LET:
		return parser.ParseLetStatement()
	case token.RETURN:
		return parser.ParseReturnStatement()
	default:
		return parser.ParseExpressionStatement()
	}
}

func (parser *Parser) ParseLetStatement() *ast.LetStatement {
	statement := &ast.LetStatement{Token: parser.currentToken}
	if !parser.ExpectPeek(token.IDENTIFIER) {
		return nil
	}
	statement.Name = &ast.Identifier{Token: parser.currentToken, Value: parser.currentToken.Literal}

	if !parser.ExpectPeek(token.ASSIGN) {
		return nil
	}

	parser.NextToken()
	statement.Value = parser.ParseExpression(LOWEST)

	if parser.peekToken.Type == token.SEMICOLON {
		parser.NextToken()
	}
	return statement
}

func (parser *Parser) ParseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{Token: parser.currentToken}

	parser.NextToken()

	statement.ReturnValue = parser.ParseExpression(LOWEST)

	if parser.peekToken.Type == token.SEMICOLON {
		parser.NextToken()
	}
	return statement
}

func (parser *Parser) ParseExpressionStatement() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{Token: parser.currentToken}

	statement.Expression = parser.ParseExpression(LOWEST)

	if parser.peekToken.Type == token.SEMICOLON {
		parser.NextToken()
	}

	return statement
}

func (parser *Parser) ParseExpression(precedence int) ast.Expression {
	prefix := parser.prefixParseFns[parser.currentToken.Type]
	if prefix == nil {
		msg := fmt.Sprintf("no prefix parse function for %s found", parser.currentToken.Type)
		parser.errors = append(parser.errors, msg)
		return nil
	}

	leftExpression := prefix()
	for parser.peekToken.Type != token.SEMICOLON && precedence < parser.PeekPrecedence() {
		infix := parser.infixParseFns[parser.peekToken.Type]
		if infix == nil {
			return leftExpression
		}

		parser.NextToken()
		leftExpression = infix(leftExpression)
	}

	return leftExpression
}

func (parser *Parser) ExpectPeek(nextTokenType string) bool {
	if parser.peekToken.Type == nextTokenType {
		parser.NextToken()
		return true
	}
	parser.AddError(nextTokenType)
	return false
}
