package lsp

import (
	"fmt"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (s *server) hover(uri string, pos position) any {
	prog, _ := s.analyze(uri)
	word := wordAt(s.docs[uri], pos)
	if word == "" || prog == nil {
		return nil
	}
	if h := annotationHover(prog, pos); h != nil {
		return h
	}
	if h := s.methodHover(prog, pos); h != nil {
		return h
	}
	if h := s.exprHover(prog, pos); h != nil {
		return h
	}
	if h := typeHover(prog, pos); h != nil {
		return h
	}
	for _, fn := range prog.File.Functions {
		if fn.Name == word {
			signature := functionSignature(prog.Info, fn)
			if fn.Routine {
				if fnInfo := prog.Info.FunctionDecls[fn]; fnInfo != nil {
					signature = functionValueSignature(prog.Info, fnInfo)
				}
			}
			return map[string]any{
				"contents": map[string]any{
					"kind":  "markdown",
					"value": fmt.Sprintf("```rune\n%s\n```", signature),
				},
				"range": symbolRange(fn.NamePos, len(fn.Name)),
			}
		}
		fnInfo := prog.Info.FunctionDecls[fn]
		for i, param := range fn.Params {
			if param.Name == word {
				typ := param.Type.Display()
				if fnInfo != nil && i < len(fnInfo.Params) && fnInfo.Params[i].Type != "" && fnInfo.Params[i].Type != checker.Unknown {
					typ = displayCheckerType(prog.Info, fnInfo.Params[i].Type)
				}
				if typ == "" {
					typ = string(checker.Unknown)
				}
				return map[string]any{
					"contents": map[string]any{
						"kind":  "markdown",
						"value": fmt.Sprintf("```rune\n%s: %s\n```", param.Name, typ),
					},
					"range": symbolRange(param.Pos, len(param.Name)),
				}
			}
		}
	}
	return nil
}

func annotationHover(prog *compiler.Program, pos position) any {
	annotation := annotationAt(prog.File, pos)
	if annotation == nil {
		return nil
	}
	if fn := prog.Info.ResolvedMacros[annotation]; fn != nil {
		return stdlibFunctionHover(annotation.Module, fn, annotation.NamePos)
	}
	if fn := prog.Info.ResolvedMacroFunctions[annotation]; fn != nil {
		return hoverResult(functionValueSignature(prog.Info, fn), annotation.NamePos, annotation.Name)
	}
	return nil
}

func (s *server) methodHover(prog *compiler.Program, pos position) any {
	if hover := methodDeclHover(prog, pos); hover != nil {
		return hover
	}
	sel := selectorAt(prog.File, pos)
	if sel == nil {
		return nil
	}
	if at, ok := sel.Receiver.(*ast.AtExpr); ok && at.Name != "" {
		return stdlibHover(prog.Info.Stdlib, at.Name, sel.Name, sel.NamePos)
	}
	receiver := prog.Info.ExprTypes[sel.Receiver]
	if moduleName, ok := checker.ModuleNamespaceName(receiver); ok {
		return stdlibHover(prog.Info.Stdlib, moduleName, sel.Name, sel.NamePos)
	}
	if fn := prog.Info.ResolvedSelectorFunctions[sel]; fn != nil {
		return hoverResult(functionValueSignature(prog.Info, fn), sel.NamePos, sel.Name)
	}
	if value := prog.Info.ResolvedSelectorValues[sel]; value != nil {
		return hoverResult(fmt.Sprintf("%s: %s", value.Name, displayCheckerType(prog.Info, value.Type)), sel.NamePos, sel.Name)
	}
	if moduleName, receiverName, ok := stdlibReceiverModule(receiver); ok {
		return stdlibReceiverHover(prog.Info.Stdlib, moduleName, receiverName, sel.Name, sel.NamePos)
	}
	if traitInfo := traitInfoForType(prog.Info, receiver); traitInfo != nil {
		if field, ok := traitInfo.ByName[sel.Name]; ok {
			return hoverResult(fmt.Sprintf("%s: %s", sel.Name, displayCheckerType(prog.Info, field.Type)), sel.NamePos, sel.Name)
		}
		if method := traitInfo.Methods[sel.Name]; method != nil {
			return hoverResult(traitMemberSignature(prog.Info, traitInfo.Name, method), sel.NamePos, sel.Name)
		}
	}
	structInfo := prog.Info.Types[baseType(receiver)]
	if structInfo == nil {
		return nil
	}
	if field, ok := structInfo.ByName[sel.Name]; ok {
		return hoverResult(fmt.Sprintf("%s: %s", sel.Name, displayCheckerType(prog.Info, field.Type)), sel.NamePos, sel.Name)
	}
	method := structInfo.Methods[sel.Name]
	if method == nil || method.Node == nil {
		return nil
	}
	return hoverResult(classMethodSignature(structInfo.Name, method), sel.NamePos, sel.Name)
}

func (s *server) exprHover(prog *compiler.Program, pos position) any {
	var found any
	signals := signalGraph(prog.File)
	walkFileStatements(prog.File, func(stmt ast.Stmt) {
		if found != nil {
			return
		}
		if let, ok := stmt.(*ast.LetStmt); ok && containsSymbol(pos, let.Pos, let.Name) {
			if typ := prog.Info.ExprTypes[let.Value]; typ != "" && typ != checker.Unknown {
				found = hoverResult(localHoverText(prog.Info, let.Name, typ, signals), let.Pos, let.Name)
			}
		}
		if destructure, ok := stmt.(*ast.ObjectDestructureStmt); ok {
			for _, field := range destructure.Fields {
				if !containsSymbol(pos, field.NamePos, field.Name) {
					continue
				}
				valueType := prog.Info.ExprTypes[destructure.Value]
				if typ, ok := checker.FieldType(prog.Info, valueType, field.Field); ok && typ != checker.Unknown {
					found = hoverResult(localHoverText(prog.Info, field.Name, typ, signals), field.NamePos, field.Name)
				}
				return
			}
		}
	})
	if found != nil {
		return found
	}
	walkFileExprs(prog.File, func(expr ast.Expr) {
		if found != nil {
			return
		}
		switch e := expr.(type) {
		case *ast.Identifier:
			if !containsSymbol(pos, e.Pos, e.Name) {
				return
			}
			if fn := prog.Info.ResolvedFunctions[e]; fn != nil && identifierHasFunctionType(prog.Info, e, fn) {
				found = hoverResult(functionValueSignature(prog.Info, fn), e.Pos, e.Name)
				return
			}
			if value := prog.Info.ResolvedValues[e]; value != nil {
				found = hoverResult(fmt.Sprintf("%s: %s", value.Name, displayCheckerType(prog.Info, value.Type)), e.Pos, e.Name)
				return
			}
			if typ := prog.Info.ExprTypes[e]; typ != "" && typ != checker.Unknown {
				found = hoverResult(localHoverText(prog.Info, e.Name, typ, signals), e.Pos, e.Name)
			}
		case *ast.ThisExpr:
			if !containsSymbol(pos, e.Pos, "this") {
				return
			}
			if typ := prog.Info.ExprTypes[e]; typ != "" && typ != checker.Unknown {
				found = hoverResult(fmt.Sprintf("this: %s", displayCheckerType(prog.Info, typ)), e.Pos, "this")
			}
		}
	})
	return found
}

func localHoverText(info *checker.Info, name string, typ checker.Type, signals map[string][]string) string {
	text := fmt.Sprintf("%s: %s", name, displayCheckerType(info, typ))
	deps, ok := signals[name]
	if !ok {
		return text
	}
	kind := "signal"
	if len(deps) > 0 {
		kind = "computed"
	}
	text += "\n" + kind
	if chain := dependencyChain(name, signals); chain != "" {
		text += "\ndeps: " + chain
	}
	return text
}

func identifierHasFunctionType(info *checker.Info, ident *ast.Identifier, fn *checker.FuncInfo) bool {
	typ := info.ExprTypes[ident]
	return typ != "" && typ != checker.Unknown && typ == funcInfoType(fn)
}

func funcInfoType(fn *checker.FuncInfo) checker.Type {
	params := make([]checker.Type, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, param.Type)
	}
	if fn.Routine {
		return checker.AsyncFuncOfTypes(params, fn.Return)
	}
	return checker.FuncOfTypes(params, fn.Return)
}

func typeHover(prog *compiler.Program, pos position) any {
	if name, tokPos, ok := structLiteralTypeAt(prog.File, pos); ok {
		if typ := structTypeByName(prog.File, name); typ != nil {
			return hoverResult(structTypeSignature(prog.Info, typ), tokPos, name)
		}
	}
	for _, typ := range prog.File.Types {
		if containsSymbol(pos, typ.NamePos, typ.Name) {
			return hoverResult(structTypeSignature(prog.Info, typ), typ.NamePos, typ.Name)
		}
	}
	for _, trait := range prog.File.Traits {
		if containsSymbol(pos, trait.NamePos, trait.Name) {
			return hoverResult(traitTypeSignature(prog.Info, trait), trait.NamePos, trait.Name)
		}
	}
	for _, enum := range prog.File.Enums {
		if containsSymbol(pos, enum.NamePos, enum.Name) {
			return hoverResult(enumTypeSignature(enum), enum.NamePos, enum.Name)
		}
	}
	return nil
}

func structLiteralTypeAt(file *ast.File, pos position) (string, lexer.Position, bool) {
	var name string
	var tokPos lexer.Position
	walkFileExprs(file, func(expr ast.Expr) {
		if name != "" {
			return
		}
		lit, ok := expr.(*ast.StructLiteral)
		if !ok || !containsSymbol(pos, lit.Pos, lit.TypeName) {
			return
		}
		name = lit.TypeName
		tokPos = lit.Pos
	})
	return name, tokPos, name != ""
}

func structTypeByName(file *ast.File, name string) *ast.StructType {
	for _, typ := range file.Types {
		if typ.Name == name {
			return typ
		}
	}
	return nil
}

func methodDeclHover(prog *compiler.Program, pos position) any {
	for _, trait := range prog.File.Traits {
		traitInfo := prog.Info.Traits[trait.Name]
		if traitInfo == nil {
			continue
		}
		for _, field := range trait.Fields {
			if !containsSymbol(pos, field.Pos, field.Name) {
				continue
			}
			if info, ok := traitInfo.ByName[field.Name]; ok {
				return hoverResult(fmt.Sprintf("%s: %s", field.Name, displayCheckerType(prog.Info, info.Type)), field.Pos, field.Name)
			}
		}
		for _, method := range trait.Methods {
			if !containsSymbol(pos, method.NamePos, method.Name) {
				continue
			}
			var info *checker.FuncInfo
			if method.Static {
				info = traitInfo.StaticMethods[method.Name]
			} else {
				info = traitInfo.Methods[method.Name]
			}
			if info == nil {
				continue
			}
			return hoverResult(traitMemberSignature(prog.Info, trait.Name, info), method.NamePos, method.Name)
		}
	}
	for _, typ := range prog.File.Types {
		structInfo := prog.Info.Types[typ.Name]
		if structInfo == nil {
			continue
		}
		for _, method := range typ.Methods {
			if !containsSymbol(pos, method.NamePos, method.Name) {
				continue
			}
			info := structInfo.Methods[method.Name]
			if info == nil {
				continue
			}
			return hoverResult(classMethodSignature(typ.Name, info), method.NamePos, method.Name)
		}
	}
	return nil
}

func stdlibHover(reg *stdlib.Registry, moduleName string, functionName string, pos lexer.Position) any {
	if reg == nil {
		return nil
	}
	fn, ok := reg.Function(moduleName, functionName)
	if !ok {
		return nil
	}
	return stdlibFunctionHover(moduleName, fn, pos)
}

func stdlibReceiverHover(reg *stdlib.Registry, moduleName string, receiverName string, functionName string, pos lexer.Position) any {
	if reg == nil {
		return nil
	}
	fn, ok := reg.ReceiverFunction(moduleName, receiverName, functionName)
	if !ok {
		return nil
	}
	return stdlibFunctionHover(moduleName, fn, pos)
}

func stdlibFunctionHover(moduleName string, fn *stdlib.Function, pos lexer.Position) any {
	value := fmt.Sprintf("```rune\n%s\n```", stdlibSignature(moduleName, fn))
	if fn.SourcePath != "" {
		value += "\n" + fn.SourcePath
	}
	return map[string]any{
		"contents": map[string]any{
			"kind":  "markdown",
			"value": value,
		},
		"range": symbolRange(pos, len(fn.Name)),
	}
}

func hoverResult(signature string, pos lexer.Position, name string) any {
	return map[string]any{
		"contents": map[string]any{
			"kind":  "markdown",
			"value": fmt.Sprintf("```rune\n%s\n```", signature),
		},
		"range": symbolRange(pos, len(name)),
	}
}
