package format

import (
	"strconv"

	"github.com/oboard/rune-lang/internal/ast"
)

type Options struct {
	Indent string
}

const defaultIndent = "  "

func File(file *ast.File) string {
	return FileWithOptions(file, Options{})
}

func FileWithOptions(file *ast.File, options Options) string {
	f := newFormatter(options)
	for _, imp := range file.Imports {
		f.linef("@%s", strconv.Quote(imp.Path))
	}
	if len(file.Imports) > 0 && len(file.GoImports) > 0 {
		f.line("")
	}
	for _, imp := range file.GoImports {
		f.linef("@go.import(%s)", strconv.Quote(imp.Path))
	}
	if (len(file.Imports) > 0 || len(file.GoImports) > 0) && (len(file.Types) > 0 || len(file.Enums) > 0 || len(file.Functions) > 0 || len(file.Tests) > 0) {
		f.line("")
	}
	for i, typ := range file.Types {
		if i > 0 {
			f.line("")
		}
		f.structType(typ)
	}
	if len(file.Types) > 0 && len(file.Enums) > 0 {
		f.line("")
	}
	for i, enum := range file.Enums {
		if i > 0 {
			f.line("")
		}
		f.enumType(enum)
	}
	if (len(file.Types) > 0 || len(file.Enums) > 0) && len(file.Functions) > 0 {
		f.line("")
	}
	for i, fn := range file.Functions {
		if i > 0 {
			f.line("")
		}
		f.function(fn)
	}
	if (len(file.Types) > 0 || len(file.Enums) > 0 || len(file.Functions) > 0) && len(file.Tests) > 0 {
		f.line("")
	}
	for i, test := range file.Tests {
		if i > 0 {
			f.line("")
		}
		f.test(test)
	}
	return f.b.String()
}

func Source(file *ast.File, source string) string {
	return SourceWithOptions(file, source, Options{})
}

func SourceWithOptions(file *ast.File, source string, options Options) string {
	indent := options.Indent
	if indent == "" {
		indent = defaultIndent
	}
	return preserveLineComments(source, FileWithOptions(file, options), indent)
}

func newFormatter(options Options) formatter {
	indent := options.Indent
	if indent == "" {
		indent = defaultIndent
	}
	return formatter{indentUnit: indent}
}
