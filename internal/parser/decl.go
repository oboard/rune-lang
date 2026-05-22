package parser

import (
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

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
	typ := &ast.StructType{Name: name.Lexeme, Generics: p.parseGenericNames(), Pos: name.Pos, NamePos: name.Pos}
	p.consume(lexer.Colon, "expected ':' after type name")
	p.skipNewlines()
	p.consume(lexer.LBrace, "expected '{' after type declaration")
	p.skipNewlines()
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		annotations := p.parseAnnotations()
		if p.looksLikeFunctionDecl() {
			method := p.parseFunctionWithReceiver(typ.Name)
			if method != nil {
				method.Annotations = annotations
				typ.Methods = append(typ.Methods, method)
			}
			p.consumeStatementEnd()
			p.skipNewlines()
			continue
		}
		fieldName := p.consume(lexer.Ident, "expected field name")
		p.consume(lexer.Colon, "expected ':' after field name")
		fieldType := p.parseTypeName()
		typ.Fields = append(typ.Fields, ast.Field{
			Name: fieldName.Lexeme,
			Type: fieldType,
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
	fn.Generics = p.parseGenericNames()
	p.consume(lexer.LParen, "expected '(' after function name")
	p.skipNewlines()
	if !p.check(lexer.RParen) {
		for {
			paramName := p.consume(lexer.Ident, "expected parameter name")
			p.consume(lexer.Colon, "expected ':' after parameter name")
			paramType := p.parseTypeName()
			fn.Params = append(fn.Params, ast.Param{
				Name: paramName.Lexeme,
				Type: paramType,
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
		fn.ReturnType = p.parseTypeName()
		p.skipNewlines()
	}
	p.consume(lexer.FatArrow, "expected '=>' after function signature")
	p.skipNewlines()
	fn.Body = p.parseBody()
	return fn
}

func (p *Parser) parseGenericNames() []string {
	if !p.match(lexer.LBracket) {
		return nil
	}
	var names []string
	for !p.check(lexer.RBracket) && !p.check(lexer.EOF) {
		if p.check(lexer.Ident) {
			names = append(names, p.advance().Lexeme)
		} else {
			p.advance()
		}
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBracket, "expected ']' after generic parameters")
	return names
}

func (p *Parser) parseTypeName() string {
	p.skipNewlines()
	if p.match(lexer.LParen) {
		var args []string
		for !p.check(lexer.RParen) && !p.check(lexer.EOF) {
			args = append(args, p.parseTypeName())
			p.skipNewlines()
			if !p.match(lexer.Comma) {
				break
			}
			p.skipNewlines()
		}
		p.consume(lexer.RParen, "expected ')' after function parameter types")
		p.skipNewlines()
		p.consume(lexer.Arrow, "expected '->' after function parameter types")
		p.skipNewlines()
		ret := p.parseTypeName()
		return "Func[" + strings.Join(append(args, ret), ",") + "]"
	}
	name := p.consume(lexer.Ident, "expected type name")
	typ := name.Lexeme
	if p.match(lexer.LBracket) {
		var args []string
		for !p.check(lexer.RBracket) && !p.check(lexer.EOF) {
			args = append(args, p.parseTypeName())
			p.skipNewlines()
			if !p.match(lexer.Comma) {
				break
			}
			p.skipNewlines()
		}
		p.consume(lexer.RBracket, "expected ']' after type arguments")
		typ += "[" + strings.Join(args, ",") + "]"
	}
	return typ
}

func (p *Parser) parseAnnotations() []ast.Annotation {
	var annotations []ast.Annotation
	for p.match(lexer.At) {
		at := p.previous()
		name := p.consume(lexer.Ident, "expected annotation name")
		annotation := ast.Annotation{Name: name.Lexeme, Pos: at.Pos}
		if p.match(lexer.LParen) {
			if p.check(lexer.String) {
				value := p.advance()
				unquoted, err := strconv.Unquote(value.Lexeme)
				if err != nil {
					p.errorAt(value, "invalid annotation string")
				} else {
					annotation.Value = unquoted
				}
			}
			depth := 1
			for !p.check(lexer.EOF) && depth > 0 {
				tok := p.advance()
				switch tok.Kind {
				case lexer.LParen:
					depth++
				case lexer.RParen:
					depth--
				}
			}
		}
		annotations = append(annotations, annotation)
		p.skipNewlines()
	}
	return annotations
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
