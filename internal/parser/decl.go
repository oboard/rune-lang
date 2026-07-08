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

func (p *Parser) looksLikeGoPackageImportDecl() bool {
	if !p.check(lexer.At) || !p.checkNext(lexer.String) {
		return false
	}
	value, err := strconv.Unquote(p.tokens[p.curr+1].Lexeme)
	if err != nil {
		return false
	}
	return strings.HasPrefix(value, "go:")
}

func (p *Parser) parseGoPackageImportDecl() *ast.GoImport {
	at := p.consume(lexer.At, "expected '@'")
	path := p.consume(lexer.String, "expected Go import specifier string")
	value, err := strconv.Unquote(path.Lexeme)
	if err != nil {
		p.errorAt(path, "invalid Go import specifier string")
	}
	goPath, ok := goPackageImportPath(value)
	if !ok {
		p.errorAt(path, "expected Go import specifier to start with \"go:\"")
	}
	return &ast.GoImport{Path: goPath, Pos: at.Pos}
}

func goPackageImportPath(spec string) (string, bool) {
	path := strings.TrimPrefix(spec, "go:")
	return path, path != spec && path != ""
}

func (p *Parser) parseImportDecl() *ast.Import {
	at := p.consume(lexer.At, "expected '@'")
	if p.check(lexer.Ident) {
		module := p.advance()
		return &ast.Import{Path: module.Lexeme, Module: true, Pos: at.Pos}
	}
	path := p.consume(lexer.String, "expected import path string or module name after '@'")
	value := ""
	if path.Kind == lexer.String {
		var err error
		value, err = strconv.Unquote(path.Lexeme)
		if err != nil {
			p.errorAt(path, "invalid import path string")
		}
	}
	return &ast.Import{Path: value, Pos: at.Pos}
}

func (p *Parser) parsePublicModifier() bool {
	if !p.match(lexer.Plus) {
		return false
	}
	p.skipNewlines()
	return true
}

func (p *Parser) parseObjectPrivateModifier() bool {
	if !p.match(lexer.Minus) {
		return false
	}
	p.skipNewlines()
	return true
}

func (p *Parser) parseTest() *ast.Test {
	start := p.consume(lexer.Question, "expected '?'")
	name := p.consume(lexer.String, "expected test name string after '?'")
	value, err := strconv.Unquote(name.Lexeme)
	if err != nil {
		p.errorAt(name, "invalid test name string")
	}
	p.skipNewlines()
	if !p.check(lexer.LBrace) {
		p.errorAt(p.peek(), "expected test body block")
		return &ast.Test{Name: value, Pos: start.Pos, NamePos: name.Pos}
	}
	body := p.parseBlock()
	return &ast.Test{Name: value, Body: body, Pos: start.Pos, NamePos: name.Pos}
}

func (p *Parser) looksLikeConstDecl() bool {
	return p.check(lexer.Ident) && p.peek().Lexeme == "const"
}

func (p *Parser) parseConstDecl(private bool) *ast.ConstDecl {
	start := p.consume(lexer.Ident, "expected 'const'")
	name := p.consume(lexer.Ident, "expected constant name")
	decl := &ast.ConstDecl{Name: name.Lexeme, Private: private, Pos: start.Pos, NamePos: name.Pos}
	if p.match(lexer.Colon) {
		p.skipNewlines()
		decl.Type = p.parseTypeName()
	}
	p.skipNewlines()
	p.consume(lexer.Assign, "expected '=' after constant name")
	p.skipNewlines()
	decl.Value = p.parseExpression(1)
	return decl
}

func (p *Parser) parseTypeDecl(private bool) (*ast.StructType, *ast.EnumType) {
	name := p.consume(lexer.Ident, "expected type name")
	if name.Kind == lexer.EOF {
		return nil, nil
	}
	generics, constraints := p.parseGenericParams()
	p.consume(lexer.Colon, "expected ':' after type name")
	p.skipNewlines()
	p.consume(lexer.LBrace, "expected '{' after type declaration")
	p.skipNewlines()
	if p.looksLikeEnumMember() {
		if len(generics) > 0 && p.looksLikeEnumValueMember() {
			p.errorAt(name, "enum declarations cannot have generic parameters")
		}
		enum := p.parseEnumBody(name, private, generics, constraints)
		return nil, enum
	}
	typ := &ast.StructType{Name: name.Lexeme, Private: private, Generics: generics, GenericConstraints: constraints, Pos: name.Pos, NamePos: name.Pos}
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		if p.looksLikeMacroFunctionDecl() {
			p.advance()
			method := p.parseFunctionWithReceiver(typ.Name, true)
			if method != nil {
				method.Macro = true
				typ.Methods = append(typ.Methods, method)
			}
			p.consumeStatementEnd()
			p.skipNewlines()
			continue
		}
		annotations := p.parseAnnotations()
		private := p.parseObjectPrivateModifier()
		static := false
		if p.looksLikeStaticFunctionDecl() {
			p.matchStaticMethodMarker()
			p.skipNewlines()
			static = true
		}
		if p.looksLikeFunctionDecl() {
			method := p.parseFunctionWithReceiver(typ.Name, private)
			if method != nil {
				method.Static = static
				method.Annotations = annotations
				typ.Methods = append(typ.Methods, method)
			}
			p.consumeStatementEnd()
			p.skipNewlines()
			continue
		}
		fieldName := p.consume(lexer.Ident, "expected field name")
		if static {
			p.errorAt(fieldName, "struct fields cannot be static")
		}
		p.consume(lexer.Colon, "expected ':' after field name")
		fieldType := p.parseTypeName()
		typ.Fields = append(typ.Fields, ast.Field{
			Name:        fieldName.Lexeme,
			Private:     private,
			Annotations: annotations,
			Type:        fieldType,
			Pos:         fieldName.Pos,
		})
		p.consumeStatementEnd()
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after type declaration")
	return typ, nil
}

func (p *Parser) looksLikeStaticFunctionDecl() bool {
	if !p.checkStaticMethodMarker() {
		return false
	}
	saved := p.curr
	defer func() { p.curr = saved }()
	p.matchStaticMethodMarker()
	p.skipNewlines()
	return p.looksLikeFunctionDecl()
}

func (p *Parser) checkStaticMethodMarker() bool {
	return p.check(lexer.DoubleColon) || (p.check(lexer.Ident) && p.peek().Lexeme == "static")
}

func (p *Parser) matchStaticMethodMarker() bool {
	if p.match(lexer.DoubleColon) {
		return true
	}
	if p.check(lexer.Ident) && p.peek().Lexeme == "static" {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) looksLikeStaticTraitMethodDecl() bool {
	if !p.checkStaticMethodMarker() {
		return false
	}
	saved := p.curr
	defer func() { p.curr = saved }()
	p.matchStaticMethodMarker()
	p.skipNewlines()
	if !p.match(lexer.Ident) {
		return false
	}
	p.parseGenericNames()
	return p.check(lexer.LParen)
}

func (p *Parser) parseStructType() *ast.StructType {
	typ, _ := p.parseTypeDecl(false)
	return typ
}

func (p *Parser) parseTraitDecl() *ast.TraitDecl {
	start := p.consume(lexer.BitAnd, "expected '&'")
	name := p.consume(lexer.Ident, "expected trait name after '&'")
	p.consume(lexer.Colon, "expected ':' after trait name")
	p.skipNewlines()
	p.consume(lexer.LBrace, "expected '{' after trait declaration")
	p.skipNewlines()
	trait := &ast.TraitDecl{Name: name.Lexeme, Pos: start.Pos, NamePos: name.Pos}
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		static := false
		if p.looksLikeStaticTraitMethodDecl() {
			p.matchStaticMethodMarker()
			p.skipNewlines()
			static = true
		}
		member := p.consume(lexer.Ident, "expected trait member name")
		if p.match(lexer.LParen) {
			method := &ast.Function{Name: member.Lexeme, Static: static, Pos: member.Pos, NamePos: member.Pos}
			p.skipNewlines()
			if !p.check(lexer.RParen) {
				for {
					param := p.consume(lexer.Ident, "expected trait method parameter")
					p.consume(lexer.Colon, "expected ':' after trait method parameter")
					method.Params = append(method.Params, ast.Param{
						Name: param.Lexeme,
						Type: p.parseTypeName(),
						Pos:  param.Pos,
					})
					p.skipNewlines()
					if !p.match(lexer.Comma) {
						break
					}
					p.skipNewlines()
				}
			}
			p.consume(lexer.RParen, "expected ')' after trait method parameters")
			p.skipNewlines()
			if p.match(lexer.Arrow) {
				p.skipNewlines()
				method.ReturnType = p.parseTypeName()
			}
			trait.Methods = append(trait.Methods, method)
		} else {
			if static {
				p.errorAt(member, "trait fields cannot be static")
			}
			p.consume(lexer.Colon, "expected ':' after trait field name")
			trait.Fields = append(trait.Fields, ast.Field{
				Name: member.Lexeme,
				Type: p.parseTypeName(),
				Pos:  member.Pos,
			})
		}
		p.consumeStatementEnd()
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after trait declaration")
	return trait
}

func (p *Parser) looksLikeEnumMember() bool {
	saved := p.curr
	defer func() { p.curr = saved }()
	p.skipAnnotationTokens()
	p.match(lexer.Plus)
	p.skipNewlines()
	if !p.match(lexer.Ident) {
		return false
	}
	if p.check(lexer.Newline) || p.check(lexer.Comma) || p.check(lexer.RBrace) || p.check(lexer.EOF) {
		return true
	}
	p.skipNewlines()
	if p.check(lexer.Assign) {
		return true
	}
	if p.check(lexer.Colon) {
		return false
	}
	if !p.match(lexer.LParen) {
		return false
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
	p.skipNewlines()
	return !p.check(lexer.Arrow) && !p.check(lexer.FatArrow)
}

func (p *Parser) looksLikeEnumValueMember() bool {
	saved := p.curr
	defer func() { p.curr = saved }()
	p.match(lexer.Plus)
	p.skipNewlines()
	if !p.match(lexer.Ident) {
		return false
	}
	p.skipNewlines()
	return p.check(lexer.Assign)
}

func (p *Parser) parseEnumBody(name lexer.Token, private bool, generics []string, constraints map[string]ast.Type) *ast.EnumType {
	enum := &ast.EnumType{Name: name.Lexeme, Private: private, Generics: generics, GenericConstraints: constraints, Pos: name.Pos, NamePos: name.Pos}
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		annotations := p.parseAnnotations()
		privateMethod := p.parseObjectPrivateModifier()
		static := false
		if p.looksLikeStaticFunctionDecl() {
			p.matchStaticMethodMarker()
			p.skipNewlines()
			static = true
		}
		if p.looksLikeFunctionDecl() {
			method := p.parseFunctionWithReceiver(enum.Name, privateMethod)
			if method != nil {
				method.Static = static
				method.Annotations = annotations
				enum.Methods = append(enum.Methods, method)
			}
			p.consumeStatementEnd()
			p.skipNewlines()
			continue
		}
		if static {
			p.errorAt(p.peek(), "enum members cannot be static")
		}
		if privateMethod {
			p.errorAt(p.peek(), "enum members use '+' for public visibility")
		}
		public := p.parsePublicModifier()
		memberName := p.consume(lexer.Ident, "expected enum member name")
		_ = public
		member := ast.EnumMember{Name: memberName.Lexeme, Annotations: annotations, Pos: memberName.Pos}
		if p.check(lexer.Newline) || p.check(lexer.Comma) || p.check(lexer.RBrace) || p.check(lexer.EOF) {
			enum.Members = append(enum.Members, member)
			p.consumeStatementEnd()
			p.match(lexer.Comma)
			p.skipNewlines()
			continue
		}
		p.skipNewlines()
		if p.match(lexer.Assign) {
			p.skipNewlines()
			value, _ := p.parseEnumValue()
			member.Value = value
			member.HasValue = true
		} else if p.match(lexer.LParen) {
			p.skipNewlines()
			member.Params = p.parseEnumConstructorParams()
			p.consume(lexer.RParen, "expected ')' after enum constructor parameters")
		} else {
			p.errorAt(memberName, "expected enum member separator, '=' or '(' after enum member name")
		}
		enum.Members = append(enum.Members, member)
		p.consumeStatementEnd()
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after enum declaration")
	return enum
}

func (p *Parser) parseEnumConstructorParams() []ast.Param {
	var params []ast.Param
	if p.check(lexer.RParen) {
		return params
	}
	for !p.check(lexer.RParen) && !p.check(lexer.EOF) {
		name := p.consume(lexer.Ident, "expected enum constructor parameter name")
		p.consume(lexer.Colon, "expected ':' after enum constructor parameter name")
		typ := p.parseTypeName()
		params = append(params, ast.Param{Name: name.Lexeme, Type: typ, Pos: name.Pos})
		p.skipNewlines()
		if !p.match(lexer.Comma) {
			break
		}
		p.skipNewlines()
	}
	return params
}

func (p *Parser) parseEnumValue() (int, lexer.Position) {
	sign := 1
	pos := p.peek().Pos
	if p.match(lexer.Minus) {
		sign = -1
		pos = p.previous().Pos
	}
	tok := p.consume(lexer.Int, "expected integer enum value")
	if tok.Kind != lexer.Int {
		return 0, pos
	}
	value, err := strconv.Atoi(tok.Lexeme)
	if err != nil {
		p.errorAt(tok, "invalid enum value")
		return 0, pos
	}
	return sign * value, pos
}

func (p *Parser) parseFunction(private bool) *ast.Function {
	return p.parseFunctionWithReceiver("", private)
}

func (p *Parser) parseFunctionWithReceiver(receiverType string, private bool) *ast.Function {
	routine := false
	if p.match(lexer.Tilde) {
		routine = true
		p.skipNewlines()
	}
	name := p.consume(lexer.Ident, "expected function name")
	if name.Kind == lexer.EOF {
		return nil
	}

	fn := &ast.Function{Name: name.Lexeme, Private: private, Routine: routine, ReceiverType: receiverType, Pos: name.Pos, NamePos: name.Pos}
	fn.Generics, fn.GenericConstraints = p.parseGenericParams()
	p.consume(lexer.LParen, "expected '(' after function name")
	p.skipNewlines()
	if !p.check(lexer.RParen) {
		for {
			paramName := p.consume(lexer.Ident, "expected parameter name")
			paramType := ast.Type{}
			if p.match(lexer.Colon) {
				paramType = p.parseTypeName()
			}
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
		ret := p.parseTypeName()
		fn.ReturnType = ret
		p.skipNewlines()
	}
	p.consume(lexer.FatArrow, "expected '=>' after function signature")
	p.skipNewlines()
	fn.Body = p.parseBody()
	return fn
}

func (p *Parser) parseGenericNames() []string {
	names, _ := p.parseGenericParams()
	return names
}

func (p *Parser) parseGenericParams() ([]string, map[string]ast.Type) {
	if !p.match(lexer.LBracket) {
		return nil, nil
	}
	var names []string
	constraints := map[string]ast.Type{}
	for !p.check(lexer.RBracket) && !p.check(lexer.EOF) {
		if p.check(lexer.Ident) {
			name := p.advance()
			names = append(names, name.Lexeme)
			if p.match(lexer.Colon) {
				p.skipNewlines()
				constraints[name.Lexeme] = p.parseTypeName()
			}
		} else {
			p.advance()
		}
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBracket, "expected ']' after generic parameters")
	if len(constraints) == 0 {
		constraints = nil
	}
	return names, constraints
}

func (p *Parser) parseTypeName() ast.Type {
	p.skipNewlines()
	if p.match(lexer.LParen) {
		var params []ast.TypeParam
		for !p.check(lexer.RParen) && !p.check(lexer.EOF) {
			arg := p.parseFunctionTypeParam()
			params = append(params, arg)
			p.skipNewlines()
			if !p.match(lexer.Comma) {
				break
			}
			p.skipNewlines()
		}
		p.consume(lexer.RParen, "expected ')' after function parameter types")
		p.skipNewlines()
		if p.match(lexer.Arrow) {
			p.skipNewlines()
			ret := p.parseTypeName()
			return ast.FunctionType(params, ret)
		}
		if len(params) == 1 && params[0].Name == "" && !params[0].Optional {
			return ast.GroupedType(params[0].Type)
		}
		return ast.TupleType(params)
	}
	typ := p.parseSimpleTypeName()
	if p.match(lexer.LBracket) {
		var args []ast.Type
		for !p.check(lexer.RBracket) && !p.check(lexer.EOF) {
			arg := p.parseTypeName()
			args = append(args, arg)
			p.skipNewlines()
			if !p.match(lexer.Comma) {
				break
			}
			p.skipNewlines()
		}
		p.consume(lexer.RBracket, "expected ']' after type arguments")
		typ = typ.WithArgs(args)
	}
	if p.match(lexer.Question) {
		typ = typ.WithNullable()
	}
	return typ
}

func (p *Parser) parseSimpleTypeName() ast.Type {
	if p.match(lexer.BitAnd) {
		name := p.consume(lexer.Ident, "expected trait name after '&'")
		return ast.NamedType("&" + name.Lexeme)
	}
	if p.match(lexer.At) {
		module := p.consume(lexer.Ident, "expected module name after '@'")
		p.consume(lexer.Dot, "expected '.' after module name")
		name := p.consume(lexer.Ident, "expected type name after module qualifier")
		return ast.QualifiedType(module.Lexeme, name.Lexeme)
	}
	name := p.consume(lexer.Ident, "expected type name")
	return ast.NamedType(name.Lexeme)
}

func (p *Parser) parseFunctionTypeParam() ast.TypeParam {
	p.skipNewlines()
	name := ""
	optional := false
	if p.check(lexer.Ident) && p.typeParamHasName() {
		name = p.advance().Lexeme
		optional = p.match(lexer.Question)
		p.consume(lexer.Colon, "expected ':' after function type parameter name")
		p.skipNewlines()
	}
	typ := p.parseTypeName()
	if name == "" {
		return ast.TypeParam{Type: typ}
	}
	return ast.TypeParam{Name: name, Optional: optional, Type: typ}
}

func (p *Parser) typeParamHasName() bool {
	if p.curr+1 >= len(p.tokens) {
		return false
	}
	switch p.tokens[p.curr+1].Kind {
	case lexer.Colon:
		return true
	case lexer.Question:
		return p.curr+2 < len(p.tokens) && p.tokens[p.curr+2].Kind == lexer.Colon
	default:
		return false
	}
}

func (p *Parser) parseAnnotations() []ast.Annotation {
	var annotations []ast.Annotation
	for p.match(lexer.Hash) || p.match(lexer.At) {
		start := p.previous()
		first := p.consume(lexer.Ident, "expected annotation name")
		annotation := ast.Annotation{Name: first.Lexeme, Pos: start.Pos, NamePos: first.Pos}
		if p.match(lexer.Dot) {
			name := p.consume(lexer.Ident, "expected annotation function name after '.'")
			annotation.Module = first.Lexeme
			annotation.Name = name.Lexeme
			annotation.NamePos = name.Pos
		}
		if p.match(lexer.LParen) {
			annotation.HasParens = true
			p.skipNewlines()
			if !p.check(lexer.RParen) {
				for {
					annotation.Args = append(annotation.Args, p.parseExpression(1))
					p.skipNewlines()
					if !p.match(lexer.Comma) {
						break
					}
					p.skipNewlines()
				}
			}
			p.consume(lexer.RParen, "expected ')' after annotation arguments")
		}
		annotations = append(annotations, annotation)
		p.skipNewlines()
	}
	return annotations
}

func (p *Parser) skipAnnotationTokens() {
	for p.match(lexer.Hash) || p.match(lexer.At) {
		if !p.match(lexer.Ident) {
			return
		}
		if p.match(lexer.Dot) {
			if !p.match(lexer.Ident) {
				return
			}
		}
		if p.match(lexer.LParen) {
			depth := 1
			for !p.check(lexer.EOF) && depth > 0 {
				switch p.advance().Kind {
				case lexer.LParen:
					depth++
				case lexer.RParen:
					depth--
				}
			}
		}
		p.skipNewlines()
	}
}

func (p *Parser) parseBody() ast.Expr {
	if p.check(lexer.LBrace) {
		if !p.looksLikePatternBranch() && p.looksLikeObjectLiteralBody() {
			return p.parseAnonymousObjectLiteral()
		}
		if !p.looksLikePatternBranch() && p.looksLikeMapLiteralBody() {
			return p.parseMapLiteral()
		}
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
			expr := p.parseBody()
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
