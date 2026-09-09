package ir

import (
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/lexer"
)

// LoopExpr is a statement-level loop that never yields a value; its body runs
// until a ReturnStmt (or, for Void functions, an explicit continuation)
// transfers control elsewhere. It is introduced by self-tail-call elimination
// to lower self-recursive functions into iterative loops. The loop carries no
// condition because it is always exited via return/continue inside its body.
type LoopExpr struct {
	ExprBase
	Body *BlockExpr
}

func (*LoopExpr) exprNode() {}

// IfStmt models a statement-level conditional with explicit then/else bodies.
type IfStmt struct {
	Cond Expr
	Then []Stmt
	Else []Stmt
	Pos  lexer.Position
}

func (*IfStmt) stmtNode() {}
func (s *IfStmt) Position() lexer.Position {
	return s.Pos
}

// ContinueStmt transfers control back to the top of the nearest enclosing loop.
type ContinueStmt struct {
	Pos lexer.Position
}

func (*ContinueStmt) stmtNode() {}
func (s *ContinueStmt) Position() lexer.Position {
	return s.Pos
}

// ReturnStmt returns a value from the enclosing function.
type ReturnStmt struct {
	Value Expr
	Pos   lexer.Position
}

func (*ReturnStmt) stmtNode() {}
func (s *ReturnStmt) Position() lexer.Position {
	return s.Pos
}

// MultiAssignStmt assigns all values before writing any target, preserving the
// argument evaluation order of a self tail call when parameters depend on one
// another (for example f(text, index + 1, text.at(index))).
type MultiAssignStmt struct {
	Names  []string
	Values []Expr
	Pos    lexer.Position
}

func (*MultiAssignStmt) stmtNode() {}
func (s *MultiAssignStmt) Position() lexer.Position {
	return s.Pos
}

// EliminateSelfTailCalls rewrites a function that is purely self-tail-recursive
// into an iterative loop. It returns true only when the rewrite was applied and
// fn.Body was replaced with a *LoopExpr. The transformation is conservative:
// it applies only when every call to the function's own name in its body is in
// tail position, preserving observable behavior while removing recursion.
func EliminateSelfTailCalls(fn *Function) bool {
	if fn == nil || fn.Body == nil || fn.Routine || len(fn.Generics) > 0 || len(fn.Params) == 0 || fn.ReceiverType != "" {
		return false
	}
	if !containsSelfCall(fn.Name, fn.Body) {
		return false
	}
	stmts, ok := buildTailLoop(fn, fn.Body)
	if !ok {
		return false
	}
	pos := fn.Body.Position()
	fn.Body = &LoopExpr{
		ExprBase: ExprBase{Pos: pos, Type: fn.Return},
		Body:     &BlockExpr{ExprBase: ExprBase{Pos: pos, Type: fn.Return}, Statements: stmts},
	}
	return true
}

// containsSelfCall reports whether expr contains a direct call to fnName
// anywhere in its subtree.
func containsSelfCall(fnName string, expr Expr) bool {
	found := false
	WalkExpr(expr, func(e Expr) {
		if call, ok := e.(*CallExpr); ok {
			if id, ok := call.Callee.(*Identifier); ok && id.Name == fnName {
				found = true
			}
		}
	})
	return found
}

// buildTailLoop lowers a tail-context expression into a flat sequence of
// statements, turning self-calls into parameter reassignments plus continue. It
// returns ok=false when it encounters a self-call in a non-tail position
// (making the whole transformation unsafe).
func buildTailLoop(fn *Function, expr Expr) ([]Stmt, bool) {
	switch e := expr.(type) {
	case *TernaryExpr:
		if e.Alternative == nil {
			return nil, false
		}
		if containsSelfCall(fn.Name, e.Condition) {
			return nil, false
		}
		thenStmts, ok := buildTailLoop(fn, e.Consequence)
		if !ok {
			return nil, false
		}
		elseStmts, ok := buildTailLoop(fn, e.Alternative)
		if !ok {
			return nil, false
		}
		return []Stmt{&IfStmt{Cond: e.Condition, Then: thenStmts, Else: elseStmts, Pos: e.Pos}}, true
	case *BlockExpr:
		// Statements before the last are effects; the final statement carries
		// the tail value. Only lower when the tail position is a bare self-call,
		// which is the shape produced by recursive scan loops in the self-host
		// compiler; anything else is left to the normal block emitter.
		if len(e.Statements) == 0 {
			return nil, false
		}
		lastIndex := len(e.Statements) - 1
		var out []Stmt
		for i := 0; i < lastIndex; i++ {
			stmt := e.Statements[i]
			if stmtContainsSelfCall(fn.Name, stmt) {
				return nil, false
			}
			out = append(out, stmt)
		}
		last := e.Statements[lastIndex]
		value := stmtValue(last)
		if value == nil {
			return nil, false
		}
		if call, ok := value.(*CallExpr); ok && isSelfCall(fn, call) {
			tail, ok := buildTailLoop(fn, call)
			if !ok {
				return nil, false
			}
			out = append(out, tail...)
			return out, true
		}
		if containsSelfCall(fn.Name, value) {
			return nil, false
		}
		out = append(out, &ReturnStmt{Value: value, Pos: value.Position()})
		return out, true
	case *CallExpr:
		if !isSelfCall(fn, e) {
			if containsSelfCall(fn.Name, e) {
				return nil, false
			}
			return []Stmt{&ReturnStmt{Value: e, Pos: e.Pos}}, true
		}
		names := make([]string, len(fn.Params))
		for i, param := range fn.Params {
			names[i] = param.Name
		}
		return []Stmt{
			&MultiAssignStmt{Names: names, Values: e.Args, Pos: e.Pos},
			&ContinueStmt{Pos: e.Pos},
		}, true
	default:
		if containsSelfCall(fn.Name, expr) {
			return nil, false
		}
		return []Stmt{&ReturnStmt{Value: expr, Pos: expr.Position()}}, true
	}
}

// isSelfCall reports whether call is a direct call to fn's own name with the
// matching arity.
func isSelfCall(fn *Function, call *CallExpr) bool {
	id, ok := call.Callee.(*Identifier)
	if !ok || id.Name != fn.Name {
		return false
	}
	return len(call.Args) == len(fn.Params)
}

// stmtValue extracts the value expression carried by a statement used as the
// final tail position of a block, or nil when the statement does not carry a value.
func stmtValue(stmt Stmt) Expr {
	switch s := stmt.(type) {
	case *ExprStmt:
		return s.Expr
	case *ReturnStmt:
		return s.Value
	case *LetStmt:
		return nil
	default:
		return nil
	}
}

// stmtContainsSelfCall reports whether a statement's subtree contains a self-call.
func stmtContainsSelfCall(fnName string, stmt Stmt) bool {
	found := false
	WalkStmt(stmt, func(e Expr) {
		if call, ok := e.(*CallExpr); ok {
			if id, ok := call.Callee.(*Identifier); ok && id.Name == fnName {
				found = true
			}
		}
	})
	return found
}

// checker is referenced to keep the import stable for type usage in this file.
var _ = checker.Unknown
