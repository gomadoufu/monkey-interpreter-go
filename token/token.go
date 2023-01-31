// token/token.go

package token

type TokenType string

// NOTE:ファイル名や行番号を付与するアイデアもある(Rustではやってみる)
type Token struct {
	Type    TokenType
	Literal string
}

const (
	//トークンや文字が未知であるとき使う
	ILLEGAL = "ILLEGAL"
	//ファイル終端
	EOF = "EOF"

	// Identifiers + literals
	IDENT = "IDENT" // add, foobar, x, y, ...
	INT   = "INT"   // 1343456

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
)
