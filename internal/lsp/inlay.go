package lsp

import (
	"fmt"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/compiler"
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
		if !sourceMatchesDocument(uri, fn.SourcePath) {
			continue
		}
		fnInfo := prog.Info.FunctionDecls[fn]
		if fnInfo == nil {
			continue
		}
		for i, param := range fn.Params {
			if !param.Type.IsZero() || i >= len(fnInfo.Params) {
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
		if !ok {
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
		if !ok {
			return
		}
		hints = append(hints, callArgumentNameHints(prog, call)...)
	})
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
		if structInfo == nil {
			return nil, "", false
		}
		method := structInfo.Methods[callee.Name]
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
