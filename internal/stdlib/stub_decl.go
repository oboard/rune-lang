package stdlib

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *stubParser) parseAnnotations() ([]annotation, error) {
	var annotations []annotation
	for p.match(lexer.At) {
		name := p.consume(lexer.Ident, "expected annotation name")
		ann := annotation{Name: name.Lexeme}
		if p.match(lexer.LParen) {
			ann.HasParens = true
			value := p.consume(lexer.String, "expected annotation string")
			p.consume(lexer.RParen, "expected ')' after annotation")
			ann.Value = unquote(value.Lexeme)
			if p.hasErrorToken(value) {
				return nil, p.errorf(p.peek(), "invalid annotation")
			}
		}
		if p.hasErrorToken(name) {
			return nil, p.errorf(p.peek(), "invalid annotation")
		}
		annotations = append(annotations, ann)
		p.skipNewlines()
	}
	return annotations, nil
}

func (p *stubParser) parseReceiverAnnotations() ([]annotation, error) {
	annotations, err := p.parseAnnotations()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.Hash) {
		name := p.consume(lexer.Ident, "expected annotation name")
		ann := annotation{Name: name.Lexeme}
		if p.match(lexer.LParen) {
			ann.HasParens = true
			value := p.consume(lexer.String, "expected annotation string")
			p.consume(lexer.RParen, "expected ')' after annotation")
			ann.Value = unquote(value.Lexeme)
			if p.hasErrorToken(value) {
				return nil, p.errorf(p.peek(), "invalid annotation")
			}
		}
		if p.hasErrorToken(name) {
			return nil, p.errorf(p.peek(), "invalid annotation")
		}
		annotations = append(annotations, ann)
		p.skipNewlines()
	}
	return annotations, nil
}

func (p *stubParser) parseTrait() (Trait, error) {
	start := p.consume(lexer.BitAnd, "expected '&'")
	name := p.consume(lexer.Ident, "expected trait name")
	p.consume(lexer.Colon, "expected ':' after trait name")
	p.skipNewlines()
	p.consume(lexer.LBrace, "expected '{' after trait name")
	p.skipNewlines()

	trait := Trait{Name: name.Lexeme, SourcePath: p.path, Pos: start.Pos}
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		static := false
		if p.matchStaticMethodMarker() {
			p.skipNewlines()
			static = true
		}
		member := p.consume(lexer.Ident, "expected trait member")
		if p.match(lexer.LParen) {
			paramNames, params, err := p.parseParams()
			if err != nil {
				return Trait{}, err
			}
			p.consume(lexer.RParen, "expected ')' after trait method parameters")
			ret := "Void"
			p.skipNewlines()
			if p.match(lexer.Arrow) {
				ret, err = p.parseTypeName()
				if err != nil {
					return Trait{}, err
				}
			}
			trait.Methods = append(trait.Methods, Function{
				Name:       member.Lexeme,
				Static:     static,
				SourcePath: p.path,
				Pos:        member.Pos,
				ParamNames: paramNames,
				Params:     params,
				Return:     ret,
			})
		} else {
			if static {
				return Trait{}, p.errorf(member, "trait fields cannot be static")
			}
			p.consume(lexer.Colon, "expected ':' after trait field")
			typ, err := p.parseTypeName()
			if err != nil {
				return Trait{}, err
			}
			trait.Fields = append(trait.Fields, Field{Name: member.Lexeme, Type: typ, Pos: member.Pos})
		}
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after trait")
	return trait, nil
}

func (p *stubParser) looksLikeReceiverBlock() bool {
	saved := p.curr
	defer func() { p.curr = saved }()
	if !p.match(lexer.Ident) {
		return false
	}
	p.parseGenericNames()
	p.skipNewlines()
	return p.check(lexer.Colon)
}

func (p *stubParser) parseReceiverBlock() (*Type, []Function, error) {
	name := p.consume(lexer.Ident, "expected receiver name")
	receiver := name.Lexeme
	generics, constraints := p.parseGenericParams()
	p.skipNewlines()
	p.consume(lexer.Colon, "expected ':' after receiver declaration")
	p.skipNewlines()
	p.consume(lexer.LBrace, "expected '{' after receiver declaration")
	p.skipNewlines()

	typ := &Type{Name: receiver, SourcePath: p.path, Pos: name.Pos, Generics: generics, GenericConstraints: constraints}
	var functions []Function
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		if p.match(lexer.Comma) {
			p.skipNewlines()
			continue
		}
		annotations, err := p.parseReceiverAnnotations()
		if err != nil {
			return nil, nil, err
		}
		p.match(lexer.Plus)
		if p.looksLikeFieldDecl() {
			field, err := p.parseField()
			if err != nil {
				return nil, nil, err
			}
			typ.Fields = append(typ.Fields, field)
			p.match(lexer.Comma)
			p.skipNewlines()
			continue
		}
		if !p.looksLikeFunctionDecl() {
			if !p.check(lexer.Ident) {
				return nil, nil, p.errorf(p.peek(), "expected receiver member")
			}
			constructor, err := p.parseConstructor()
			if err != nil {
				return nil, nil, err
			}
			typ.Constructors = append(typ.Constructors, constructor)
			p.match(lexer.Comma)
			p.skipNewlines()
			continue
		}
		fn, err := p.parseFunction(receiver, annotations)
		if err != nil {
			return nil, nil, err
		}
		functions = append(functions, fn)
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after receiver block")
	return typ, functions, nil
}

func (p *stubParser) looksLikeFieldDecl() bool {
	saved := p.curr
	defer func() { p.curr = saved }()
	if !p.match(lexer.Ident) {
		return false
	}
	return p.check(lexer.Colon)
}

func (p *stubParser) looksLikeFunctionDecl() bool {
	saved := p.curr
	defer func() { p.curr = saved }()
	p.match(lexer.Tilde)
	p.skipNewlines()
	if !p.match(lexer.Ident) {
		return false
	}
	p.parseGenericNames()
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
	if p.match(lexer.Arrow) {
		p.skipTypeNameTokens()
	}
	p.skipNewlines()
	return p.check(lexer.FatArrow)
}

func (p *stubParser) parseField() (Field, error) {
	name := p.consume(lexer.Ident, "expected field name")
	p.consume(lexer.Colon, "expected ':' after field name")
	typ, err := p.parseTypeName()
	if err != nil {
		return Field{}, err
	}
	return Field{Name: name.Lexeme, Type: typ, Pos: name.Pos}, nil
}

func (p *stubParser) parseConstructor() (Constructor, error) {
	name := p.consume(lexer.Ident, "expected constructor name")
	if !p.match(lexer.LParen) {
		return Constructor{Name: name.Lexeme, Pos: name.Pos}, nil
	}
	paramNames, params, err := p.parseParams()
	if err != nil {
		return Constructor{}, err
	}
	p.consume(lexer.RParen, "expected ')' after constructor parameters")
	return Constructor{Name: name.Lexeme, ParamNames: paramNames, Params: params, Pos: name.Pos}, nil
}

func (p *stubParser) parseFunction(receiver string, annotations []annotation) (Function, error) {
	routine := false
	if p.match(lexer.Tilde) {
		routine = true
		p.skipNewlines()
	}
	name := p.consume(lexer.Ident, "expected function name")
	generics, constraints := p.parseGenericParams()
	p.consume(lexer.LParen, "expected '(' after function name")
	paramNames, params, err := p.parseParams()
	if err != nil {
		return Function{}, err
	}
	p.consume(lexer.RParen, "expected ')' after parameter list")

	returnType := ""
	p.skipNewlines()
	if p.match(lexer.Arrow) {
		returnType, err = p.parseTypeName()
		if err != nil {
			return Function{}, err
		}
	}
	p.skipNewlines()
	p.consume(lexer.FatArrow, "expected '=>' after function signature")
	p.skipNewlines()
	body, err := p.parseBodyExpr()
	if err != nil {
		return Function{}, err
	}

	fn := Function{
		Name:               name.Lexeme,
		Routine:            routine,
		SourcePath:         p.path,
		Pos:                name.Pos,
		Receiver:           receiver,
		Generics:           generics,
		GenericConstraints: constraints,
		ParamNames:         paramNames,
		Params:             params,
		Return:             returnType,
		Body:               body,
	}
	for _, ann := range annotations {
		switch ann.Name {
		case "alias":
			if !ann.HasParens {
				return Function{}, fmt.Errorf("%s.%s @alias requires a string argument", p.moduleName, fn.Name)
			}
			fn.Alias = ann.Value
		default:
			return Function{}, fmt.Errorf("%s.%s unknown core annotation @%s", p.moduleName, fn.Name, ann.Name)
		}
	}
	if lit, ok := body.(*ast.StringLiteral); ok {
		if fn.Macro {
			return Function{}, fmt.Errorf("%s.%s macro body must be written in Rune", p.moduleName, fn.Name)
		}
		spec := lit.Value
		if !strings.HasPrefix(spec, "%") {
			return Function{}, fmt.Errorf("%s.%s intrinsic must start with %%", p.moduleName, fn.Name)
		}
		if err := applyIntrinsicSpec(p.moduleName, &fn, strings.TrimPrefix(spec, "%")); err != nil {
			return Function{}, err
		}
		fn.Body = nil
	}
	if fn.Return == "" {
		fn.Return = inferredReturn(fn)
	}
	return fn, nil
}

func (p *stubParser) parseParams() ([]string, []string, error) {
	var names []string
	var params []string
	p.skipNewlines()
	if p.check(lexer.RParen) {
		return names, params, nil
	}
	for {
		name := p.consume(lexer.Ident, "expected parameter name")
		p.consume(lexer.Colon, "expected ':' after parameter name")
		typ, err := p.parseTypeName()
		if err != nil {
			return nil, nil, err
		}
		names = append(names, name.Lexeme)
		params = append(params, typ)
		p.skipNewlines()
		if !p.match(lexer.Comma) {
			break
		}
		p.skipNewlines()
	}
	return names, params, nil
}

func (p *stubParser) parseGenericNames() []string {
	names, _ := p.parseGenericParams()
	return names
}

func (p *stubParser) parseGenericParams() ([]string, map[string]string) {
	if !p.match(lexer.LBracket) {
		return nil, nil
	}
	var names []string
	constraints := map[string]string{}
	for !p.check(lexer.RBracket) && !p.check(lexer.EOF) {
		if p.check(lexer.Ident) {
			name := p.advance()
			names = append(names, name.Lexeme)
			if p.match(lexer.Colon) {
				p.skipNewlines()
				constraint, err := p.parseTypeName()
				if err != nil {
					constraints[name.Lexeme] = ""
				} else {
					constraints[name.Lexeme] = strings.TrimPrefix(constraint, "&")
				}
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

func (p *stubParser) parseTypeName() (string, error) {
	p.skipNewlines()
	if p.match(lexer.LParen) {
		var args []string
		p.skipNewlines()
		for !p.check(lexer.RParen) && !p.check(lexer.EOF) {
			arg, err := p.parseFunctionTypeParam()
			if err != nil {
				return "", err
			}
			args = append(args, arg)
			p.skipNewlines()
			if !p.match(lexer.Comma) {
				break
			}
			p.skipNewlines()
		}
		p.skipNewlines()
		p.consume(lexer.RParen, "expected ')' after function parameter types")
		p.skipNewlines()
		if p.match(lexer.Arrow) {
			p.skipNewlines()
			ret, err := p.parseTypeName()
			if err != nil {
				return "", err
			}
			return "Func[" + strings.Join(append(args, ret), ",") + "]", nil
		}
		if len(args) == 1 {
			return args[0], nil
		}
		return "Tuple[" + strings.Join(args, ",") + "]", nil
	}
	typ, err := p.parseSimpleTypeName()
	if err != nil {
		return "", err
	}
	if p.match(lexer.LBracket) {
		var args []string
		p.skipNewlines()
		for !p.check(lexer.RBracket) && !p.check(lexer.EOF) {
			arg, err := p.parseTypeName()
			if err != nil {
				return "", err
			}
			args = append(args, arg)
			p.skipNewlines()
			if !p.match(lexer.Comma) {
				break
			}
			p.skipNewlines()
		}
		p.skipNewlines()
		p.consume(lexer.RBracket, "expected ']' after type arguments")
		typ += "[" + strings.Join(args, ",") + "]"
	}
	if p.match(lexer.Question) {
		typ += "?"
	}
	return typ, nil
}

func (p *stubParser) parseSimpleTypeName() (string, error) {
	if p.match(lexer.BitAnd) {
		name := p.consume(lexer.Ident, "expected trait name after '&'")
		if p.hasErrorToken(name) {
			return "", p.errorf(name, "expected trait name")
		}
		return "&" + name.Lexeme, nil
	}
	if p.match(lexer.At) {
		module := p.consume(lexer.Ident, "expected module name after '@'")
		p.consume(lexer.Dot, "expected '.' after module name")
		name := p.consume(lexer.Ident, "expected type name after module qualifier")
		if p.hasErrorToken(module) || p.hasErrorToken(name) {
			return "", p.errorf(module, "expected qualified type name")
		}
		return name.Lexeme, nil
	}
	name := p.consume(lexer.Ident, "expected type name")
	if p.hasErrorToken(name) {
		return "", p.errorf(name, "expected type name")
	}
	return name.Lexeme, nil
}

func (p *stubParser) parseFunctionTypeParam() (string, error) {
	p.skipNewlines()
	if p.check(lexer.Ident) && p.typeParamHasName() {
		p.advance()
		p.match(lexer.Question)
		p.consume(lexer.Colon, "expected ':' after function type parameter name")
		p.skipNewlines()
	}
	return p.parseTypeName()
}

func (p *stubParser) typeParamHasName() bool {
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
