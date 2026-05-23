package repl

import (
	"bytes"
	"strings"
	"testing"
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
