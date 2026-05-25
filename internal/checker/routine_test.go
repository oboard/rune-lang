package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestRoutineResultUnwrapInfersResultReturn(t *testing.T) {
	src := `~ read() => {
  file := @fs.readFile("1.txt")?
  file
}

main() => {
  task := read()
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	read := info.Functions["read"]
	if read == nil {
		t.Fatalf("missing read function")
	}
	if got, want := DisplayType(read.Return), "Result[@io.Data, Error]"; got != want {
		t.Fatalf("read return = %q, want %q", got, want)
	}

	readBody := file.Functions[0].Body.(*ast.BlockExpr)
	let := readBody.Statements[0].(*ast.LetStmt)
	unwrap := let.Value.(*ast.ResultUnwrapExpr)
	fsCall := unwrap.Expr.(*ast.CallExpr)
	if !info.AsyncCalls[fsCall] || !info.AwaitCalls[fsCall] {
		t.Fatalf("@fs.readFile call async=%v await=%v, want both true", info.AsyncCalls[fsCall], info.AwaitCalls[fsCall])
	}
	if got, want := info.ExprTypes[unwrap], Data; got != want {
		t.Fatalf("unwrap type = %s, want %s", got, want)
	}

	mainBody := file.Functions[1].Body.(*ast.BlockExpr)
	taskLet := mainBody.Statements[0].(*ast.LetStmt)
	readCall := taskLet.Value.(*ast.CallExpr)
	if !info.AsyncCalls[readCall] {
		t.Fatalf("read call outside routine was not marked async")
	}
	if info.AwaitCalls[readCall] {
		t.Fatalf("read call outside routine was marked await")
	}
	if got, want := DisplayType(info.ExprTypes[readCall]), "Task[Result[@io.Data, Error]]"; got != want {
		t.Fatalf("read call type = %q, want %q", got, want)
	}
}

func TestResultUnwrapRequiresRoutine(t *testing.T) {
	file, parseErrs := parser.Parse(`readResult() -> Result[Int, Error] => Ok(1)

read() => {
  readResult()?
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	var found bool
	for _, diag := range diags {
		if strings.Contains(diag.Message, "operator '?' can only be used inside a routine") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("diagnostics = %#v, want routine-only result unwrap error", diags)
	}
}

func TestConstructorPatternBindsResultPayloads(t *testing.T) {
	src := `read(flag: Bool) -> Result[String, Error] => flag {
  true => Ok("Ada")
  false => Err(Error {
    code: 1
    message: "no"
    cause: null
  })
}

message(flag: Bool) => read(flag) {
  Ok(user) => user
  Err(e) => e.message
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	messageBody := file.Functions[1].Body.(*ast.MatchExpr)
	okExpr := messageBody.Branches[0].Expr.(*ast.Identifier)
	errExpr := messageBody.Branches[1].Expr.(*ast.SelectorExpr)
	if got, want := info.ExprTypes[okExpr], String; got != want {
		t.Fatalf("Ok binding type = %s, want %s", got, want)
	}
	if got, want := info.ExprTypes[errExpr.Receiver], Error; got != want {
		t.Fatalf("Err binding type = %s, want %s", got, want)
	}
}
