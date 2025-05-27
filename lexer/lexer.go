package lexer

import "interpreter/token"

type Lexer struct {
	input        string
	position     int
	readPosition int
	currentChar  byte
}

func (lexer *Lexer) ReadChar() {
	if lexer.readPosition >= len(lexer.input) {
		lexer.currentChar = 0
	} else {
		lexer.currentChar = lexer.input[lexer.readPosition]
	}
	lexer.position = lexer.readPosition
	lexer.readPosition += 1
}

func (lexer *Lexer) ReadIdentifier() string {
	position := lexer.position
	for IsLetter(lexer.currentChar) {
		lexer.ReadChar()
	}
	return lexer.input[position:lexer.position]
}

func (lexer *Lexer) ReadNumber() string {
	position := lexer.position
	for IsDigit(lexer.currentChar) {
		lexer.ReadChar()
	}
	return lexer.input[position:lexer.position]
}

func (lexer *Lexer) PeekChar() byte {
	if lexer.readPosition >= len(lexer.input) {
		return 0
	}
	return lexer.input[lexer.readPosition]
}

func New(input string) *Lexer {
	lexer := &Lexer{input: input}
	lexer.ReadChar()
	return lexer
}

func (lexer *Lexer) NextToken() token.Token {
	var nextToken token.Token
	lexer.SkipWhitespaces()
	switch lexer.currentChar {
	case '=':
		if lexer.PeekChar() == '=' {
			nextToken = token.Token{Type: token.EQ, Literal: "=="}
			lexer.ReadChar()
		} else {
			nextToken = NewToken(token.ASSIGN, lexer.currentChar)
		}
	case ';':
		nextToken = NewToken(token.SEMICOLON, lexer.currentChar)
	case '(':
		nextToken = NewToken(token.LPAREN, lexer.currentChar)
	case ')':
		nextToken = NewToken(token.RPAREN, lexer.currentChar)
	case ',':
		nextToken = NewToken(token.COMMA, lexer.currentChar)
	case '+':
		nextToken = NewToken(token.PLUS, lexer.currentChar)
	case '-':
		nextToken = NewToken(token.MINUS, lexer.currentChar)
	case '*':
		nextToken = NewToken(token.ASTERISK, lexer.currentChar)
	case '!':
		if lexer.PeekChar() == '=' {
			nextToken = token.Token{Type: token.NOT_EQ, Literal: "!="}
			lexer.ReadChar()
		} else {
			nextToken = NewToken(token.BANG, lexer.currentChar)
		}
	case '/':
		nextToken = NewToken(token.SLASH, lexer.currentChar)
	case '<':
		nextToken = NewToken(token.LTHAN, lexer.currentChar)
	case '>':
		nextToken = NewToken(token.GTHAN, lexer.currentChar)
	case '{':
		nextToken = NewToken(token.LBRACE, lexer.currentChar)
	case '}':
		nextToken = NewToken(token.RBRACE, lexer.currentChar)
	case '"':
		nextToken.Type = token.STRING
		nextToken.Literal = lexer.ReadString()
	case '[':
		nextToken = NewToken(token.LBRACKET, lexer.currentChar)
	case ']':
		nextToken = NewToken(token.RBRACKET, lexer.currentChar)
	case 0:
		nextToken.Literal = ""
		nextToken.Type = token.EOF
	default:
		if IsLetter(lexer.currentChar) {
			nextToken.Literal = lexer.ReadIdentifier()
			tokenMap := map[string]string{
				"let":    token.LET,
				"fn":     token.FUNCTION,
				"true":   token.TRUE,
				"false":  token.FALSE,
				"if":     token.IF,
				"else":   token.ELSE,
				"return": token.RETURN,
			}
			if tokenType, exists := tokenMap[nextToken.Literal]; exists {
				nextToken.Type = tokenType
			} else {
				nextToken.Type = token.IDENTIFIER
			}
			return nextToken
		}
		if IsDigit(lexer.currentChar) {
			nextToken.Literal = lexer.ReadNumber()
			nextToken.Type = token.INT
			return nextToken
		}
		nextToken = NewToken(token.ILLEGAL, lexer.currentChar)
	}

	lexer.ReadChar()
	return nextToken
}

func (lexer *Lexer) ReadString() string {
	position := lexer.position + 1
	for {
		lexer.ReadChar()
		if lexer.currentChar == '"' || lexer.currentChar == 0 {
			break
		}
	}
	return lexer.input[position:lexer.position]
}

func NewToken(tokenType string, currentChar byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(currentChar)}
}

func IsLetter(char byte) bool {
	return 'a' <= char && char <= 'z' || 'A' <= char && char <= 'Z' || char == '_'
}

func IsDigit(char byte) bool {
	return '0' <= char && char <= '9'
}

func (lexer *Lexer) SkipWhitespaces() {
	for IsWhitespace(lexer.currentChar) {
		lexer.ReadChar()
	}
}

func IsWhitespace(char byte) bool {
	return char == ' ' || char == '\n' || char == '\t' || char == '\r'
}
