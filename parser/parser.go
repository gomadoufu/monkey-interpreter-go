package parser

import (
	"fmt"
	"gomadoufu/monkey-interpreter-go/ast"
	"gomadoufu/monkey-interpreter-go/lexer"
	"gomadoufu/monkey-interpreter-go/token"
	"strconv"
)

// 優先順位
const (
	_           int = iota
	LOWEST          //最も低い優先順位
	EQUALS          // ==
	LESSGREATER     // > or <
	SUM             // +
	PRODUCT         // *
	PREFIX          // -X or !X
	CALL            // myFunction(X)
)

// 演算子優先順位テーブル
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

type Parser struct {
	// 字句解析機へのポインタ
	l      *lexer.Lexer
	errors []string

	// 現在調べているトークン
	curToken token.Token
	// curTokenだけで判断がつかない時に見る、curTokenの次のトークン
	peekToken token.Token

	// curToken.Typeに関連づけられた構文解析関数を検索するためのマップ
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	// Parserにある2つのマップを初期化し、それぞれのトークンに対応する構文解析関数を登録する
	// すべての構文解析関数は、関連づけられたトークンがcurTokenにセットされている状態で動作を開始する。そして、この関数の処理対象である式の一番最後のトークンがcurTokenにセットされた状態になるまで進んで終了する。
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)

	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	p.registerInfix(token.LPAREN, p.parseCallExpression)

	p.registerPrefix(token.STRING, p.parseStringLiteral)

	//２つトークンを読み込む。curTokenとpeekTokenの両方がセットされる
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) ParseProgram() *ast.Program {
	// ASTのルートノードを生成
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// EOFに達するまで、入力のトークンを繰り返し読む
	for p.curToken.Type != token.EOF {
		// 文を構文解析する
		stmt := p.parseStatement()
		// 文がnilでなければ、Statementsに追加する
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// curTokenとpeekTokenを進めるヘルパーメソッド
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	// もし現在のトークンがLETなら、LetStatementを構文解析する
	case token.LET:
		return p.parseLetStatement()
	// もし現在のトークンがRETURNなら、ReturnStatementを構文解析する
	case token.RETURN:
		return p.parseReturnStatement()
	// それ以外なら、式文を構文解析する
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	// LETトークンに基づいた、LetStatement ASTノードを構築
	stmt := &ast.LetStatement{Token: p.curToken}

	// 文法チェックをしつつ進める
	// 識別子(変数名)を期待する
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// 識別子ノードを構築
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 等号を期待する
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// 文法チェック用のアサーション関数
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		// 不正な入力ならエラー
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

// expectPeek関数で期待した値が現れなかった時に呼ばれる
// エラーメッセージをerrorsに追加することで、親オブジェクトにエラーを伝搬する
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	// RETURNトークンに基づいた、ReturnStatement ASTノードを構築
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

type (
	prefixParseFn func() ast.Expression               // 前置構文解析関数
	infixParseFn  func(ast.Expression) ast.Expression // 中置構文解析関数
)

// 前置演算子用の構文解析関数を登録する
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// 中置演算子用の構文解析関数を登録する
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	// defer untrace(trace("parseExpressionStatement"))

	// 式文のトークンに基づいた、ExpressionStatement ASTノードを構築
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	// 構文解析関数を呼び出す
	stmt.Expression = p.parseExpression(LOWEST)

	// 省略可能なセミコロンをチェックする
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	// defer untrace(trace("parseExpression"))
	// p.curToken.Typeの前置に対応する構文解析関数があるか、マップをチェック
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		// もしなければ、nilを返す
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	// もしあれば、呼び出して、その結果を後に返す
	leftExp := prefix()

	//次のトークンの左結合力が現在の右結合力よりも高いかを判定する
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// 構文解析関数。現在のトークンをTokenフィールドに、トークンのリテラル値をValueフィールドに格納する。
func (p *Parser) parseIdentifier() ast.Expression {
	// 識別子のトークンに基づいた、Identifier ASTノードを構築して返す
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// 構文解析関数。ast.IntegerLiteralのValueフィールドに格納するために、p.curToken.Literalの文字列をstrconv.ParseIntでint64に変換する。
func (p *Parser) parseIntegerLiteral() ast.Expression {
	// defer untrace(trace("parseIntegerLiteral"))
	// 整数リテラルのトークンに基づいた、IntegerLiteral ASTノードを構築
	lit := &ast.IntegerLiteral{Token: p.curToken}

	// リテラル値をint64に変換
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// 見やすいエラーメッセージを出力するためのヘルパーメソッド
// フォーマットしたエラーメッセージをerrorsフィールドに追加する
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// 前置演算子用の構文解析関数。
func (p *Parser) parsePrefixExpression() ast.Expression {
	// defer untrace(trace("parsePrefixExpression"))
	// 前置演算子のトークンに基づいた、PrefixExpression ASTノードを構築
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	// トークンを消費する
	// ここで、p.curTokenは前置演算子のトークンになっている
	p.nextToken()
	// ここで、p.curTokenは前置演算子の右辺になるトークンになっている
	// こうすることで、"-5"のような式を正しくパースできる

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// p.peekTokenのトークンタイプに対応している優先順位を、テーブルから探して返す
// もし見つけられなければLOWESTを返す
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// p.curTokenのトークンタイプに対応している優先順位を、テーブルから探して返す
// もし見つけられなければLOWESTを返す
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

// 中置演算子用の構文解析関数。
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	// defer untrace(trace("parseInfixExpression"))
	// 中置演算子のトークンに基づいた、InfixExpression ASTノードを構築
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		// parsePrefixExpressionと違い、中置演算子の左辺になる式を、left引数に渡す
		Left: left,
	}

	// 現在のトークンである、中置演算子そのものの優先順位を保存しておく
	precedence := p.curPrecedence()
	// トークンを進める
	p.nextToken()
	// parseInfixExpressionを再度呼び出し、ast.InfixExpressionのRightフィールドを埋める
	expression.Right = p.parseExpression(precedence)

	return expression
}

// 真偽値用の構文解析関数。
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// かっこ()で囲まれた(グループ化された）式をパースするための構文解析関数。
// "トークンタイプに関数を関連づけるという考え方がここにきて本当に輝くんだ！"
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// if式をパースするための構文解析関数。
func (p *Parser) parseIfExpression() ast.Expression {
	// defer untrace(trace("parseIfExpression"))
	// if式のトークンに基づいた、IfExpression ASTノードを構築
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// { }で囲まれたブロックをパースするための構文解析関数。
// parseIfExpressionとparseFunctionLiteralで使われる
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	// defer untrace(trace("parseBlockStatement"))
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// 関数リテラルをパースするための構文解析関数。
func (p *Parser) parseFunctionLiteral() ast.Expression {
	// defer untrace(trace("parseFunctionLiteral"))
	// 関数リテラルのトークンに基づいた、FunctionLiteral ASTノードを構築
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// 関数の引数リストをパースするための構文解析関数。
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

// 関数呼び出しをパースするための構文解析関数。
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	// defer untrace(trace("parseCallExpression"))
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// 関数呼び出し時の引数リストをパースするための構文解析関数。
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return args
}

// 文字列リテラルをパースするための構文解析関数。
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}
