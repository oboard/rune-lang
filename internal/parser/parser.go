package parser

import (
	"fmt"

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
		private := p.parsePrivateModifier()
		if private && (p.check(lexer.At) || p.check(lexer.Question)) {
			p.errorAt(p.peek(), "expected private declaration after '-'")
		}
		if !private && p.check(lexer.At) {
			if p.checkNext(lexer.String) {
				if imp := p.parseImportDecl(); imp != nil {
					file.Imports = append(file.Imports, *imp)
				}
			} else if imp := p.parseGoImportDecl(); imp != nil {
				file.GoImports = append(file.GoImports, *imp)
			}
		} else if !private && p.check(lexer.Question) {
			test := p.parseTest()
			if test != nil {
				file.Tests = append(file.Tests, test)
			}
		} else if p.looksLikeTypeDecl() {
			typ, enum := p.parseTypeDecl(private)
			if typ != nil {
				file.Types = append(file.Types, typ)
			}
			if enum != nil {
				file.Enums = append(file.Enums, enum)
			}
		} else if p.looksLikeFunctionDecl() {
			fn := p.parseFunction(private)
			if fn != nil {
				file.Functions = append(file.Functions, fn)
			}
		} else {
			p.errorAt(p.peek(), "expected declaration")
			p.advance()
		}
		p.skipNewlines()
	}
	return file, p.errors
}
