package lsp

import (
	"fmt"
	"sort"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (s *server) inlayHints(uri string) any {
	text := s.docs[uri]
	prog, _ := s.analyze(uri)
	if prog == nil {
		return []any{}
	}
	var hints []map[string]any
	for _, fn := range prog.File.Functions {
		if !sourceMatchesDocument(uri, fn.SourcePath) || !hasSourcePosition(fn.NamePos) {
			continue
		}
		fnInfo := prog.Info.FunctionDecls[fn]
		if fnInfo == nil {
			continue
		}
		for i, param := range fn.Params {
			if !hasSourcePosition(param.Pos) || !param.Type.IsZero() || i >= len(fnInfo.Params) {
				continue
			}
			typ := fnInfo.Params[i].Type
			if typ == "" || typ == checker.Unknown {
				continue
			}
			pos := position{
				Line:      max(param.Pos.Line-1, 0),
				Character: max(param.Pos.Column-1, 0) + len(param.Name),
			}
			hints = append(hints, map[string]any{
				"position": pos,
				"label":    ": " + displayCheckerTypeOneLine(prog.Info, typ),
				"kind":     1,
				"tooltip":  functionSignature(prog.Info, fn),
			})
		}
		if !fn.ReturnType.IsZero() || fnInfo.Return == "" || fnInfo.Return == checker.Unknown {
			continue
		}
		if pos, ok := fatArrowPosition(text, fn); ok {
			hints = append(hints, map[string]any{
				"position": pos,
				"label":    "-> " + displayCheckerTypeOneLine(prog.Info, fnInfo.Return) + " ",
				"kind":     1,
				"tooltip":  functionSignature(prog.Info, fn),
			})
		}
	}
	walkDocumentExprs(uri, prog.File, func(expr ast.Expr) {
		lambda, ok := expr.(*ast.LambdaExpr)
		if !ok || !hasSourcePosition(lambda.Pos) {
			return
		}
		params, ret, ok := parseRawDisplayFuncType(string(prog.Info.ExprTypes[lambda]))
		if !ok {
			return
		}
		for i, name := range lambda.Params {
			if i >= len(lambda.ParamPos) || i >= len(params) {
				continue
			}
			if !hasSourcePosition(lambda.ParamPos[i]) {
				continue
			}
			if i < len(lambda.ParamTypes) && !lambda.ParamTypes[i].IsZero() {
				continue
			}
			typ := checker.Type(params[i])
			if typ == "" || typ == checker.Unknown {
				continue
			}
			pos := position{
				Line:      max(lambda.ParamPos[i].Line-1, 0),
				Character: max(lambda.ParamPos[i].Column-1, 0) + len(name),
			}
			hints = append(hints, map[string]any{
				"position": pos,
				"label":    ": " + displayCheckerTypeOneLine(prog.Info, typ),
				"kind":     1,
				"tooltip":  displayCheckerType(prog.Info, prog.Info.ExprTypes[lambda]),
			})
		}
		if lambda.Implicit || !lambda.ReturnType.IsZero() || ret == "" || ret == string(checker.Unknown) {
			return
		}
		if pos, ok := fatArrowPositionFromOffset(text, lambda.Pos.Offset); ok {
			hints = append(hints, map[string]any{
				"position": pos,
				"label":    "-> " + displayCheckerTypeOneLine(prog.Info, checker.Type(ret)) + " ",
				"kind":     1,
				"tooltip":  displayCheckerType(prog.Info, prog.Info.ExprTypes[lambda]),
			})
		}
	})
	walkDocumentExprs(uri, prog.File, func(expr ast.Expr) {
		call, ok := expr.(*ast.CallExpr)
		if !ok || !hasSourcePosition(call.Position()) {
			return
		}
		hints = append(hints, callArgumentNameHints(prog, call)...)
	})
	walkDocumentAnnotations(uri, prog.File, func(annotation *ast.Annotation) {
		if !hasSourcePosition(annotation.Pos) {
			return
		}
		hints = append(hints, annotationArgumentNameHints(prog, annotation)...)
	})
	return normalizeInlayHints(hints)
}

func annotationArgumentNameHints(prog *compiler.Program, annotation *ast.Annotation) []map[string]any {
	var params []checker.ParamInfo
	var tooltip string
	if fn := prog.Info.ResolvedMacros[annotation]; fn != nil {
		params = stdlibParamInfos(fn)
		if len(params) >= 2 {
			params = params[2:]
		}
		tooltip = stdlibSignature(annotation.Module, fn)
	} else if fn := prog.Info.ResolvedMacroFunctions[annotation]; fn != nil {
		params = fn.Params
		if len(params) >= 2 {
			params = params[2:]
		}
		tooltip = functionValueSignature(prog.Info, fn)
	} else {
		return nil
	}
	hints := make([]map[string]any, 0, min(len(annotation.Args), len(params)))
	for i, arg := range annotation.Args {
		if i >= len(params) {
			break
		}
		if !hasSourcePosition(arg.Position()) {
			continue
		}
		name := params[i].Name
		if name == "" || name == "_" {
			continue
		}
		hints = append(hints, map[string]any{
			"position": position{
				Line:      max(arg.Position().Line-1, 0),
				Character: max(arg.Position().Column-1, 0),
			},
			"label":   name + "=",
			"kind":    2,
			"tooltip": tooltip,
		})
	}
	return hints
}

func callArgumentNameHints(prog *compiler.Program, call *ast.CallExpr) []map[string]any {
	params, tooltip, ok := callParameterInfo(prog, call)
	if !ok {
		return nil
	}
	hints := make([]map[string]any, 0, min(len(call.Args), len(params)))
	for i, arg := range call.Args {
		if i >= len(params) {
			break
		}
		if !hasSourcePosition(arg.Position()) {
			continue
		}
		name := params[i].Name
		if name == "" || name == "_" {
			continue
		}
		pos := arg.Position()
		hints = append(hints, map[string]any{
			"position": position{
				Line:      max(pos.Line-1, 0),
				Character: max(pos.Column-1, 0),
			},
			"label":   name + "=",
			"kind":    2,
			"tooltip": tooltip,
		})
	}
	return hints
}

func callParameterInfo(prog *compiler.Program, call *ast.CallExpr) ([]checker.ParamInfo, string, bool) {
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		fn := prog.Info.ResolvedFunctions[callee]
		if fn == nil {
			return nil, "", false
		}
		return fn.Params, functionValueSignature(prog.Info, fn), true
	case *ast.SelectorExpr:
		if at, ok := callee.Receiver.(*ast.AtExpr); ok && at.Name != "" {
			fn, ok := prog.Info.Stdlib.Function(at.Name, callee.Name)
			if !ok {
				return nil, "", false
			}
			return stdlibParamInfos(fn), stdlibSignature(at.Name, fn), true
		}
		receiver := prog.Info.ExprTypes[callee.Receiver]
		if moduleName, ok := checker.ModuleNamespaceName(receiver); ok {
			fn, ok := prog.Info.Stdlib.Function(moduleName, callee.Name)
			if !ok {
				return nil, "", false
			}
			return stdlibParamInfos(fn), stdlibSignature(moduleName, fn), true
		}
		if fn := prog.Info.ResolvedSelectorFunctions[callee]; fn != nil {
			return fn.Params, functionValueSignature(prog.Info, fn), true
		}
		if moduleName, receiverName, ok := stdlibReceiverModule(receiver); ok {
			fn, ok := prog.Info.Stdlib.ReceiverFunction(moduleName, receiverName, callee.Name)
			if !ok {
				return nil, "", false
			}
			return stdlibParamInfos(fn), stdlibSignature(moduleName, fn), true
		}
		structInfo := prog.Info.Types[baseType(receiver)]
		if callee.Static {
			if ident, ok := callee.Receiver.(*ast.Identifier); ok {
				structInfo = prog.Info.Types[ident.Name]
			}
		}
		if structInfo == nil {
			return nil, "", false
		}
		method := structInfo.Methods[callee.Name]
		if callee.Static {
			method = structInfo.StaticMethods[callee.Name]
		}
		if method == nil {
			return nil, "", false
		}
		return method.Params, functionValueSignature(prog.Info, method), true
	default:
		return nil, "", false
	}
}

func stdlibParamInfos(fn *stdlib.Function) []checker.ParamInfo {
	params := make([]checker.ParamInfo, 0, len(fn.Params))
	for i := range fn.Params {
		name := fmt.Sprintf("arg%d", i+1)
		if i < len(fn.ParamNames) && fn.ParamNames[i] != "" {
			name = fn.ParamNames[i]
		}
		params = append(params, checker.ParamInfo{Name: name})
	}
	return params
}

func hasSourcePosition(pos lexer.Position) bool {
	return pos.Line > 0 && pos.Column > 0
}

func normalizeInlayHints(hints []map[string]any) []map[string]any {
	sort.SliceStable(hints, func(i, j int) bool {
		left := hints[i]["position"].(position)
		right := hints[j]["position"].(position)
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Character != right.Character {
			return left.Character < right.Character
		}
		return fmt.Sprint(hints[i]["label"]) < fmt.Sprint(hints[j]["label"])
	})

	normalized := hints[:0]
	var previous string
	for _, hint := range hints {
		pos := hint["position"].(position)
		key := fmt.Sprintf("%d:%d:%v:%v", pos.Line, pos.Character, hint["kind"], hint["label"])
		if len(normalized) > 0 && key == previous {
			continue
		}
		normalized = append(normalized, hint)
		previous = key
	}
	return normalized
}
