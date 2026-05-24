package format

import (
	"fmt"
	"strings"
)

type formatter struct {
	b          strings.Builder
	indent     int
	indentUnit string
}

func (f *formatter) line(s string) {
	if s != "" {
		f.b.WriteString(f.indentString(f.indent))
	}
	f.b.WriteString(s)
	f.b.WriteByte('\n')
}

func (f *formatter) linef(format string, args ...any) {
	f.line(fmt.Sprintf(format, args...))
}

func (f *formatter) indentString(level int) string {
	return strings.Repeat(f.indentUnit, level)
}
