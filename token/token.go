package token

// トークンタイプ = 識別子 | キーワード | 記号 | ILLEGAL | EOF
// 識別子 = 数や変数名など、ユーザが決定するもの。字句解析や構文解析の段階では、識別子であることさえわかれば良い
// キーワード = if, else, true, false, return, let, fn などの予約語。識別子に見えるが、実際は言語の一部であるもの
// 記号 = +, -, *, /, =, ==, !=, <, >, !, (, ), {, }, ;, , などの記号
type TokenType string

// NOTE:ファイル名や行番号を付与するアイデアもある(Rustではやってみる)

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// 識別子 + リテラル
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	INT    = "INT"    // 1343456
	STRING = "STRING" // "foobar"

	// 演算子
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

	// デリミタ
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	// キーワード
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
)

// トークン = トークンタイプ + リテラル
// リテラル = トークンの値。AST構築の時まで、実際のトークンが何であったか保持する
type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

// 渡された識別子がキーワードかどうかを判定する
func LookupIdent(ident string) TokenType {
	// キーワードのTokenType定数を返す
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	// ユーザ定義識別子に対応するTokenTypeを返す
	return IDENT
}
