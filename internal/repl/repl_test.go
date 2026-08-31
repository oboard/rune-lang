package repl

import (
	"bytes"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/compiler"
)

func TestSessionEvalPersistsBindings(t *testing.T) {
	var out bytes.Buffer
	session := NewSession(&out)
	if err := session.Eval("x := 40"); err != nil {
		t.Fatalf("Eval let: %v", err)
	}
	if err := session.Eval("x + 2"); err != nil {
		t.Fatalf("Eval expr: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "42" {
		t.Fatalf("output = %q, want 42", got)
	}
}

func TestSessionEvalFunctionAndArray(t *testing.T) {
	var out bytes.Buffer
	session := NewSession(&out)
	if err := session.Eval(`fib(n: Int) -> Int => {
  0 => 0
  1 => 1
  _ => fib(n - 1) + fib(n - 2)
}`); err != nil {
		t.Fatalf("Eval function: %v", err)
	}
	if err := session.Eval("arr := [fib(5), fib(6)]"); err != nil {
		t.Fatalf("Eval array: %v", err)
	}
	if err := session.Eval("arr.push(fib(7))"); err != nil {
		t.Fatalf("Eval push: %v", err)
	}
	if err := session.Eval("arr.length()"); err != nil {
		t.Fatalf("Eval length: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "3\n3" {
		t.Fatalf("output = %q, want push length and array length", got)
	}
}

func TestSessionEvalUsesAnalyzeOverride(t *testing.T) {
	var out bytes.Buffer
	session := NewSession(&out)
	called := 0
	prevAnalyze := analyzeSource
	prevCheck := selfhostCheckSource
	analyzeSource = func(path string, text string) (*compiler.Program, []compiler.Diagnostic) {
		called++
		if path != "<repl>" {
			t.Fatalf("analyze path = %q, want <repl>", path)
		}
		if called == 1 {
			if text != "x := 1" {
				t.Fatalf("first analyze text = %q, want declaration probe", text)
			}
		} else if called == 2 {
			if !strings.Contains(text, "x := 1") || !strings.Contains(text, "+ __rune_repl() => {") {
				t.Fatalf("second analyze text = %q, want wrapped repl source", text)
			}
		} else if called == 3 {
			if text != "x + 1" {
				t.Fatalf("third analyze text = %q, want statement declaration probe", text)
			}
		} else if !strings.Contains(text, "x := 1") || !strings.Contains(text, "x + 1") {
			t.Fatalf("analyze text = %q, want accumulated repl source", text)
		}
		return prevAnalyze(path, text)
	}
	selfhostCheckSource = func(source string, path string) SelfhostCompileResult {
		if path != "<repl>" {
			t.Fatalf("check path = %q, want <repl>", path)
		}
		return SelfhostCompileResult{Ok: true}
	}
	defer func() {
		analyzeSource = prevAnalyze
		selfhostCheckSource = prevCheck
	}()

	if err := session.Eval("x := 1"); err != nil {
		t.Fatalf("Eval declaration: %v", err)
	}
	if err := session.Eval("x + 1"); err != nil {
		t.Fatalf("Eval expression: %v", err)
	}
	if called < 2 {
		t.Fatalf("analyze override calls = %d, want at least 2", called)
	}
}

func TestIsDeclarationUsesSelfhostCheckOverride(t *testing.T) {
	prevCheck := selfhostCheckSource
	defer func() { selfhostCheckSource = prevCheck }()
	calls := 0
	selfhostCheckSource = func(source string, path string) SelfhostCompileResult {
		calls++
		if source != "broken declaration" {
			t.Fatalf("check source = %q, want broken declaration", source)
		}
		if path != "<repl>" {
			t.Fatalf("check path = %q, want <repl>", path)
		}
		return SelfhostCompileResult{Ok: false, Errors: []string{"expected declaration"}}
	}
	if isDeclaration("broken declaration") {
		t.Fatal("isDeclaration() = true, want false when selfhost check fails")
	}
	if calls != 1 {
		t.Fatalf("selfhost check calls = %d, want 1", calls)
	}
}
