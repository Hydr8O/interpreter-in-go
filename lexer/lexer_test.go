package lexer

import "testing"
import "interpreter/token"

func TestNextToken(t *testing.T) {
	input := `
   let five = 5;
   let add = fn(x, y) {
      x + y;
   };
   !-/*5;
   5 < 10 > 5;
   true false if else return == !=;
   `
	testCases := []struct {
		expectedTokenType string
		expectedLiteral   string
	}{
		{token.LET, "let"},
		{token.IDENTIFIER, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENTIFIER, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENTIFIER, "x"},
		{token.COMMA, ","},
		{token.IDENTIFIER, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENTIFIER, "x"},
		{token.PLUS, "+"},
		{token.IDENTIFIER, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LTHAN, "<"},
		{token.INT, "10"},
		{token.GTHAN, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.TRUE, "TRUE"},
		{token.FALSE, "FALSE"},
		{token.IF, "IF"},
		{token.ELSE, "ELSE"},
		{token.RETURN, "RETURN"},
		{token.EQ, "=="},
		{token.NOT_EQ, "!=="},
		{token.SEMICOLON, ";"},
	}

	lexer := New(input)
	for _, test := range testCases {
		token := lexer.NextToken()
		if token.Type != test.expectedTokenType {
			t.Fatalf("Token type is wrong. Expected: %q, Got: %q", test.expectedTokenType, token.Type)
		}
	}
}
