package lsp

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func functionTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	for _, fn := range prog.File.Functions {
		if containsSymbol(pos, fn.NamePos, fn.Name) {
			return &methodTarget{uri: sourceURI(uri, fn.SourcePath), name: fn.Name, pos: fn.NamePos}
		}
	}
	if ident := identifierAt(prog.File, pos); ident != nil {
		if fn := prog.Info.ResolvedFunctions[ident]; fn != nil {
			return functionInfoTarget(uri, fn)
		}
	}
	call := functionCallAt(prog.File, pos)
	if call == nil {
		return nil
	}
	if fn := prog.Info.ResolvedFunctions[call]; fn != nil {
		return functionInfoTarget(uri, fn)
	}
	return nil
}

func functionInfoTarget(uri string, fn *checker.FuncInfo) *methodTarget {
	if fn == nil {
		return nil
	}
	pos := fn.NamePos
	if pos.Line == 0 && fn.Node != nil {
		pos = fn.Node.NamePos
	}
	return &methodTarget{uri: sourceURI(uri, fn.SourcePath), name: fn.Name, pos: pos, external: fn.External}
}

func externalValueTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	ident := identifierAt(prog.File, pos)
	if ident == nil {
		return nil
	}
	value := prog.Info.ResolvedValues[ident]
	if value == nil {
		return nil
	}
	return &methodTarget{uri: sourceURI(uri, value.SourcePath), name: value.Name, pos: value.NamePos, external: true}
}

func functionCallAt(file *ast.File, pos position) *ast.Identifier {
	var found *ast.Identifier
	walkFileCalls(file, func(call *ast.CallExpr) {
		if found != nil {
			return
		}
		ident, ok := call.Callee.(*ast.Identifier)
		if ok && containsSymbol(pos, ident.Pos, ident.Name) {
			found = ident
		}
	})
	return found
}

type methodTarget struct {
	uri        string
	name       string
	pos        lexer.Position
	structName string
	external   bool
	scope      ast.Expr
}

func (t *methodTarget) location() map[string]any {
	return map[string]any{
		"uri":   t.uri,
		"range": symbolRange(t.pos, len(t.name)),
	}
}

func (s *server) methodTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	if target := methodDeclTarget(uri, prog.File, pos); target != nil {
		return target
	}
	sel := selectorAt(prog.File, pos)
	if sel == nil {
		return nil
	}
	if at, ok := sel.Receiver.(*ast.AtExpr); ok && at.Name != "" {
		return stdlibTarget(prog.Info.Stdlib, at.Name, sel.Name)
	}
	receiver := prog.Info.ExprTypes[sel.Receiver]
	if moduleName, ok := checker.ModuleNamespaceName(receiver); ok {
		return stdlibTarget(prog.Info.Stdlib, moduleName, sel.Name)
	}
	if fn := prog.Info.ResolvedSelectorFunctions[sel]; fn != nil {
		return functionInfoTarget(uri, fn)
	}
	if value := prog.Info.ResolvedSelectorValues[sel]; value != nil {
		return &methodTarget{uri: sourceURI(uri, value.SourcePath), name: value.Name, pos: value.NamePos, external: true}
	}
	if moduleName, receiverName, ok := stdlibReceiverModule(receiver); ok {
		return stdlibReceiverTarget(prog.Info.Stdlib, moduleName, receiverName, sel.Name)
	}
	structInfo := prog.Info.Types[baseType(receiver)]
	if structInfo == nil {
		return nil
	}
	method := structInfo.Methods[sel.Name]
	if method == nil || method.Node == nil {
		return nil
	}
	return &methodTarget{
		uri:        sourceURI(uri, method.Node.SourcePath),
		name:       method.Node.Name,
		pos:        method.Node.NamePos,
		structName: structInfo.Name,
	}
}

func methodDeclTarget(uri string, file *ast.File, pos position) *methodTarget {
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			if containsSymbol(pos, method.NamePos, method.Name) {
				return &methodTarget{
					uri:        sourceURI(uri, method.SourcePath),
					name:       method.Name,
					pos:        method.NamePos,
					structName: typ.Name,
				}
			}
		}
	}
	return nil
}

func stdlibTarget(reg *stdlib.Registry, moduleName string, functionName string) *methodTarget {
	if reg == nil {
		return nil
	}
	fn, ok := reg.Function(moduleName, functionName)
	if !ok || fn.SourcePath == "" {
		return nil
	}
	return &methodTarget{
		uri:  fileURI(fn.SourcePath),
		name: fn.Name,
		pos:  fn.Pos,
	}
}

func stdlibReceiverTarget(reg *stdlib.Registry, moduleName string, receiverName string, functionName string) *methodTarget {
	if reg == nil {
		return nil
	}
	fn, ok := reg.ReceiverFunction(moduleName, receiverName, functionName)
	if !ok || fn.SourcePath == "" {
		return nil
	}
	return &methodTarget{
		uri:  fileURI(fn.SourcePath),
		name: fn.Name,
		pos:  fn.Pos,
	}
}
