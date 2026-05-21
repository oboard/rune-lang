package parser

import (
	"fmt"
	"strconv"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

type Error struct {
	Message string
	Pos     lexer.Position
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

type Parser struct {
	tokens []lexer.Token
	curr   int
	errors []Error
}

func Parse(src string) (*ast.File, []Error) {
	return New(lexer.Lex(src)).ParseFile()
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) ParseFile() (*ast.File, []Error) {
	file := &ast.File{}
	p.skipNewlines()
	for !p.check(lexer.EOF) {
		if p.check(lexer.At) {
			if imp := p.parseGoImportDecl(); imp != nil {
				file.GoImports = append(file.GoImports, *imp)
			}
		} else if p.check(lexer.Ident) && p.checkNext(lexer.Colon) {
			typ := p.parseStructType()
			if typ != nil {
				file.Types = append(file.Types, typ)
			}
		} else {
			fn := p.parseFunction()
			if fn != nil {
				file.Functions = append(file.Functions, fn)
			}
		}
		p.skipNewlines()
	}
	return file, p.errors
}

func (p *Parser) parseGoImportDecl() *ast.GoImport {
	at := p.consume(lexer.At, "expected '@'")
	module := p.consume(lexer.Ident, "expected module name after '@'")
	if module.Lexeme != "go" {
		p.errorAt(module, "only @go.import can appear at the top level")
		return nil
	}
	p.consume(lexer.Dot, "expected '.' after @go")
	name := p.consume(lexer.Ident, "expected import after @go.")
	if name.Lexeme != "import" {
		p.errorAt(name, "only @go.import can appear at the top level")
		return nil
	}
	p.consume(lexer.LParen, "expected '(' after @go.import")
	path := p.consume(lexer.String, "expected Go import path string")
	value, err := strconv.Unquote(path.Lexeme)
	if err != nil {
		p.errorAt(path, "invalid Go import path string")
	}
	p.consume(lexer.RParen, "expected ')' after @go.import")
	return &ast.GoImport{Path: value, Pos: at.Pos}
}

func (p *Parser) parseStructType() *ast.StructType {
	name := p.consume(lexer.Ident, "expected type name")
	if name.Kind == lexer.EOF {
		return nil
	}
	typ := &ast.StructType{Name: name.Lexeme, Pos: name.Pos, NamePos: name.Pos}
	p.consume(lexer.Colon, "expected ':' after type name")
	p.skipNewlines()
	p.consume(lexer.LBrace, "expected '{' after type declaration")
	p.skipNewlines()
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		if p.check(lexer.Ident) && p.checkNext(lexer.LParen) {
			method := p.parseFunctionWithReceiver(typ.Name)
			if method != nil {
				typ.Methods = append(typ.Methods, method)
			}
			p.consumeStatementEnd()
			p.skipNewlines()
			continue
		}
		fieldName := p.consume(lexer.Ident, "expected field name")
		p.consume(lexer.Colon, "expected ':' after field name")
		fieldType := p.consume(lexer.Ident, "expected field type")
		typ.Fields = append(typ.Fields, ast.Field{
			Name: fieldName.Lexeme,
			Type: fieldType.Lexeme,
			Pos:  fieldName.Pos,
		})
		p.consumeStatementEnd()
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after type declaration")
	return typ
}

func (p *Parser) parseFunction() *ast.Function {
	return p.parseFunctionWithReceiver("")
}

func (p *Parser) parseFunctionWithReceiver(receiverType string) *ast.Function {
	name := p.consume(lexer.Ident, "expected function name")
	if name.Kind == lexer.EOF {
		return nil
	}

	fn := &ast.Function{Name: name.Lexeme, ReceiverType: receiverType, Pos: name.Pos, NamePos: name.Pos}
	p.consume(lexer.LParen, "expected '(' after function name")
	p.skipNewlines()
	if !p.check(lexer.RParen) {
		for {
			paramName := p.consume(lexer.Ident, "expected parameter name")
			p.consume(lexer.Colon, "expected ':' after parameter name")
			paramType := p.consume(lexer.Ident, "expected parameter type")
			fn.Params = append(fn.Params, ast.Param{
				Name: paramName.Lexeme,
				Type: paramType.Lexeme,
				Pos:  paramName.Pos,
			})
			p.skipNewlines()
			if !p.match(lexer.Comma) {
				break
			}
			p.skipNewlines()
		}
	}
	p.consume(lexer.RParen, "expected ')' after parameter list")
	p.skipNewlines()
	if p.match(lexer.Arrow) {
		p.skipNewlines()
		returnType := p.consume(lexer.Ident, "expected return type after '->'")
		fn.ReturnType = returnType.Lexeme
		p.skipNewlines()
	}
	p.consume(lexer.FatArrow, "expected '=>' after function signature")
	p.skipNewlines()
	fn.Body = p.parseBody()
	return fn
}

func (p *Parser) parseBody() ast.Expr {
	if p.check(lexer.LBrace) {
		return p.parseBlock()
	}
	return p.parseExpression(1)
}

func (p *Parser) parseBlock() ast.Expr {
	start := p.consume(lexer.LBrace, "expected '{'")
	p.skipNewlines()
	if p.check(lexer.RBrace) {
		p.advance()
		return &ast.BlockExpr{Pos: start.Pos}
	}

	if p.looksLikePatternBranch() {
		block := &ast.PatternBlock{Pos: start.Pos}
		for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
			p.skipNewlines()
			if p.check(lexer.RBrace) {
				break
			}
			pattern := p.parsePattern()
			p.consume(lexer.FatArrow, "expected '=>' after pattern")
			p.skipNewlines()
			expr := p.parseExpression(1)
			block.Branches = append(block.Branches, ast.PatternBranch{
				Pattern: pattern,
				Expr:    expr,
				Pos:     pattern.Position(),
			})
			p.consumeStatementEnd()
		}
		p.consume(lexer.RBrace, "expected '}' after pattern block")
		return block
	}

	block := &ast.BlockExpr{Pos: start.Pos}
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		p.skipNewlines()
		if p.check(lexer.RBrace) {
			break
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.consumeStatementEnd()
	}
	p.consume(lexer.RBrace, "expected '}' after block")
	return block
}

func (p *Parser) parseStatement() ast.Stmt {
	if p.check(lexer.Ident) {
		if p.checkNext(lexer.Declare) || p.checkNext(lexer.MutDeclare) {
			name := p.advance()
			mutable := p.match(lexer.MutDeclare)
			if !mutable {
				p.consume(lexer.Declare, "expected ':=' after binding name")
			}
			p.skipNewlines()
			return &ast.LetStmt{
				Name:    name.Lexeme,
				Mutable: mutable,
				Value:   p.parseExpression(1),
				Pos:     name.Pos,
			}
		}
		if p.checkNext(lexer.Assign) {
			name := p.advance()
			p.advance()
			p.skipNewlines()
			return &ast.AssignStmt{
				Name:  name.Lexeme,
				Value: p.parseExpression(1),
				Pos:   name.Pos,
			}
		}
	}
	expr := p.parseExpression(1)
	return &ast.ExprStmt{Expr: expr, Pos: expr.Position()}
}

func (p *Parser) parsePattern() ast.Pattern {
	tok := p.peek()
	switch tok.Kind {
	case lexer.Underscore:
		p.advance()
		return &ast.WildcardPattern{Pos: tok.Pos}
	case lexer.Int, lexer.String:
		return &ast.LiteralPattern{Value: p.parsePrimary(), Pos: tok.Pos}
	case lexer.Ident:
		if tok.Lexeme == "true" || tok.Lexeme == "false" {
			return &ast.LiteralPattern{Value: p.parsePrimary(), Pos: tok.Pos}
		}
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		op := p.advance()
		value := p.parsePrimary()
		return &ast.ComparePattern{Op: op.Kind, Value: value, Pos: op.Pos}
	case lexer.LParen:
		start := p.advance()
		p.skipNewlines()
		var elems []ast.Pattern
		if !p.check(lexer.RParen) {
			for {
				elems = append(elems, p.parsePattern())
				p.skipNewlines()
				if !p.match(lexer.Comma) {
					break
				}
				p.skipNewlines()
			}
		}
		p.consume(lexer.RParen, "expected ')' after tuple pattern")
		return &ast.TuplePattern{Elements: elems, Pos: start.Pos}
	}

	p.errorAt(tok, "expected pattern")
	p.advance()
	return &ast.WildcardPattern{Pos: tok.Pos}
}

func (p *Parser) parseExpression(minPrec int) ast.Expr {
	left := p.parseUnary()
	for {
		if p.check(lexer.LBrace) {
			ident, ok := left.(*ast.Identifier)
			if !ok {
				break
			}
			left = p.parseStructLiteral(ident)
			continue
		}
		if p.match(lexer.LParen) {
			call := &ast.CallExpr{Callee: left, Pos: left.Position()}
			p.skipNewlines()
			if !p.check(lexer.RParen) {
				for {
					call.Args = append(call.Args, p.parseExpression(1))
					p.skipNewlines()
					if !p.match(lexer.Comma) {
						break
					}
					p.skipNewlines()
				}
			}
			p.consume(lexer.RParen, "expected ')' after arguments")
			left = call
			continue
		}
		if p.match(lexer.Dot) {
			name := p.consume(lexer.Ident, "expected selector name after '.'")
			left = &ast.SelectorExpr{Receiver: left, Name: name.Lexeme, Pos: left.Position()}
			continue
		}

		prec := precedence(p.peek().Kind)
		if prec < minPrec {
			break
		}
		op := p.advance()
		right := p.parseExpression(prec + 1)
		left = &ast.BinaryExpr{Left: left, Op: op.Kind, Right: right, Pos: left.Position()}
	}
	return left
}

func (p *Parser) parseStructLiteral(typeName *ast.Identifier) ast.Expr {
	lit := &ast.StructLiteral{TypeName: typeName.Name, Pos: typeName.Pos}
	p.consume(lexer.LBrace, "expected '{' after type name")
	p.skipNewlines()
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		fieldName := p.consume(lexer.Ident, "expected field name")
		p.consume(lexer.Colon, "expected ':' after field name")
		p.skipNewlines()
		value := p.parseExpression(1)
		lit.Fields = append(lit.Fields, ast.FieldValue{
			Name:  fieldName.Lexeme,
			Value: value,
			Pos:   fieldName.Pos,
		})
		p.consumeStatementEnd()
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after struct literal")
	return lit
}

func (p *Parser) parseUnary() ast.Expr {
	if p.match(lexer.Minus, lexer.Bang) {
		op := p.previous()
		return &ast.UnaryExpr{Op: op.Kind, Expr: p.parseUnary(), Pos: op.Pos}
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() ast.Expr {
	tok := p.peek()
	switch tok.Kind {
	case lexer.Int:
		p.advance()
		value, err := strconv.Atoi(tok.Lexeme)
		if err != nil {
			p.errorAt(tok, "invalid integer literal")
		}
		return &ast.IntegerLiteral{Value: value, Pos: tok.Pos}
	case lexer.String:
		p.advance()
		value, err := strconv.Unquote(tok.Lexeme)
		if err != nil {
			p.errorAt(tok, "invalid string literal")
			value = tok.Lexeme
		}
		return &ast.StringLiteral{Value: value, Pos: tok.Pos}
	case lexer.Ident:
		p.advance()
		if tok.Lexeme == "true" {
			return &ast.BoolLiteral{Value: true, Pos: tok.Pos}
		}
		if tok.Lexeme == "false" {
			return &ast.BoolLiteral{Value: false, Pos: tok.Pos}
		}
		return &ast.Identifier{Name: tok.Lexeme, Pos: tok.Pos}
	case lexer.At:
		at := p.advance()
		name := p.consume(lexer.Ident, "expected module name after '@'")
		return &ast.AtExpr{Name: name.Lexeme, Pos: at.Pos}
	case lexer.Dot:
		dot := p.advance()
		name := p.consume(lexer.Ident, "expected field name after '.'")
		return &ast.SelectorExpr{
			Receiver: &ast.ThisExpr{Pos: dot.Pos},
			Name:     name.Lexeme,
			Pos:      dot.Pos,
		}
	case lexer.LParen:
		p.advance()
		p.skipNewlines()
		expr := p.parseExpression(1)
		p.skipNewlines()
		p.consume(lexer.RParen, "expected ')' after expression")
		return expr
	default:
		p.errorAt(tok, fmt.Sprintf("expected expression, got %s", tok.Kind))
		p.advance()
		return &ast.Identifier{Name: "<error>", Pos: tok.Pos}
	}
}

func precedence(kind lexer.Kind) int {
	switch kind {
	case lexer.EqualEqual, lexer.BangEqual:
		return 1
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return 2
	case lexer.Plus, lexer.Minus:
		return 3
	case lexer.Star, lexer.Slash, lexer.Percent:
		return 4
	default:
		return 0
	}
}

func (p *Parser) looksLikePatternBranch() bool {
	i := p.curr
	for i < len(p.tokens) && p.tokens[i].Kind == lexer.Newline {
		i++
	}
	if i >= len(p.tokens) {
		return false
	}
	switch p.tokens[i].Kind {
	case lexer.Underscore, lexer.Int, lexer.String:
		i++
	case lexer.Ident:
		if p.tokens[i].Lexeme != "true" && p.tokens[i].Lexeme != "false" {
			return false
		}
		i++
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		i++
		if i >= len(p.tokens) {
			return false
		}
		if p.tokens[i].Kind != lexer.Int && p.tokens[i].Kind != lexer.String && p.tokens[i].Kind != lexer.Ident {
			return false
		}
		i++
	case lexer.LParen:
		depth := 1
		i++
		for i < len(p.tokens) && depth > 0 {
			switch p.tokens[i].Kind {
			case lexer.LParen:
				depth++
			case lexer.RParen:
				depth--
			}
			i++
		}
	default:
		return false
	}
	for i < len(p.tokens) && p.tokens[i].Kind == lexer.Newline {
		i++
	}
	return i < len(p.tokens) && p.tokens[i].Kind == lexer.FatArrow
}

func (p *Parser) consumeStatementEnd() {
	if p.match(lexer.Newline) {
		p.skipNewlines()
	}
}

func (p *Parser) skipNewlines() {
	for p.match(lexer.Newline) {
	}
}

func (p *Parser) consume(kind lexer.Kind, message string) lexer.Token {
	if p.check(kind) {
		return p.advance()
	}
	tok := p.peek()
	p.errorAt(tok, message)
	return tok
}

func (p *Parser) match(kinds ...lexer.Kind) bool {
	for _, kind := range kinds {
		if p.check(kind) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(kind lexer.Kind) bool {
	if p.curr >= len(p.tokens) {
		return kind == lexer.EOF
	}
	return p.peek().Kind == kind
}

func (p *Parser) checkNext(kind lexer.Kind) bool {
	if p.curr+1 >= len(p.tokens) {
		return kind == lexer.EOF
	}
	return p.tokens[p.curr+1].Kind == kind
}

func (p *Parser) advance() lexer.Token {
	if !p.check(lexer.EOF) {
		p.curr++
	}
	return p.previous()
}

func (p *Parser) peek() lexer.Token {
	if p.curr >= len(p.tokens) {
		return lexer.Token{Kind: lexer.EOF}
	}
	return p.tokens[p.curr]
}

func (p *Parser) previous() lexer.Token {
	if p.curr == 0 {
		return lexer.Token{Kind: lexer.EOF}
	}
	return p.tokens[p.curr-1]
}

func (p *Parser) errorAt(tok lexer.Token, message string) {
	p.errors = append(p.errors, Error{Message: message, Pos: tok.Pos})
}
