package macro

import (
	"fmt"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/interpreter"
	"github.com/oboard/rune-lang/internal/ir"
)

func Expand(file *ast.File, info *checker.Info) (bool, []checker.Diagnostic) {
	invocations := Plan(file, info)
	if len(invocations) == 0 {
		return false, nil
	}
	changed := false
	var diags []checker.Diagnostic
	for _, invocation := range invocations {
		didChange, err := executeInvocation(invocation, file, info)
		if err != nil {
			diags = append(diags, checker.Diagnostic{
				Message: fmt.Sprintf("macro #%s failed: %v", invocationName(invocation), err),
				Pos:     invocation.Annotation.Pos,
			})
			continue
		}
		changed = changed || didChange
	}
	return changed, diags
}

func executeInvocation(invocation Invocation, file *ast.File, info *checker.Info) (bool, error) {
	runtime := interpreter.New(&ir.File{Stdlib: info.Stdlib}, interpreter.WithCompileTime())
	args := make([]interpreter.Value, 0, len(invocation.Annotation.Args))
	for _, arg := range invocation.Annotation.Args {
		value, err := runtime.Eval(ir.LowerExpr(arg, info))
		if err != nil {
			return false, fmt.Errorf("evaluate argument: %w", err)
		}
		args = append(args, value)
	}

	body, paramNames, returnType, err := invocationBody(invocation)
	if err != nil {
		return false, err
	}
	const hiddenParams = 2
	if len(paramNames)-hiddenParams != len(args) {
		return false, fmt.Errorf("expects %d args, got %d", len(paramNames)-hiddenParams, len(args))
	}

	refs := newSyntaxRefs()
	bindings := make(map[string]interpreter.Value, len(args)+hiddenParams)
	bindings[paramNames[0]] = syntaxFileValueForSource(file, refs, invocation.Target.SourcePath)
	bindings[paramNames[1]] = macroContextValue(invocation.Target)
	for i, name := range paramNames[hiddenParams:] {
		bindings[name] = args[i]
	}
	result, err := runtime.EvalWithBindings(ir.LowerExprExpected(body, info, returnType), bindings)
	if err != nil {
		return false, err
	}
	expanded, err := decodeSyntaxFile(result, file, refs)
	if err != nil {
		return false, err
	}
	*file = *expanded
	return true, nil
}

func invocationBody(invocation Invocation) (ast.Expr, []string, checker.Type, error) {
	switch {
	case invocation.Macro != nil:
		if invocation.Macro.Body == nil {
			return nil, nil, checker.Unknown, fmt.Errorf("macro body is not available")
		}
		return invocation.Macro.Body, invocation.Macro.ParamNames, checker.Type(invocation.Macro.Return), nil
	case invocation.LocalMacro != nil && invocation.LocalMacro.Node != nil:
		fn := invocation.LocalMacro.Node
		if fn.Body == nil {
			return nil, nil, checker.Unknown, fmt.Errorf("macro body is not available")
		}
		names := make([]string, 0, len(fn.Params))
		for _, param := range fn.Params {
			names = append(names, param.Name)
		}
		return fn.Body, names, invocation.LocalMacro.Return, nil
	default:
		return nil, nil, checker.Unknown, fmt.Errorf("resolved macro has no Rune body")
	}
}

func invocationName(invocation Invocation) string {
	if invocation.Annotation.Module == "" {
		return invocation.Annotation.Name
	}
	return invocation.Annotation.Module + "." + invocation.Annotation.Name
}
