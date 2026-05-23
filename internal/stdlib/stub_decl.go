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
		p.consume(lexer.LParen, "expected '(' after annotation name")
		value := p.consume(lexer.String, "expected annotation string")
		p.consume(lexer.RParen, "expected ')' after annotation")
		if p.hasErrorToken(name) || p.hasErrorToken(value) {
			return nil, p.errorf(p.peek(), "invalid annotation")
		}
		annotations = append(annotations, annotation{Name: name.Lexeme, Value: unquote(value.Lexeme)})
		p.skipNewlines()
	}
	return annotations, nil
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

func (p *stubParser) parseReceiverBlock() ([]Function, error) {
	receiver := p.consume(lexer.Ident, "expected receiver name").Lexeme
	p.parseGenericNames()
	p.skipNewlines()
	p.consume(lexer.Colon, "expected ':' after receiver declaration")
	p.skipNewlines()
	p.consume(lexer.LBrace, "expected '{' after receiver declaration")
	p.skipNewlines()

	var functions []Function
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		annotations, err := p.parseAnnotations()
		if err != nil {
			return nil, err
		}
		fn, err := p.parseFunction(receiver, annotations)
		if err != nil {
			return nil, err
		}
		functions = append(functions, fn)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after receiver block")
	return functions, nil
}

func (p *stubParser) parseFunction(receiver string, annotations []annotation) (Function, error) {
	name := p.consume(lexer.Ident, "expected function name")
	generics := p.parseGenericNames()
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
	body, err := p.parseBodyExpr()
	if err != nil {
		return Function{}, err
	}

	fn := Function{
		Name:       name.Lexeme,
		SourcePath: p.path,
		Pos:        name.Pos,
		Receiver:   receiver,
		Generics:   generics,
		ParamNames: paramNames,
		Params:     params,
		Return:     returnType,
		Body:       body,
	}
	for _, ann := range annotations {
		if ann.Name == "alias" {
			fn.Alias = ann.Value
		}
	}
	if lit, ok := body.(*ast.StringLiteral); ok {
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
	}
	p.consume(lexer.RBracket, "expected ']' after generic parameters")
	return names
}

func (p *stubParser) parseTypeName() (string, error) {
	if p.match(lexer.LParen) {
		var args []string
		for !p.check(lexer.RParen) && !p.check(lexer.EOF) {
			arg, err := p.parseFunctionTypeParam()
			if err != nil {
				return "", err
			}
			args = append(args, arg)
			if !p.match(lexer.Comma) {
				break
			}
		}
		p.consume(lexer.RParen, "expected ')' after function parameter types")
		p.skipNewlines()
		p.consume(lexer.Arrow, "expected '->' after function parameter types")
		p.skipNewlines()
		ret, err := p.parseTypeName()
		if err != nil {
			return "", err
		}
		return "Func[" + strings.Join(append(args, ret), ",") + "]", nil
	}
	name := p.consume(lexer.Ident, "expected type name")
	if p.hasErrorToken(name) {
		return "", p.errorf(name, "expected type name")
	}
	typ := name.Lexeme
	if p.match(lexer.LBracket) {
		var args []string
		for !p.check(lexer.RBracket) && !p.check(lexer.EOF) {
			arg, err := p.parseTypeName()
			if err != nil {
				return "", err
			}
			args = append(args, arg)
			if !p.match(lexer.Comma) {
				break
			}
		}
		p.consume(lexer.RBracket, "expected ']' after type arguments")
		typ += "[" + strings.Join(args, ",") + "]"
	}
	return typ, nil
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
