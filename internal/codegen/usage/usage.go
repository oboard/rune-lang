package usage

import (
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

type Usage struct {
	types      []checker.Type
	intrinsics []string

	AsyncCall    bool
	ResultUnwrap bool
	Routine      bool
	Signal       bool
	GoFFI        bool
	Template     bool
}

func Collect(file *ir.File) Usage {
	var usage Usage
	for _, fn := range file.Functions {
		usage.collectFunction(file, fn)
	}
	for _, test := range file.Tests {
		usage.collectExpr(file, test.Body)
	}
	for _, typ := range file.Types {
		for _, field := range typ.Fields {
			usage.addType(field.Type)
		}
		for _, method := range typ.Methods {
			usage.collectFunction(file, method)
		}
	}
	return usage
}

func (u *Usage) collectFunction(file *ir.File, fn *ir.Function) {
	if fn.Routine {
		u.Routine = true
	}
	u.addType(fn.Return)
	for _, param := range fn.Params {
		u.addType(param.Type)
	}
	u.collectExpr(file, fn.Body)
}

func (u *Usage) collectExpr(file *ir.File, expr ir.Expr) {
	ir.WalkExpr(expr, func(expr ir.Expr) {
		u.addType(expr.ResultType())
		switch e := expr.(type) {
		case *ir.CallExpr:
			if e.Async {
				u.AsyncCall = true
			}
			u.collectCallIntrinsic(file, e)
		case *ir.ResultUnwrapExpr:
			u.ResultUnwrap = true
		case *ir.TemplateLiteral:
			u.Template = true
		case *ir.SelectorExpr:
			if at, ok := e.Receiver.(*ir.AtExpr); ok && at.Name == "go" {
				u.GoFFI = true
			}
		case *ir.WatchExpr, *ir.ReactiveLiteral:
			u.Signal = true
		case *ir.BlockExpr:
			for _, stmt := range e.Statements {
				if let, ok := stmt.(*ir.LetStmt); ok && let.Signal {
					u.Signal = true
				}
			}
		}
	})
}

func (u *Usage) collectCallIntrinsic(file *ir.File, call *ir.CallExpr) {
	if file == nil || file.Stdlib == nil {
		return
	}
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return
	}
	if at, ok := sel.Receiver.(*ir.AtExpr); ok {
		if fn, ok := file.Stdlib.Function(at.Name, sel.Name); ok {
			u.addIntrinsic(fn.Intrinsic)
		}
		return
	}
	moduleName, receiverName, ok := checker.StdlibReceiverModule(sel.Receiver.ResultType())
	if !ok {
		return
	}
	if fn, ok := file.Stdlib.ReceiverFunction(moduleName, receiverName, sel.Name); ok {
		u.addIntrinsic(fn.Intrinsic)
	}
}

func (u *Usage) addType(typ checker.Type) {
	if typ == "" || typ == checker.Unknown {
		return
	}
	u.types = append(u.types, typ)
}

func (u *Usage) addIntrinsic(intrinsic string) {
	if intrinsic == "" {
		return
	}
	u.intrinsics = append(u.intrinsics, intrinsic)
}

func (u Usage) HasType(typ checker.Type) bool {
	for _, candidate := range u.types {
		if typeContains(candidate, typ) {
			return true
		}
	}
	return false
}

func (u Usage) HasGeneric(base string) bool {
	for _, candidate := range u.types {
		if typeUsesGeneric(candidate, base) {
			return true
		}
	}
	return false
}

func (u Usage) HasIntrinsicPrefix(prefix string) bool {
	for _, intrinsic := range u.intrinsics {
		if strings.HasPrefix(intrinsic, prefix) {
			return true
		}
	}
	return false
}

func (u Usage) UsesAsyncRuntime() bool {
	return u.Routine || u.AsyncCall || u.ResultUnwrap
}

func typeContains(candidate checker.Type, typ checker.Type) bool {
	if candidate == typ {
		return true
	}
	return strings.Contains(string(candidate), string(typ))
}

func typeUsesGeneric(candidate checker.Type, base string) bool {
	name := string(candidate)
	return strings.HasPrefix(name, base+"[") || strings.Contains(name, ","+base+"[") || strings.Contains(name, "["+base+"[")
}
