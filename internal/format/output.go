package format

import (
	"fmt"
	"strings"
)

type formatter struct {
	b      strings.Builder
	indent int
}

func (f *formatter) line(s string) {
	if s != "" {
		f.b.WriteString(indentString(f.indent))
	}
	f.b.WriteString(s)
	f.b.WriteByte('\n')
}

func (f *formatter) linef(format string, args ...any) {
	f.line(fmt.Sprintf(format, args...))
}

func indentString(level int) string {
	return strings.Repeat("  ", level)
}
