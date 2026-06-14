package lsp

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func classMethodSignature(typeName string, fn *checker.FuncInfo) string {
	ret := fn.Return
	if ret == "" {
		ret = checker.Void
	}
	sig := methodNodeSignature(fn.Node)
	return fmt.Sprintf("%s.%s -> %s", typeName, sig, ret)
}

func structTypeSignature(info *checker.Info, typ *ast.StructType) string {
	lines := []string{fmt.Sprintf("%s%s: {", typ.Name, formatSignatureGenerics(typ.Generics))}
	var structInfo *checker.StructInfo
	if info != nil {
		structInfo = info.Types[typ.Name]
	}
	for _, field := range typ.Fields {
		fieldType := field.Type.Display()
		if structInfo != nil {
			if inferred, ok := structInfo.ByName[field.Name]; ok && inferred.Type != "" && inferred.Type != checker.Unknown {
				fieldType = displayCheckerType(info, inferred.Type)
			}
		}
		if fieldType == "" {
			fieldType = string(checker.Unknown)
		}
		lines = append(lines, fmt.Sprintf("  %s: %s", field.Name, fieldType))
	}
	for _, method := range typ.Methods {
		lines = append(lines, "  "+methodSignature(info, typ.Name, method))
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func methodSignature(info *checker.Info, typeName string, fn *ast.Function) string {
	params := make([]string, 0, len(fn.Params))
	var methodInfo *checker.FuncInfo
	if info != nil {
		if structInfo := info.Types[typeName]; structInfo != nil {
			methodInfo = structInfo.Methods[fn.Name]
		}
	}
	for i, param := range fn.Params {
		typ := param.Type.Display()
		if methodInfo != nil && i < len(methodInfo.Params) && methodInfo.Params[i].Type != "" && methodInfo.Params[i].Type != checker.Unknown {
			typ = displayCheckerType(info, methodInfo.Params[i].Type)
		}
		if typ == "" {
			typ = string(checker.Unknown)
		}
		params = append(params, fmt.Sprintf("%s: %s", param.Name, typ))
	}
	ret := fn.ReturnType.Display()
	if methodInfo != nil && methodInfo.Return != "" && methodInfo.Return != checker.Unknown {
		ret = displayCheckerType(info, methodInfo.Return)
	}
	if ret == "" {
		ret = string(checker.Void)
	}
	return fmt.Sprintf("%s%s%s(%s) -> %s", methodPrefix(fn), fn.Name, formatSignatureGenerics(fn.Generics), strings.Join(params, ", "), ret)
}

func methodNodeSignature(fn *ast.Function) string {
	sig := fn.Signature()
	if fn.Private {
		return "- " + sig
	}
	return strings.TrimPrefix(sig, "+ ")
}

func methodPrefix(fn *ast.Function) string {
	prefix := ""
	if fn.Private {
		prefix += "- "
	}
	if fn.Routine {
		prefix += "~ "
	}
	return prefix
}

func enumTypeSignature(enum *ast.EnumType) string {
	if len(enum.Members) == 0 {
		return fmt.Sprintf("%s: {}", enum.Name)
	}
	lines := []string{fmt.Sprintf("%s: {", enum.Name)}
	for _, member := range enum.Members {
		lines = append(lines, fmt.Sprintf("  %s = %d", member.Name, member.Value))
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func functionSignature(info *checker.Info, fn *ast.Function) string {
	fnInfo := info.FunctionDecls[fn]
	if fnInfo == nil {
		return fn.Signature()
	}
	params := make([]string, 0, len(fn.Params))
	for i, param := range fn.Params {
		typ := param.Type.Display()
		if i < len(fnInfo.Params) && fnInfo.Params[i].Type != "" && fnInfo.Params[i].Type != checker.Unknown {
			typ = displayCheckerType(info, fnInfo.Params[i].Type)
		}
		if typ == "" {
			typ = string(checker.Unknown)
		}
		params = append(params, fmt.Sprintf("%s: %s", param.Name, typ))
	}
	ret := fn.ReturnType.Display()
	if fnInfo.Return != "" && fnInfo.Return != checker.Unknown {
		ret = displayCheckerType(info, fnInfo.Return)
	}
	if ret == "" {
		ret = string(checker.Void)
	}
	return fmt.Sprintf("%s%s(%s) -> %s", fn.Name, formatSignatureGenerics(fn.Generics), strings.Join(params, ", "), ret)
}

func functionValueSignature(info *checker.Info, fn *checker.FuncInfo) string {
	fnParams := fn.Params
	if fn.Macro && len(fnParams) >= 2 {
		fnParams = fnParams[2:]
	}
	params := make([]string, 0, len(fnParams))
	for _, param := range fnParams {
		typ := param.Type
		if typ == "" || typ == checker.Unknown {
			typ = checker.Unknown
		}
		params = append(params, fmt.Sprintf("%s: %s", param.Name, displayCheckerType(info, typ)))
	}
	ret := fn.Return
	if ret == "" || (ret == checker.Unknown && !fn.External) {
		ret = checker.Void
	}
	prefix := ""
	if fn.Routine {
		prefix = "~ "
	}
	name := fn.Name
	if fn.Macro {
		name = "#" + name
	}
	return fmt.Sprintf("%s: %s%s(%s) -> %s", name, prefix, formatSignatureGenerics(fn.Generics), strings.Join(params, ", "), displayCheckerType(info, ret))
}

func formatSignatureGenerics(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return "[" + strings.Join(names, ", ") + "]"
}

func stdlibSignature(moduleName string, fn *stdlib.Function) string {
	owner := "@" + moduleName
	if fn.Receiver != "" {
		owner = fn.Receiver
	}
	generics := ""
	if len(fn.Generics) > 0 {
		generics = "[" + strings.Join(fn.Generics, ", ") + "]"
	}
	paramTypes := fn.Params
	paramNames := fn.ParamNames
	if fn.Macro && len(paramTypes) >= 2 {
		owner = "#" + moduleName
		paramTypes = paramTypes[2:]
		paramNames = paramNames[2:]
	}
	params := make([]string, 0, len(paramTypes))
	for i, typ := range paramTypes {
		name := fmt.Sprintf("arg%d", i+1)
		if i < len(paramNames) && paramNames[i] != "" {
			name = paramNames[i]
		}
		params = append(params, fmt.Sprintf("%s: %s", name, displayType(typ)))
	}
	ret := displayType(fn.Return)
	if ret == "" {
		ret = string(checker.Void)
	}
	return fmt.Sprintf("%s.%s%s(%s) -> %s", owner, fn.Name, generics, strings.Join(params, ", "), ret)
}

func stdlibMemberSignature(info *checker.Info, receiver checker.Type, fn stdlib.Function) string {
	bindings := memberTypeBindings(receiver)
	owner := displayCheckerTypeOneLine(info, receiver)
	if owner == "" || owner == string(checker.Unknown) {
		owner = fn.Receiver
	}
	params := make([]string, 0, len(fn.Params))
	for i, typ := range fn.Params {
		name := fmt.Sprintf("arg%d", i+1)
		if i < len(fn.ParamNames) && fn.ParamNames[i] != "" {
			name = fn.ParamNames[i]
		}
		params = append(params, fmt.Sprintf("%s: %s", name, displayType(substituteStdlibType(typ, bindings))))
	}
	ret := displayType(substituteStdlibType(fn.Return, bindings))
	if ret == "" {
		ret = string(checker.Void)
	}
	return fmt.Sprintf("%s.%s%s(%s) -> %s", owner, fn.Name, formatSignatureGenerics(fn.Generics), strings.Join(params, ", "), ret)
}

func memberTypeBindings(receiver checker.Type) map[string]string {
	bindings := map[string]string{}
	if elem, ok := checker.ArrayElement(receiver); ok {
		bindings["T"] = string(elem)
		return bindings
	}
	base, args, ok := parseDisplayGenericName(string(receiver))
	if !ok {
		return bindings
	}
	switch base {
	case "Map", "WeakMap":
		if len(args) >= 2 {
			bindings["K"] = args[0]
			bindings["V"] = args[1]
		}
	case "Set", "WeakSet", "Iter":
		if len(args) >= 1 {
			bindings["T"] = args[0]
		}
	}
	return bindings
}

func substituteStdlibType(name string, bindings map[string]string) string {
	if bound, ok := bindings[name]; ok {
		return bound
	}
	if params, ret, ok := parseRawDisplayFuncType(name); ok {
		parts := make([]string, 0, len(params)+1)
		for _, param := range params {
			parts = append(parts, substituteStdlibType(param, bindings))
		}
		parts = append(parts, substituteStdlibType(ret, bindings))
		return "Func[" + strings.Join(parts, ",") + "]"
	}
	if params, ret, ok := parseRawDisplayAsyncFuncType(name); ok {
		parts := make([]string, 0, len(params)+1)
		for _, param := range params {
			parts = append(parts, substituteStdlibType(param, bindings))
		}
		parts = append(parts, substituteStdlibType(ret, bindings))
		return "AsyncFunc[" + strings.Join(parts, ",") + "]"
	}
	if strings.HasSuffix(name, "?") && name != "?" {
		return substituteStdlibType(strings.TrimSuffix(name, "?"), bindings) + "?"
	}
	if base, args, ok := parseDisplayGenericName(name); ok {
		for i, arg := range args {
			args[i] = substituteStdlibType(arg, bindings)
		}
		return base + "[" + strings.Join(args, ",") + "]"
	}
	return name
}

func parseDisplayGenericName(name string) (string, []string, bool) {
	idx := strings.IndexByte(name, '[')
	if idx <= 0 || !strings.HasSuffix(name, "]") {
		return "", nil, false
	}
	args := splitDisplayTypeList(strings.TrimSuffix(name[idx+1:], "]"))
	if len(args) == 0 {
		return "", nil, false
	}
	return name[:idx], args, true
}

func displayType(name string) string {
	params, ret, ok := parseDisplayFuncType(name)
	if ok {
		return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), ret)
	}
	params, ret, ok = parseDisplayAsyncFuncType(name)
	if ok {
		return "~ " + displayFuncType(params, ret)
	}
	return name
}

func displayCheckerType(info *checker.Info, typ checker.Type) string {
	return displayCheckerTypeIndent(info, typ, 0)
}

func displayCheckerTypeOneLine(info *checker.Info, typ checker.Type) string {
	return strings.Join(strings.Fields(displayCheckerType(info, typ)), " ")
}

func displayCheckerTypeIndent(info *checker.Info, typ checker.Type, indent int) string {
	name := string(typ)
	if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
		if info != nil {
			if structInfo := info.Types[name]; structInfo != nil {
				if len(structInfo.Fields) == 0 {
					return "{}"
				}
				currentIndent := strings.Repeat("  ", indent)
				fieldIndent := strings.Repeat("  ", indent+1)
				lines := []string{"{"}
				for _, field := range structInfo.Fields {
					lines = append(lines, fmt.Sprintf("%s%s: %s", fieldIndent, field.Name, displayCheckerTypeIndent(info, field.Type, indent+1)))
				}
				lines = append(lines, currentIndent+"}")
				return strings.Join(lines, "\n")
			}
		}
	}
	if params, ret, ok := parseRawDisplayFuncType(name); ok {
		for i, param := range params {
			params[i] = displayCheckerTypeIndent(info, checker.Type(param), indent)
		}
		return displayFuncType(params, displayCheckerTypeIndent(info, checker.Type(ret), indent))
	}
	if params, ret, ok := parseRawDisplayAsyncFuncType(name); ok {
		for i, param := range params {
			params[i] = displayCheckerTypeIndent(info, checker.Type(param), indent)
		}
		return "~ " + displayFuncType(params, displayCheckerTypeIndent(info, checker.Type(ret), indent))
	}
	return name
}

func displayFuncType(params []string, ret string) string {
	switch len(params) {
	case 0:
		return "() -> " + ret
	case 1:
		return params[0] + " -> " + ret
	default:
		return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), ret)
	}
}

func parseDisplayFuncType(name string) ([]string, string, bool) {
	params, ret, ok := parseRawDisplayFuncType(name)
	if !ok {
		return nil, "", false
	}
	for i, param := range params {
		params[i] = displayType(param)
	}
	return params, displayType(ret), true
}

func parseDisplayAsyncFuncType(name string) ([]string, string, bool) {
	params, ret, ok := parseRawDisplayAsyncFuncType(name)
	if !ok {
		return nil, "", false
	}
	for i, param := range params {
		params[i] = displayType(param)
	}
	return params, displayType(ret), true
}

func parseRawDisplayFuncType(name string) ([]string, string, bool) {
	if !strings.HasPrefix(name, "Func[") || !strings.HasSuffix(name, "]") {
		return nil, "", false
	}
	parts := splitDisplayTypeList(strings.TrimSuffix(strings.TrimPrefix(name, "Func["), "]"))
	if len(parts) == 0 {
		return nil, "", false
	}
	return parts[:len(parts)-1], parts[len(parts)-1], true
}

func parseRawDisplayAsyncFuncType(name string) ([]string, string, bool) {
	if !strings.HasPrefix(name, "AsyncFunc[") || !strings.HasSuffix(name, "]") {
		return nil, "", false
	}
	parts := splitDisplayTypeList(strings.TrimSuffix(strings.TrimPrefix(name, "AsyncFunc["), "]"))
	if len(parts) == 0 {
		return nil, "", false
	}
	return parts[:len(parts)-1], parts[len(parts)-1], true
}

func splitDisplayTypeList(src string) []string {
	var out []string
	depth := 0
	start := 0
	for i, ch := range src {
		switch ch {
		case '[', '{', '(':
			depth++
		case ']', '}', ')':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(src[start:i]))
				start = i + 1
			}
		}
	}
	out = append(out, strings.TrimSpace(src[start:]))
	return out
}
