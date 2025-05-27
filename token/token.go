package token

type Token struct {
	Type    string
	Literal string
}

const (
	LBRACKET   = "["
	RBRACKET   = "]"
	STRING     = "STRING"
	ILLEGAL    = "ILLEGAL"
	EOF        = "EOF"
	IDENTIFIER = "IDENTIFIER"
	INT        = "INT"
	ASSIGN     = "="
	PLUS       = "+"
	COMMA      = ","
	SEMICOLON  = ";"
	LPAREN     = "("
	RPAREN     = ")"
	LBRACE     = "{"
	RBRACE     = "}"
	FUNCTION   = "FUNCTION"
	LET        = "LET"
	BANG       = "!"
	ASTERISK   = "*"
	SLASH      = "/"
	LTHAN      = "<"
	GTHAN      = ">"
	MINUS      = "-"
	TRUE       = "TRUE"
	FALSE      = "FALSE"
	IF         = "IF"
	ELSE       = "ELSE"
	RETURN     = "RETURN"
	EQ         = "=="
	NOT_EQ     = "!="
)
