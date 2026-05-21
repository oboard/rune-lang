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
		for i := 0; i < f.indent; i++ {
			f.b.WriteString("  ")
		}
	}
	f.b.WriteString(s)
	f.b.WriteByte('\n')
}

func (f *formatter) linef(format string, args ...any) {
	f.line(fmt.Sprintf(format, args...))
}
