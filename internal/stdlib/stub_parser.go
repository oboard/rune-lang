package stdlib

import (
	"fmt"

	"github.com/oboard/rune-lang/internal/lexer"
)

type annotation struct {
	Name  string
	Value string
}

type stubParser struct {
	moduleName string
	path       string
	tokens     []lexer.Token
	curr       int
}

func newStubParser(moduleName string, path string, src string) *stubParser {
	return &stubParser{
		moduleName: moduleName,
		path:       path,
		tokens:     lexer.Lex(src),
	}
}

func (p *stubParser) parse() (*Module, error) {
	mod := &Module{
		Name:       p.moduleName,
		byName:     map[string]*Function{},
		byReceiver: map[string]map[string]*Function{},
		byAlias:    map[string]*Function{},
	}
	seen := map[string]bool{}
	p.skipNewlines()
	for !p.check(lexer.EOF) {
		annotations, err := p.parseAnnotations()
		if err != nil {
			return nil, err
		}
		if p.check(lexer.EOF) {
			break
		}
		if !p.check(lexer.Ident) && !p.check(lexer.Tilde) {
			return nil, p.errorf(p.peek(), "expected core declaration")
		}

		if p.check(lexer.Ident) && p.looksLikeReceiverBlock() {
			typ, functions, err := p.parseReceiverBlock()
			if err != nil {
				return nil, err
			}
			if typ != nil {
				mod.Types = append(mod.Types, *typ)
			}
			for _, fn := range functions {
				if err := addFunction(mod, seen, fn); err != nil {
					return nil, err
				}
			}
		} else {
			fn, err := p.parseFunction("", annotations)
			if err != nil {
				return nil, err
			}
			if err := addFunction(mod, seen, fn); err != nil {
				return nil, err
			}
		}
		p.skipNewlines()
	}
	for i := range mod.Functions {
		fn := &mod.Functions[i]
		if fn.Receiver == "" {
			mod.byName[fn.Name] = fn
		} else {
			if _, exists := mod.byName[fn.Name]; !exists {
				mod.byName[fn.Name] = fn
			}
			methods := mod.byReceiver[fn.Receiver]
			if methods == nil {
				methods = map[string]*Function{}
				mod.byReceiver[fn.Receiver] = methods
			}
			methods[fn.Name] = fn
		}
		if fn.Alias != "" {
			mod.byAlias[fn.Alias] = fn
		}
	}
	return mod, nil
}

func addFunction(mod *Module, seen map[string]bool, fn Function) error {
	key := fn.Receiver + "." + fn.Name
	if seen[key] {
		return fmt.Errorf("duplicate function %s.%s", mod.Name, key)
	}
	seen[key] = true
	mod.Functions = append(mod.Functions, fn)
	return nil
}
