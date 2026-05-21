package repl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/interpreter"
	"github.com/oboard/rune-lang/internal/ir"
)

const replFunctionName = "__rune_repl"

type Session struct {
	decls  []string
	stmts  []string
	interp *interpreter.Interpreter
	out    io.Writer
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

func (s *Session) Eval(src string) error {
	if isDeclaration(src) {
		return s.addDeclaration(src)
	}
	return s.evalStatement(src)
}

func (s *Session) addDeclaration(src string) error {
	s.decls = append(s.decls, src)
	prog, diags := compiler.AnalyzeSource("<repl>", s.source())
	if len(diags) > 0 {
		s.decls = s.decls[:len(s.decls)-1]
		printDiagnostics(s.out, diags)
		return fmt.Errorf("declaration rejected")
	}
	s.load(prog.IR)
	return nil
}

func (s *Session) evalStatement(src string) error {
	s.stmts = append(s.stmts, src)
	prog, diags := compiler.AnalyzeSource("<repl>", s.source())
	if len(diags) > 0 {
		s.stmts = s.stmts[:len(s.stmts)-1]
		printDiagnostics(s.out, diags)
		return fmt.Errorf("statement rejected")
	}
	s.load(prog.IR)
	stmt, ok := currentStatement(prog.IR)
	if !ok {
		return fmt.Errorf("internal repl statement not found")
	}
	value, show, err := s.interp.Exec(stmt)
	if err != nil {
		s.stmts = s.stmts[:len(s.stmts)-1]
		return err
	}
	if show {
		fmt.Fprintln(s.out, interpreter.Format(value))
	}
	return nil
}

func (s *Session) load(file *ir.File) {
	if s.interp == nil {
		s.interp = interpreter.New(file, interpreter.WithOutput(s.out))
		return
	}
	s.interp.Load(file)
}

func (s *Session) source() string {
	var b strings.Builder
	for _, decl := range s.decls {
		b.WriteString(decl)
		b.WriteString("\n\n")
	}
	b.WriteString(replFunctionName)
	b.WriteString("() => {\n")
	for _, stmt := range s.stmts {
		b.WriteString(stmt)
		b.WriteByte('\n')
	}
	b.WriteString("}\n")
	return b.String()
}

func currentStatement(file *ir.File) (ir.Stmt, bool) {
	for _, fn := range file.Functions {
		if fn.Name != replFunctionName {
			continue
		}
		block, ok := fn.Body.(*ir.BlockExpr)
		if !ok || len(block.Statements) == 0 {
			return nil, false
		}
		return block.Statements[len(block.Statements)-1], true
	}
	return nil, false
}

func isDeclaration(src string) bool {
	prog, diags := compiler.AnalyzeSource("<repl>", src)
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
