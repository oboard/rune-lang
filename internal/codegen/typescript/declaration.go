package tscodegen

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/ir"
)

func GenerateDeclarationsIR(file *ir.File) (string, error) {
	g := &generator{file: file}
	g.declarationPreamble()
	for _, typ := range file.Types {
		if typ.Private {
			continue
		}
		g.declareStructType(typ)
	}
	for _, enum := range file.Enums {
		if enum.Private {
			continue
		}
		g.declareEnumType(enum)
	}
	for _, constant := range file.Constants {
		if constant.Private {
			continue
		}
		g.declareConst(constant)
	}
	for _, fn := range file.Functions {
		if fn.Private {
			continue
		}
		g.declareFunction(fn)
	}
	g.declarationExports()
	return g.buf.String(), nil
}

func (g *generator) declarationPreamble() {
	g.line("type RuneResult<T, E> = { ok: true; value: T } | { ok: false; error: E };")
	g.line("type RuneError = { code: number; message: string; cause: RuneError | null };")
	g.line("type RuneIter<T> = { next: () => [T, boolean] };")
	g.line("type RuneFileStat = { size: number; isFile: boolean; isDirectory: boolean };")
	g.line("type RuneTCPConnection = { socket: unknown };")
	g.line("type RuneTCPListener = { server: unknown; address: string };")
	g.line("declare class RuneStringBuffer {}")
	g.line("declare class RuneBuffer {}")
	g.line("declare class RuneReader {}")
	g.line("declare class RuneWriter {}")
	g.line("")
}

func (g *generator) declareStructType(typ *ir.StructType) {
	g.linef("type %s%s = {", mangleIdent(typ.Name), tsGenerics(typ.Generics, nil))
	g.indent++
	for _, field := range typ.Fields {
		if field.Private {
			continue
		}
		g.linef("%s: %s;", tsPropertyName(field.Name), tsType(field.Type))
	}
	g.indent--
	g.line("};")
	g.line("")
}

func (g *generator) declareEnumType(enum *ir.EnumType) {
	name := mangleIdent(enum.Name)
	if enumHasPayload(enum) {
		g.linef("type %s%s = { tag: number; payload: any[] };", name, tsGenerics(enum.Generics, nil))
	} else {
		g.linef("type %s%s = number;", name, tsGenerics(enum.Generics, nil))
	}
	g.linef("declare const %s: {", name)
	g.indent++
	for i, member := range enum.Members {
		if member.Private {
			continue
		}
		value := i
		if member.HasValue {
			value = member.Value
		}
		g.linef("readonly %s: %d;", tsPropertyName(member.Name), value)
	}
	g.indent--
	g.line("};")
	g.line("")
}

func (g *generator) declareConst(constant *ir.ConstDecl) {
	g.linef("declare const %s: %s;", mangleIdent(constant.Name), tsType(constant.Type))
}

func (g *generator) declareFunction(fn *ir.Function) {
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s: %s", mangleIdent(param.Name), tsType(param.Type)))
	}
	ret := tsType(fn.Return)
	if fn.Routine {
		ret = "Promise<" + ret + ">"
	}
	g.linef("declare function %s%s(%s): %s;", FunctionSymbolName(fn), tsGenerics(fn.Generics, fn.GenericConstraints), strings.Join(params, ", "), ret)
}

func (g *generator) declarationExports() {
	if len(g.file.Types)+len(g.file.Enums)+len(g.file.Constants)+len(g.file.Functions) == 0 {
		return
	}
	g.line("")
	for _, typ := range g.file.Types {
		if typ.Private {
			continue
		}
		g.exportTypeAlias(typ.Name, mangleIdent(typ.Name))
	}
	for _, enum := range g.file.Enums {
		if enum.Private {
			continue
		}
		g.exportTypeAlias(enum.Name, mangleIdent(enum.Name))
		g.exportValueAlias(enum.Name, mangleIdent(enum.Name))
	}
	for _, constant := range g.file.Constants {
		if constant.Private {
			continue
		}
		g.exportValueAlias(constant.Name, mangleIdent(constant.Name))
	}
	for _, fn := range g.file.Functions {
		if fn.Private {
			continue
		}
		g.exportValueAlias(fn.SourceName, FunctionSymbolName(fn))
	}
}

func (g *generator) exportTypeAlias(name string, local string) {
	if tsCanUseBareProperty(name) {
		g.linef("export type %s = %s;", name, local)
		return
	}
	g.linef("export type { %s as %s };", local, tsExportName(name))
}

func (g *generator) exportValueAlias(name string, local string) {
	if tsCanUseBareProperty(name) {
		g.linef("export declare const %s: typeof %s;", name, local)
		return
	}
	g.linef("export { %s as %s };", local, tsExportName(name))
}
