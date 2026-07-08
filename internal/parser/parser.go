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
		if p.looksLikeGoImportDecl() {
			if imp := p.parseGoImportDecl(); imp != nil {
				file.GoImports = append(file.GoImports, *imp)
			}
			p.skipNewlines()
			continue
		}
		if p.check(lexer.At) && (p.checkNext(lexer.String) || p.checkNext(lexer.Ident)) {
			if p.looksLikeGoPackageImportDecl() {
				if imp := p.parseGoPackageImportDecl(); imp != nil {
					file.GoImports = append(file.GoImports, *imp)
				}
				p.skipNewlines()
				continue
			}
			if imp := p.parseImportDecl(); imp != nil {
				file.Imports = append(file.Imports, *imp)
			}
			p.skipNewlines()
			continue
		}
		if p.looksLikeMacroFunctionDecl() {
			p.advance()
			fn := p.parseFunction(true)
			if fn != nil {
				fn.Macro = true
				file.Functions = append(file.Functions, fn)
			}
			p.skipNewlines()
			continue
		}
		annotations := p.parseAnnotations()
		public := p.parsePublicModifier()
		private := !public
		if public && p.check(lexer.Question) {
			p.errorAt(p.peek(), "expected public declaration after '+'")
		}
		if p.check(lexer.BitAnd) {
			if len(annotations) > 0 {
				p.errorAt(p.peek(), "annotations cannot be applied to traits")
			}
			if trait := p.parseTraitDecl(); trait != nil {
				file.Traits = append(file.Traits, trait)
			}
		} else if !public && p.check(lexer.Question) {
			if len(annotations) > 0 {
				p.errorAt(p.peek(), "annotations cannot be applied to tests")
			}
			test := p.parseTest()
			if test != nil {
				file.Tests = append(file.Tests, test)
			}
		} else if p.looksLikeConstDecl() {
			if len(annotations) > 0 {
				p.errors = append(p.errors, Error{Message: "annotations cannot be applied to constants", Pos: annotations[0].Pos})
			}
			constant := p.parseConstDecl(private)
			if constant != nil {
				file.Constants = append(file.Constants, constant)
			}
		} else if p.looksLikeTypeDecl() {
			typ, enum := p.parseTypeDecl(private)
			if typ != nil {
				typ.Annotations = annotations
				file.Types = append(file.Types, typ)
			}
			if enum != nil {
				enum.Annotations = annotations
				file.Enums = append(file.Enums, enum)
			}
		} else if p.looksLikeFunctionDecl() {
			fn := p.parseFunction(private)
			if fn != nil {
				fn.Annotations = annotations
				file.Functions = append(file.Functions, fn)
			}
		} else {
			if len(annotations) > 0 {
				p.errors = append(p.errors, Error{
					Message: "expected declaration after annotation",
					Pos:     annotations[0].Pos,
				})
			}
			p.errorAt(p.peek(), "expected declaration")
			p.advance()
		}
		p.skipNewlines()
	}
	return file, p.errors
}

func (p *Parser) looksLikeGoImportDecl() bool {
	if !p.check(lexer.At) || p.curr+4 >= len(p.tokens) {
		return false
	}
	return p.tokens[p.curr+1].Kind == lexer.Ident &&
		p.tokens[p.curr+1].Lexeme == "go" &&
		p.tokens[p.curr+2].Kind == lexer.Dot &&
		p.tokens[p.curr+3].Kind == lexer.Ident &&
		p.tokens[p.curr+3].Lexeme == "import" &&
		p.tokens[p.curr+4].Kind == lexer.LParen
}
