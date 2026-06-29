package lsp

import (
	"strconv"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (s *server) definition(uri string, pos position) any {
	if target := s.importDefinition(uri, pos); target != nil {
		return target
	}
	prog, _ := s.analyze(uri)
	word := wordAt(s.docs[uri], pos)
	if word == "" {
		return nil
	}
	if prog == nil {
		if target := s.stdlibModuleDefinitionFromTokens(uri, pos); target != nil {
			return target.location()
		}
		return nil
	}
	if target := annotationTarget(uri, prog, pos); target != nil {
		return target.location()
	}
	if target := s.methodTarget(uri, prog, pos); target != nil {
		return target.location()
	}
	if target := fieldTarget(uri, prog, pos); target != nil {
		return target.location()
	}
	if target := traitTarget(uri, prog, pos); target != nil {
		return target.location()
	}
	if target := typeTarget(uri, prog, pos); target != nil {
		return target.location()
	}
	if target := localTarget(uri, prog, pos); target != nil {
		return target.location()
	}
	if target := externalValueTarget(uri, prog, pos); target != nil {
		return target.location()
	}
	if target := functionTarget(uri, prog, pos); target != nil {
		return target.location()
	}
	if target := s.stdlibModuleDefinitionFromTokens(uri, pos); target != nil {
		return target.location()
	}
	return nil
}

func (s *server) stdlibModuleDefinitionFromTokens(uri string, pos position) *methodTarget {
	reg, err := stdlib.LoadDefault()
	if err != nil {
		return nil
	}
	return stdlibModuleDefinitionInSource(s.docs[uri], reg, pos)
}

func stdlibModuleDefinitionInSource(text string, reg *stdlib.Registry, pos position) *methodTarget {
	tokens := lexer.Lex(text)
	for i := 0; i+3 < len(tokens); i++ {
		if tokens[i].Kind != lexer.At || tokens[i+1].Kind != lexer.Ident || tokens[i+2].Kind != lexer.Dot || tokens[i+3].Kind != lexer.Ident {
			continue
		}
		module := tokens[i+1]
		name := tokens[i+3]
		if !containsToken(pos, module) && !containsToken(pos, name) {
			continue
		}
		return stdlibTarget(reg, module.Lexeme, name.Lexeme)
	}
	return nil
}

func annotationTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	annotation := annotationAt(prog.File, pos)
	if annotation == nil {
		return nil
	}
	if fn := prog.Info.ResolvedMacros[annotation]; fn != nil {
		return stdlibMacroTarget(prog.Info.Stdlib, annotation.Module, fn.Name)
	}
	if fn := prog.Info.ResolvedMacroFunctions[annotation]; fn != nil {
		return functionInfoTarget(uri, fn)
	}
	return nil
}

func (s *server) importDefinition(uri string, pos position) any {
	text := s.docs[uri]
	fromPath := filePathFromURI(uri)
	if fromPath == "" {
		return nil
	}
	tokens := lexer.Lex(text)
	for i := 0; i+1 < len(tokens); i++ {
		at := tokens[i]
		path := tokens[i+1]
		if at.Kind != lexer.At || path.Kind != lexer.String {
			continue
		}
		if !containsToken(pos, at) && !containsToken(pos, path) {
			continue
		}
		spec, err := strconv.Unquote(path.Lexeme)
		if err != nil {
			return nil
		}
		targetPath, err := compiler.ResolveRuneImport(fromPath, spec)
		if err != nil {
			return nil
		}
		zero := position{Line: 0, Character: 0}
		return map[string]any{
			"uri": fileURI(targetPath),
			"range": map[string]any{
				"start": zero,
				"end":   zero,
			},
		}
	}
	return nil
}

func typeTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	name := wordAt(prog.Source, pos)
	if name == "" {
		return nil
	}
	for _, typ := range prog.File.Types {
		if typ.Name == name {
			return &methodTarget{uri: sourceURI(uri, typ.SourcePath), name: typ.Name, pos: typ.NamePos}
		}
	}
	for _, enum := range prog.File.Enums {
		if enum.Name == name {
			return &methodTarget{uri: sourceURI(uri, enum.SourcePath), name: enum.Name, pos: enum.NamePos}
		}
	}
	for _, trait := range prog.File.Traits {
		if trait.Name == name {
			return &methodTarget{uri: sourceURI(uri, trait.SourcePath), name: trait.Name, pos: trait.NamePos, traitName: trait.Name}
		}
	}
	return nil
}

func fieldTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	sel := selectorAt(prog.File, pos)
	if sel == nil {
		return nil
	}
	receiver := prog.Info.ExprTypes[sel.Receiver]
	structInfo := prog.Info.Types[baseType(receiver)]
	if structInfo == nil {
		return nil
	}
	if structInfo.Node != nil {
		for _, field := range structInfo.Node.Fields {
			if field.Name == sel.Name {
				return &methodTarget{uri: sourceURI(uri, structInfo.Node.SourcePath), name: field.Name, pos: field.Pos, structName: structInfo.Name}
			}
		}
	}
	return anonymousFieldTarget(uri, prog, structInfo.Name, sel.Name)
}

func anonymousFieldTarget(uri string, prog *compiler.Program, typeName string, fieldName string) *methodTarget {
	var found *methodTarget
	walkFileExprs(prog.File, func(expr ast.Expr) {
		if found != nil {
			return
		}
		obj, ok := expr.(*ast.AnonymousObjectLiteral)
		if !ok {
			return
		}
		if baseType(prog.Info.ExprTypes[obj]) != typeName {
			objTypeName := baseType(prog.Info.ExprTypes[obj])
			if !anonymousObjectTypeCanSatisfyField(prog.Info, objTypeName, typeName, fieldName) {
				return
			}
		}
		for _, field := range obj.Fields {
			if field.Name == fieldName {
				found = &methodTarget{uri: uri, name: field.Name, pos: field.Pos, structName: typeName}
				return
			}
		}
	})
	return found
}

func anonymousObjectTypeCanSatisfyField(info *checker.Info, objectType string, targetType string, fieldName string) bool {
	if info == nil {
		return false
	}
	objectInfo := info.Types[objectType]
	targetInfo := info.Types[targetType]
	if objectInfo == nil || targetInfo == nil {
		return false
	}
	if _, ok := targetInfo.ByName[fieldName]; !ok {
		return false
	}
	for _, targetField := range targetInfo.Fields {
		if _, ok := objectInfo.ByName[targetField.Name]; !ok {
			return false
		}
	}
	return true
}
