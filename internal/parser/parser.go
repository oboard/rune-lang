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
