package ast

import (
	"bytes"
	"gomadoufu/monkey-interpreter-go/token"
	"strings"
)

// ASTノード
type Node interface {
	TokenLiteral() string
	// デバッグ用のメソッド
	String() string
}

// 文
type Statement interface {
	Node
	// 開発用のダミーメソッド
	statementNode()
}

// 式
type Expression interface {
	Node
	// 開発用のダミーメソッド
	expressionNode()
}

// ASTのルートノード
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

// バッファを作成し、それぞれの文のString()メソッドの戻り値を書き込む
func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// let文 let x = 5;
type LetStatement struct {
	//let文であることを示すtoken.LETトークン
	Token token.Token
	// 左辺の識別子(変数名)を保持する
	Name *Identifier
	// 右辺の式を保持する
	Value Expression
}

// Statementインターフェイスを満たす
func (ls *LetStatement) statementNode() {}

// Nodeインターフェイスを満たす
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

// ast.Program.String()に呼ばれる
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

// 識別子
type Identifier struct {
	//token.IDENTトークン
	Token token.Token
	// 識別子自身の文字列表現
	Value string
}

// 識別子は式ではないが、簡単のため式として扱う
func (i *Identifier) expressionNode() {}

// Nodeインターフェイスを満たす
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// ast.Program.String()に呼ばれる
func (i *Identifier) String() string { return i.Value }

// return文
type ReturnStatement struct {
	// 'return' トークン
	Token token.Token
	// returnの後に続く返り値の式
	ReturnValue Expression
}

// Statementインターフェイスを満たす
func (rs *ReturnStatement) statementNode() {}

// Nodeインターフェイスを満たす
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// ast.Program.String()に呼ばれる
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

// 式文
type ExpressionStatement struct {
	//式の最初のトークン
	Token token.Token
	// 式そのもの
	Expression Expression
}

// Statementインターフェイスを満たす
func (es *ExpressionStatement) statementNode() {}

// Nodeインターフェイスを満たす
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// ast.Program.String()に呼ばれる
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// 整数リテラル
type IntegerLiteral struct {
	// INTトークン
	Token token.Token
	// 整数リテラルが表現している実際の整数の値
	Value int64
}

// Expressionインターフェイスを満たす
func (il *IntegerLiteral) expressionNode() {}

// Nodeインターフェイスを満たす
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// ast.Program.String()に呼ばれる
func (il *IntegerLiteral) String() string { return il.Token.Literal }

// 前置演算子
type PrefixExpression struct {
	//前置トークン、例えば「!」
	Token token.Token
	// 演算子そのもの
	Operator string
	// 演算子の右側の式
	Right Expression
}

// Expressionインターフェイスを満たす
func (pe *PrefixExpression) expressionNode() {}

// Nodeインターフェイスを満たす
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

// ast.Program.String()に呼ばれる
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

// 中置演算子
type InfixExpression struct {
	//演算子トークン、例えば「+」
	Token token.Token
	// 演算子の左側の式
	Left Expression
	// 演算子そのもの
	Operator string
	// 演算子の右側の式
	Right Expression
}

// Expressionインターフェイスを満たす
func (oe *InfixExpression) expressionNode() {}

// Nodeインターフェイスを満たす
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }

// ast.Program.String()に呼ばれる
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" " + oe.Operator + " ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")

	return out.String()
}

// 真偽値
type Boolean struct {
	// TRUE or FALSE トークン
	Token token.Token
	// (Go言語の)真偽値
	Value bool
}

// Expressionインターフェイスを満たす
func (b *Boolean) expressionNode() {}

// Nodeインターフェイスを満たす
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }

// ast.Program.String()に呼ばれる
func (b *Boolean) String() string { return b.Token.Literal }

// if式
type IfExpression struct {
	// 'if' トークン
	Token token.Token
	// ifの後に続く条件式
	Condition Expression
	// ifの後に続く条件式の後に続くブロック
	Consequence *BlockStatement
	// elseの後に続く条件式の後に続くブロック
	Alternative *BlockStatement
}

// Expressionインターフェイスを満たす
func (ie *IfExpression) expressionNode() {}

// Nodeインターフェイスを満たす
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }

// ast.Program.String()に呼ばれる
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

type BlockStatement struct {
	// '{' トークン
	Token token.Token
	// ブロック内の文
	Statements []Statement
}

// Statementインターフェイスを満たす
func (bs *BlockStatement) statementNode() {}

// Nodeインターフェイスを満たす
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// ast.Program.String()に呼ばれる
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// 関数リテラル
type FunctionLiteral struct {
	// 'fn' トークン
	Token token.Token
	// 引数リスト
	Parameters []*Identifier
	// 関数の本体
	Body *BlockStatement
}

// Expressionインターフェイスを満たす
func (fl *FunctionLiteral) expressionNode() {}

// Nodeインターフェイスを満たす
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }

// ast.Program.String()に呼ばれる
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString(fl.Body.String())

	return out.String()
}

// 関数呼び出し
type CallExpression struct {
	// '(' トークン
	Token token.Token
	// 関数名
	Function Expression
	// 引数リスト
	Arguments []Expression
}

// Expressionインターフェイスを満たす
func (ce *CallExpression) expressionNode() {}

// Nodeインターフェイスを満たす
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }

// ast.Program.String()に呼ばれる
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}
