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

func (p *Parser) parseImportDecl() *ast.Import {
	at := p.consume(lexer.At, "expected '@'")
	path := p.consume(lexer.String, "expected import path string after '@'")
	value, err := strconv.Unquote(path.Lexeme)
	if err != nil {
		p.errorAt(path, "invalid import path string")
	}
	return &ast.Import{Path: value, Pos: at.Pos}
}

func (p *Parser) parsePrivateModifier() bool {
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

func (p *Parser) parseTypeDecl(private bool) (*ast.StructType, *ast.EnumType) {
	name := p.consume(lexer.Ident, "expected type name")
	if name.Kind == lexer.EOF {
		return nil, nil
	}
	generics := p.parseGenericNames()
	p.consume(lexer.Colon, "expected ':' after type name")
	p.skipNewlines()
	p.consume(lexer.LBrace, "expected '{' after type declaration")
	p.skipNewlines()
	if p.looksLikeEnumMember() {
		if len(generics) > 0 && p.looksLikeEnumValueMember() {
			p.errorAt(name, "enum declarations cannot have generic parameters")
		}
		enum := p.parseEnumBody(name, private, generics)
		return nil, enum
	}
	typ := &ast.StructType{Name: name.Lexeme, Private: private, Generics: generics, Pos: name.Pos, NamePos: name.Pos}
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		annotations := p.parseAnnotations()
		private := p.parsePrivateModifier()
		if p.looksLikeFunctionDecl() {
			method := p.parseFunctionWithReceiver(typ.Name, private)
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
			Name:        fieldName.Lexeme,
			Private:     private,
			Type:        fieldType.canonical,
			TypeDisplay: fieldType.display,
			Pos:         fieldName.Pos,
		})
		p.consumeStatementEnd()
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after type declaration")
	return typ, nil
}

func (p *Parser) parseStructType() *ast.StructType {
	typ, _ := p.parseTypeDecl(false)
	return typ
}

func (p *Parser) looksLikeEnumMember() bool {
	saved := p.curr
	defer func() { p.curr = saved }()
	p.match(lexer.Minus)
	p.skipNewlines()
	if !p.match(lexer.Ident) {
		return false
	}
	p.skipNewlines()
	if p.check(lexer.Assign) {
		return true
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
	p.match(lexer.Minus)
	p.skipNewlines()
	if !p.match(lexer.Ident) {
		return false
	}
	p.skipNewlines()
	return p.check(lexer.Assign)
}

func (p *Parser) parseEnumBody(name lexer.Token, private bool, generics []string) *ast.EnumType {
	enum := &ast.EnumType{Name: name.Lexeme, Private: private, Generics: generics, Pos: name.Pos, NamePos: name.Pos}
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		memberPrivate := p.parsePrivateModifier()
		memberName := p.consume(lexer.Ident, "expected enum member name")
		p.skipNewlines()
		member := ast.EnumMember{Name: memberName.Lexeme, Private: memberPrivate, Pos: memberName.Pos}
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
			p.errorAt(memberName, "expected '=' or '(' after enum member name")
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
		params = append(params, ast.Param{Name: name.Lexeme, Type: typ.canonical, TypeDisplay: typ.display, Pos: name.Pos})
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
	fn.Generics = p.parseGenericNames()
	p.consume(lexer.LParen, "expected '(' after function name")
	p.skipNewlines()
	if !p.check(lexer.RParen) {
		for {
			paramName := p.consume(lexer.Ident, "expected parameter name")
			paramType := parsedType{}
			if p.match(lexer.Colon) {
				paramType = p.parseTypeName()
			}
			fn.Params = append(fn.Params, ast.Param{
				Name:        paramName.Lexeme,
				Type:        paramType.canonical,
				TypeDisplay: paramType.display,
				Pos:         paramName.Pos,
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
		fn.ReturnType = ret.canonical
		fn.ReturnDisplay = ret.display
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

type parsedType struct {
	canonical string
	display   string
}

func (p *Parser) parseTypeName() parsedType {
	p.skipNewlines()
	if p.match(lexer.LParen) {
		var canonicalArgs []string
		var displayArgs []string
		for !p.check(lexer.RParen) && !p.check(lexer.EOF) {
			arg := p.parseFunctionTypeParam()
			canonicalArgs = append(canonicalArgs, arg.canonical)
			displayArgs = append(displayArgs, arg.display)
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
			return parsedType{
				canonical: "Func[" + strings.Join(append(canonicalArgs, ret.canonical), ",") + "]",
				display:   "(" + strings.Join(displayArgs, ", ") + ") -> " + ret.display,
			}
		}
		if len(canonicalArgs) == 1 {
			return parsedType{canonical: canonicalArgs[0], display: "(" + displayArgs[0] + ")"}
		}
		return parsedType{
			canonical: "Tuple[" + strings.Join(canonicalArgs, ",") + "]",
			display:   "(" + strings.Join(displayArgs, ", ") + ")",
		}
	}
	typ, display := p.parseSimpleTypeName()
	if p.match(lexer.LBracket) {
		var canonicalArgs []string
		var displayArgs []string
		for !p.check(lexer.RBracket) && !p.check(lexer.EOF) {
			arg := p.parseTypeName()
			canonicalArgs = append(canonicalArgs, arg.canonical)
			displayArgs = append(displayArgs, arg.display)
			p.skipNewlines()
			if !p.match(lexer.Comma) {
				break
			}
			p.skipNewlines()
		}
		p.consume(lexer.RBracket, "expected ']' after type arguments")
		typ += "[" + strings.Join(canonicalArgs, ",") + "]"
		display += "[" + strings.Join(displayArgs, ", ") + "]"
	}
	if p.match(lexer.Question) {
		typ += "?"
		display += "?"
	}
	return parsedType{canonical: typ, display: display}
}

func (p *Parser) parseSimpleTypeName() (string, string) {
	if p.match(lexer.At) {
		module := p.consume(lexer.Ident, "expected module name after '@'")
		p.consume(lexer.Dot, "expected '.' after module name")
		name := p.consume(lexer.Ident, "expected type name after module qualifier")
		display := "@" + module.Lexeme + "." + name.Lexeme
		return name.Lexeme, display
	}
	name := p.consume(lexer.Ident, "expected type name")
	return name.Lexeme, name.Lexeme
}

func (p *Parser) parseFunctionTypeParam() parsedType {
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
		return typ
	}
	suffix := ": "
	if optional {
		suffix = "?: "
	}
	typ.display = name + suffix + typ.display
	return typ
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
