package checker

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (c *checker) checkMacros(file *ast.File) {
	for _, typ := range file.Types {
		c.checkAnnotationList(typ.Annotations, typ.SourcePath)
		for i := range typ.Fields {
			c.checkAnnotationList(typ.Fields[i].Annotations, typ.SourcePath)
		}
		for _, method := range typ.Methods {
			if method.Macro {
				c.errorf(method.NamePos, "macro declarations must be top-level functions")
			}
			c.checkAnnotationList(method.Annotations, method.SourcePath)
		}
	}
	for _, enum := range file.Enums {
		c.checkAnnotationList(enum.Annotations, enum.SourcePath)
		for i := range enum.Members {
			c.checkAnnotationList(enum.Members[i].Annotations, enum.SourcePath)
		}
	}
	for _, fn := range file.Functions {
		if fn.Macro && !isLocalSyntaxMacro(fn) {
			c.errorf(
				fn.NamePos,
				"macro %s must accept SyntaxFile and MacroContext first and return SyntaxFile",
				fn.Name,
			)
		}
		c.checkAnnotationList(fn.Annotations, fn.SourcePath)
	}
}

func (c *checker) checkAnnotationList(annotations []ast.Annotation, sourcePath string) {
	c.withSourcePath(sourcePath, func() {
		for i := range annotations {
			c.checkAnnotation(&annotations[i])
		}
	})
}

func (c *checker) checkAnnotation(annotation *ast.Annotation) {
	if annotation.Module == "" && annotation.Name == "alias" {
		return
	}
	if annotation.Module == "" {
		fn, ok := c.resolveFunction(annotation.Name, annotation.NamePos)
		if !ok || fn == nil {
			c.errorf(annotation.NamePos, "unknown macro #%s", annotation.Name)
			return
		}
		if !fn.Macro {
			c.errorf(annotation.NamePos, "#%s refers to a function that is not a macro", annotation.Name)
			return
		}
		c.info.ResolvedMacroFunctions[annotation] = fn
		c.checkMacroFunctionArgs(annotation, fn)
		return
	}
	if c.info.Stdlib == nil {
		c.errorf(annotation.NamePos, "unknown macro #%s.%s", annotation.Module, annotation.Name)
		return
	}
	fn, ok := c.info.Stdlib.MacroFunction(annotation.Module, annotation.Name)
	if !ok {
		if ordinary, exists := c.info.Stdlib.Function(annotation.Module, annotation.Name); exists && !ordinary.Macro {
			c.errorf(annotation.NamePos, "#%s.%s refers to a function that is not a macro", annotation.Module, annotation.Name)
			return
		}
		c.errorf(annotation.NamePos, "unknown macro #%s.%s", annotation.Module, annotation.Name)
		return
	}
	c.info.ResolvedMacros[annotation] = fn
	if message := c.stdlibMacroPurityError(fn); message != "" {
		c.errorf(annotation.NamePos, "macro #%s.%s is not pure: %s", annotation.Module, annotation.Name, message)
		return
	}
	argTypes := make([]Type, 0, len(annotation.Args))
	for _, arg := range annotation.Args {
		argTypes = append(argTypes, c.inferExpr(arg, map[string]Type{}))
	}
	bindings := c.stdlibTypeBindings(fn)
	if isStdlibSyntaxMacro(fn) {
		copy := *fn
		copy.ParamNames = append([]string(nil), fn.ParamNames[2:]...)
		copy.Params = append([]string(nil), fn.Params[2:]...)
		fn = &copy
	}
	c.checkStdlibGenericArgs(annotation.Module, annotation.Name, fn, annotation.Args, argTypes, bindings, map[string]Type{}, annotation.Pos)
}

func (c *checker) checkMacroFunctionArgs(annotation *ast.Annotation, fn *FuncInfo) {
	argTypes := make([]Type, 0, len(annotation.Args))
	for _, arg := range annotation.Args {
		argTypes = append(argTypes, c.inferExpr(arg, map[string]Type{}))
	}
	params := fn.Params
	if isLocalSyntaxMacro(fn.Node) {
		params = params[2:]
	}
	c.checkArgs(fn.Name, params, annotation.Args, argTypes, annotation.Pos)
}

func isLocalSyntaxMacro(fn *ast.Function) bool {
	return fn != nil &&
		len(fn.Params) >= 2 &&
		fn.Params[0].Type.Canonical() == "SyntaxFile" &&
		fn.Params[1].Type.Canonical() == "MacroContext" &&
		fn.ReturnType.Canonical() == "SyntaxFile"
}

func isStdlibSyntaxMacro(fn *stdlib.Function) bool {
	return fn != nil &&
		len(fn.Params) >= 2 &&
		fn.Params[0] == "SyntaxFile" &&
		fn.Params[1] == "MacroContext" &&
		fn.Return == "SyntaxFile"
}
