package stdlib

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/parser"
)

func (p *stubParser) parseBodyExpr() (ast.Expr, error) {
	start := p.curr
	depth := 0
	for !p.check(lexer.EOF) {
		tok := p.peek()
		if depth == 0 && (tok.Kind == lexer.Newline || tok.Kind == lexer.RBrace) {
			break
		}
		switch tok.Kind {
		case lexer.LParen, lexer.LBracket, lexer.LBrace:
			depth++
		case lexer.RParen, lexer.RBracket, lexer.RBrace:
			if depth > 0 {
				depth--
			}
		}
		p.advance()
	}
	bodySrc := tokensSource(p.tokens[start:p.curr])
	file, errs := parser.Parse("stub() => " + bodySrc)
	if len(errs) > 0 || len(file.Functions) != 1 {
		var messages []string
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
		return nil, fmt.Errorf("%s: invalid stub body %q: %s", p.path, bodySrc, strings.Join(messages, "; "))
	}
	return file.Functions[0].Body, nil
}

func tokensSource(tokens []lexer.Token) string {
	parts := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if tok.Kind == lexer.EOF {
			continue
		}
		parts = append(parts, tok.Lexeme)
	}
	return strings.Join(parts, " ")
}
