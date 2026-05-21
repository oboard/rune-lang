package format

import (
	"strconv"

	"github.com/oboard/rune-lang/internal/ast"
)

func File(file *ast.File) string {
	f := formatter{}
	for _, imp := range file.GoImports {
		f.linef("@go.import(%s)", strconv.Quote(imp.Path))
	}
	if len(file.GoImports) > 0 && (len(file.Types) > 0 || len(file.Functions) > 0) {
		f.line("")
	}
	for i, typ := range file.Types {
		if i > 0 {
			f.line("")
		}
		f.structType(typ)
	}
	if len(file.Types) > 0 && len(file.Functions) > 0 {
		f.line("")
	}
	for i, fn := range file.Functions {
		if i > 0 {
			f.line("")
		}
		f.function(fn)
	}
	return f.b.String()
}
