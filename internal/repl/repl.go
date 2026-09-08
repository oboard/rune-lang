package repl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/selfhostrunner"
)

type analyzeSourceFunc func(string, string) (*compiler.Program, []compiler.Diagnostic)
type selfhostCheckSourceFunc func(string, string) SelfhostCompileResult

var analyzeSource = compiler.AnalyzeSource
var selfhostCheckSource selfhostCheckSourceFunc

type SelfhostCompileResult struct {
	Ok     bool
	Output string
	Errors []string
}

const replFunctionName = "__rune_repl"

type Session struct {
	decls []string
	stmts []string
	out   io.Writer
}

func NewSession(out io.Writer) *Session {
	return &Session{out: out}
}

func Serve(in io.Reader, out io.Writer) error {
	session := NewSession(out)
	reader := bufio.NewReader(in)
	fmt.Fprintln(out, "Rune REPL")
	fmt.Fprintln(out, "Type .exit or press Ctrl-D to exit.")

	var form strings.Builder
	for {
		if form.Len() == 0 {
			fmt.Fprint(out, "rune> ")
		} else {
			fmt.Fprint(out, "...   ")
		}
		line, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) && line == "" {
			fmt.Fprintln(out)
			return nil
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		form.WriteString(line)
		src := strings.TrimSpace(form.String())
		if src == ".exit" || src == ".quit" {
			return nil
		}
		if src == "" {
			form.Reset()
			if errors.Is(err, io.EOF) {
				return nil
			}
			continue
		}
		if isIncomplete(src) && !errors.Is(err, io.EOF) {
			continue
		}
		if evalErr := session.Eval(src); evalErr != nil {
			fmt.Fprintf(out, "error: %s\n", evalErr)
		}
		form.Reset()
		if errors.Is(err, io.EOF) {
			return nil
		}
	}
}

func RegisterSelfhostAnalyzer(check selfhostCheckSourceFunc, analyze analyzeSourceFunc) {
	selfhostCheckSource = check
	if analyze != nil {
		analyzeSource = analyze
	}
}

func (s *Session) Eval(src string) error {
	if isDeclaration(src) {
		return s.addDeclaration(src)
	}
	return s.evalStatement(src)
}

func (s *Session) addDeclaration(src string) error {
	s.decls = append(s.decls, src)
	_, diags := analyzeSource("<repl>", s.source())
	if len(diags) > 0 {
		s.decls = s.decls[:len(s.decls)-1]
		printDiagnostics(s.out, diags)
		return fmt.Errorf("declaration rejected")
	}
	return nil
}

func (s *Session) evalStatement(src string) error {
	s.stmts = append(s.stmts, src)
	prog, diags := analyzeSource("<repl>", s.source())
	if len(diags) > 0 {
		s.stmts = s.stmts[:len(s.stmts)-1]
		printDiagnostics(s.out, diags)
		return fmt.Errorf("statement rejected")
	}
	result := selfhostrunner.RunFunctionIR(prog.IR, replFunctionName, nil)
	if result.Err != nil {
		s.stmts = s.stmts[:len(s.stmts)-1]
		return result.Err
	}
	if result.Value.Kind != "Void" {
		fmt.Fprintln(s.out, formatValue(result.Value))
	}
	return nil
}

func formatValue(value selfhostrunner.Value) string {
	switch value.Kind {
	case "Void":
		return "void"
	case "Null":
		return "null"
	case "Bool":
		return strconv.FormatBool(value.Bool)
	case "Int":
		return strconv.Itoa(value.Int)
	case "Double":
		return strconv.FormatFloat(value.Double, 'f', -1, 64)
	case "String":
		return strconv.Quote(value.Text)
	case "Char":
		return strconv.Quote(value.Char)
	case "Array":
		return formatValueList("[", "]", value.Values)
	case "Tuple":
		return formatValueList("(", ")", value.Values)
	}
	return value.Kind
}

func formatValueList(open string, close string, values []selfhostrunner.Value) string {
	parts := make([]string, 0, len(values))
	for _, item := range values {
		parts = append(parts, formatValue(item))
	}
	return open + strings.Join(parts, ", ") + close
}

func (s *Session) source() string {
	var b strings.Builder
	for _, decl := range s.decls {
		b.WriteString(decl)
		b.WriteString("\n\n")
	}
	b.WriteString("+ ")
	b.WriteString(replFunctionName)
	b.WriteString("() => {\n")
	for _, stmt := range s.stmts {
		b.WriteString(stmt)
		b.WriteByte('\n')
	}
	b.WriteString("}\n")
	return b.String()
}

func isDeclaration(src string) bool {
	if selfhostCheckSource != nil {
		checked := selfhostCheckSource(src, "<repl>")
		if !checked.Ok {
			return false
		}
	}
	prog, diags := analyzeSource("<repl>", src)
	return len(diags) == 0 && prog != nil && (len(prog.File.GoImports) > 0 || len(prog.File.Types) > 0 || len(prog.File.Functions) > 0)
}

func isIncomplete(src string) bool {
	if strings.HasSuffix(src, "=>") || strings.HasSuffix(src, "{") {
		return true
	}
	var paren, brace, bracket int
	inString := false
	escaped := false
	for _, ch := range src {
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '(':
			paren++
		case ')':
			paren--
		case '{':
			brace++
		case '}':
			brace--
		case '[':
			bracket++
		case ']':
			bracket--
		}
	}
	return paren > 0 || brace > 0 || bracket > 0
}

func printDiagnostics(out io.Writer, diags []compiler.Diagnostic) {
	for _, diag := range diags {
		if diag.Pos.Line > 0 {
			fmt.Fprintf(out, "<repl>:%d:%d: %s\n", diag.Pos.Line, diag.Pos.Column, diag.Message)
		} else {
			fmt.Fprintf(out, "<repl>: %s\n", diag.Message)
		}
	}
}
