package checker

import (
	"fmt"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func Lint(file *ast.File, info *Info) []Diagnostic {
	if file == nil || info == nil {
		return nil
	}
	l := &linter{
		file:                   file,
		info:                   info,
		usedFunctions:          map[*FuncInfo]bool{},
		usedFields:             map[string]bool{},
		constructedEnumMembers: map[string]bool{},
	}
	l.run()
	return l.diags
}

func LintErrors(file *ast.File, info *Info) []Diagnostic {
	return lintErrorsForSourcePath(file, info, "")
}

func lintErrorsForSourcePath(file *ast.File, info *Info, sourcePath string) []Diagnostic {
	if file == nil || info == nil {
		return nil
	}
	l := &linter{
		file:                   file,
		info:                   info,
		sourcePath:             normalizeSourcePath(sourcePath),
		usedFunctions:          map[*FuncInfo]bool{},
		usedFields:             map[string]bool{},
		constructedEnumMembers: map[string]bool{},
	}
	l.run()
	diags := l.diags
	if len(diags) == 0 {
		return nil
	}
	var out []Diagnostic
	for _, diag := range diags {
		if diag.Severity != SeverityWarning {
			out = append(out, diag)
		}
	}
	return out
}

func LintErrorsForSourcePath(file *ast.File, info *Info, sourcePath string) []Diagnostic {
	return lintErrorsForSourcePath(file, info, sourcePath)
}

type linter struct {
	file                   *ast.File
	info                   *Info
	diags                  []Diagnostic
	usedFunctions          map[*FuncInfo]bool
	usedFields             map[string]bool
	constructedEnumMembers map[string]bool
	currentFunction        *ast.Function
	currentSourcePath      string
	sourcePath             string
}

func (l *linter) run() {
	for _, typ := range l.file.Types {
		for _, method := range typ.Methods {
			l.visitFunction(method)
		}
	}
	for _, fn := range l.file.Functions {
		l.visitFunction(fn)
	}
	for _, test := range l.file.Tests {
		prevSourcePath := l.currentSourcePath
		l.currentSourcePath = normalizeSourcePath(test.SourcePath)
		l.visitExpr(test.Body)
		l.currentSourcePath = prevSourcePath
	}
	l.warnUnusedFunctions()
	l.warnUnusedFields()
	l.warnUnusedConstructors()
}

func (l *linter) visitFunction(fn *ast.Function) {
	prev := l.currentFunction
	prevSourcePath := l.currentSourcePath
	l.currentFunction = fn
	l.currentSourcePath = normalizeSourcePath(fn.SourcePath)
	if len(fn.Params) == 1 {
		if block, ok := fn.Body.(*ast.PatternBlock); ok {
			l.lintPatternBranches(block.Branches, Type(fn.Params[0].Type.Canonical()))
		}
	}
	l.visitExpr(fn.Body)
	l.currentFunction = prev
	l.currentSourcePath = prevSourcePath
}

func (l *linter) visitExpr(expr ast.Expr) {
	if expr == nil {
		return
	}
	l.visitExprNode(expr)
	switch e := expr.(type) {
	case *ast.Identifier, *ast.AtExpr, *ast.ThisExpr, *ast.IntegerLiteral, *ast.DoubleLiteral, *ast.BigIntLiteral,
		*ast.StringLiteral, *ast.CharLiteral, *ast.RegexLiteral, *ast.BoolLiteral, *ast.NullLiteral:
	case *ast.TemplateLiteral:
		for _, part := range e.Parts {
			l.visitExpr(part.Expr)
		}
	case *ast.UnaryExpr:
		l.visitExpr(e.Expr)
	case *ast.PostfixExpr:
		l.visitExpr(e.Expr)
	case *ast.ResultUnwrapExpr:
		l.visitExpr(e.Expr)
	case *ast.CompileTimeExpr:
		l.visitExpr(e.Expr)
	case *ast.BinaryExpr:
		l.visitExpr(e.Left)
		l.visitExpr(e.Right)
	case *ast.TernaryExpr:
		l.visitExpr(e.Condition)
		l.visitExpr(e.Consequence)
		l.visitExpr(e.Alternative)
	case *ast.AssignExpr:
		l.visitExpr(e.Target)
		l.visitExpr(e.Value)
	case *ast.CallExpr:
		l.visitExpr(e.Callee)
		for _, arg := range e.Args {
			l.visitExpr(arg)
		}
	case *ast.LambdaExpr:
		l.visitExpr(e.Body)
	case *ast.SelectorExpr:
		l.visitExpr(e.Receiver)
	case *ast.IndexExpr:
		l.visitExpr(e.Receiver)
		l.visitExpr(e.Index)
	case *ast.ArrayLiteral:
		for _, elem := range e.Elements {
			l.visitExpr(elem)
		}
	case *ast.TupleLiteral:
		for _, elem := range e.Elements {
			l.visitExpr(elem)
		}
	case *ast.SpreadExpr:
		l.visitExpr(e.Expr)
	case *ast.ReactiveLiteral:
		l.visitExpr(e.Value)
	case *ast.MapLiteral:
		for _, entry := range e.Entries {
			l.visitExpr(entry.Key)
			l.visitExpr(entry.Value)
		}
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			l.visitExpr(field.Value)
		}
	case *ast.AnonymousObjectLiteral:
		for _, field := range e.Fields {
			l.visitExpr(field.Value)
		}
	case *ast.XMLElement:
		for _, attr := range e.Attrs {
			l.visitExpr(attr.Value)
		}
		for _, child := range e.Children {
			l.visitExpr(child.Expr)
		}
	case *ast.BlockExpr:
		l.visitBlock(e)
	case *ast.PatternBlock:
		for _, branch := range e.Branches {
			l.visitExpr(branch.Expr)
		}
	case *ast.MatchExpr:
		subject := l.info.ExprTypes[e.Subject]
		l.visitExpr(e.Subject)
		l.lintPatternBranches(e.Branches, subject)
		for _, branch := range e.Branches {
			l.visitExpr(branch.Expr)
		}
	case *ast.WatchExpr:
		l.visitExpr(e.Target)
		l.visitExpr(e.Handler)
	}
}

func (l *linter) visitBlock(block *ast.BlockExpr) {
	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.LetStmt:
			l.visitExpr(s.Value)
		case *ast.ObjectDestructureStmt:
			l.markDestructuredFields(s)
			l.visitExpr(s.Value)
		case *ast.AssignStmt:
			l.visitExpr(s.Value)
		case *ast.ExprStmt:
			l.visitExpr(s.Expr)
		}
	}
}

func (l *linter) visitExprNode(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Identifier:
		if fn := l.info.ResolvedFunctions[e]; fn != nil {
			l.usedFunctions[fn] = true
		}
	case *ast.BinaryExpr:
		l.lintPatternPredicateBinary(e)
	case *ast.TernaryExpr:
		l.lintPatternPredicateTernary(e)
	case *ast.SelectorExpr:
		if fn := l.info.ResolvedSelectorFunctions[e]; fn != nil {
			l.usedFunctions[fn] = true
		}
		if enumName, memberName, ok := l.enumMemberSelector(e); ok {
			l.constructedEnumMembers[enumMemberKey(enumName, memberName)] = true
		}
		if e.Static {
			return
		}
		receiver := l.info.ExprTypes[e.Receiver]
		if typ := l.info.Types[baseTypeName(receiver)]; typ != nil {
			if _, ok := typ.ByName[e.Name]; ok {
				l.usedFields[fieldKey(typ.Name, e.Name)] = true
			}
		}
	}
}

func (l *linter) lintPatternPredicateBinary(expr *ast.BinaryExpr) {
	if expr.Op != lexer.OrOr {
		return
	}
	comparisons := flattenBinaryExpr(expr, lexer.OrOr)
	if len(comparisons) < 3 {
		return
	}
	subject := ""
	for _, comparison := range comparisons {
		binary, ok := comparison.(*ast.BinaryExpr)
		if !ok || binary.Op != lexer.EqualEqual || !isPatternPredicateValue(binary.Right) {
			return
		}
		key := patternPredicateSubjectKey(binary.Left)
		if key == "" {
			return
		}
		if subject == "" {
			subject = key
			continue
		}
		if subject != key {
			return
		}
	}
	l.error(expr.Pos, "0013", "prefer_pattern_match", "Use '~' with an or-pattern instead of chained equality checks")
}

func (l *linter) lintPatternPredicateTernary(expr *ast.TernaryExpr) {
	subject := ""
	count := 0
	current := expr
	for current != nil {
		binary, ok := current.Condition.(*ast.BinaryExpr)
		if !ok || binary.Op != lexer.EqualEqual || !isPatternPredicateValue(binary.Right) {
			break
		}
		key := patternPredicateSubjectKey(binary.Left)
		if key == "" {
			break
		}
		if subject == "" {
			subject = key
		} else if subject != key {
			break
		}
		count++
		next, _ := current.Alternative.(*ast.TernaryExpr)
		current = next
	}
	if count < 3 {
		return
	}
	l.error(expr.Pos, "0013", "prefer_pattern_match", "Use pattern matching instead of a chained equality ternary")
}

func flattenBinaryExpr(expr ast.Expr, op lexer.Kind) []ast.Expr {
	binary, ok := expr.(*ast.BinaryExpr)
	if !ok || binary.Op != op {
		return []ast.Expr{expr}
	}
	out := flattenBinaryExpr(binary.Left, op)
	return append(out, flattenBinaryExpr(binary.Right, op)...)
}

func patternPredicateSubjectKey(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return "ident:" + e.Name
	case *ast.SelectorExpr:
		receiver := patternPredicateSubjectKey(e.Receiver)
		if receiver == "" {
			return ""
		}
		return receiver + "." + e.Name
	default:
		return ""
	}
}

func isPatternPredicateValue(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.BoolLiteral, *ast.IntegerLiteral, *ast.DoubleLiteral, *ast.BigIntLiteral, *ast.StringLiteral, *ast.CharLiteral, *ast.NullLiteral, *ast.SelectorExpr:
		return true
	default:
		return false
	}
}

func (l *linter) markDestructuredFields(stmt *ast.ObjectDestructureStmt) {
	valueType := l.info.ExprTypes[stmt.Value]
	typ := l.info.Types[baseTypeName(valueType)]
	if typ == nil {
		return
	}
	for _, field := range stmt.Fields {
		l.usedFields[fieldKey(typ.Name, field.Field)] = true
	}
}

func (l *linter) warnUnusedFunctions() {
	for _, fn := range l.info.FunctionDecls {
		if fn == nil || fn.Node == nil || fn.External || fn.Macro || fn.Name == "main" {
			continue
		}
		if l.usedFunctions[fn] {
			continue
		}
		l.warn(fn.NamePos, "0001", "unused_value", "Function %q is never used", fn.Name)
	}
}

func (l *linter) warnUnusedFields() {
	for _, typ := range l.info.Types {
		if typ == nil || typ.Node == nil {
			continue
		}
		for _, field := range typ.Node.Fields {
			if l.usedFields[fieldKey(typ.Name, field.Name)] {
				continue
			}
			l.warn(field.Pos, "0007", "unused_field", "Field %q is never read", field.Name)
		}
	}
}

func (l *linter) warnUnusedConstructors() {
	for _, enum := range l.info.Enums {
		if enum == nil || enum.Node == nil {
			continue
		}
		for _, member := range enum.Node.Members {
			if l.constructedEnumMembers[enumMemberKey(enum.Name, member.Name)] {
				continue
			}
			l.warn(member.Pos, "0006", "unused_constructor", "Variant %q is never constructed", member.Name)
		}
	}
}

func (l *linter) lintPatternBranches(branches []ast.PatternBranch, subject Type) {
	enum := l.info.Enums[string(subject)]
	if enum == nil {
		return
	}
	covered := map[string]bool{}
	exhaustive := false
	for _, branch := range branches {
		if exhaustive {
			l.warn(branch.Pattern.Position(), "0012", "unreachable_code", "Unreachable pattern branch")
			continue
		}
		members, all := l.patternEnumCoverage(branch.Pattern, enum.Name)
		if all {
			exhaustive = true
			continue
		}
		for _, member := range members {
			covered[member] = true
			l.constructedEnumMembers[enumMemberKey(enum.Name, member)] = true
		}
		if len(covered) == len(enum.Members) {
			exhaustive = true
		}
	}
}

func (l *linter) patternEnumCoverage(pattern ast.Pattern, enumName string) ([]string, bool) {
	switch p := pattern.(type) {
	case *ast.WildcardPattern, *ast.BindingPattern:
		return nil, true
	case *ast.LiteralPattern:
		if member, ok := l.enumMemberPattern(p.Value, enumName); ok {
			return []string{member}, false
		}
	case *ast.ConstructorPattern:
		if enum := l.info.Enums[enumName]; enum != nil {
			if _, ok := enum.ByName[p.Name]; ok {
				return []string{p.Name}, false
			}
		}
	case *ast.OrPattern:
		var out []string
		for _, alt := range p.Alternatives {
			members, all := l.patternEnumCoverage(alt, enumName)
			if all {
				return nil, true
			}
			out = append(out, members...)
		}
		return out, false
	}
	return nil, false
}

func (l *linter) enumMemberPattern(expr ast.Expr, enumName string) (string, bool) {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	ident, ok := sel.Receiver.(*ast.Identifier)
	if !ok || ident.Name != enumName {
		return "", false
	}
	enum := l.info.Enums[enumName]
	if enum == nil {
		return "", false
	}
	if _, ok := enum.ByName[sel.Name]; !ok {
		return "", false
	}
	return sel.Name, true
}

func (l *linter) enumMemberSelector(sel *ast.SelectorExpr) (string, string, bool) {
	ident, ok := sel.Receiver.(*ast.Identifier)
	if !ok {
		return "", "", false
	}
	enum := l.info.Enums[ident.Name]
	if enum == nil || l.info.ExprTypes[sel] != Type(enum.Name) {
		return "", "", false
	}
	if _, ok := enum.ByName[sel.Name]; !ok {
		return "", "", false
	}
	return enum.Name, sel.Name, true
}

func (l *linter) warn(pos lexer.Position, code string, kind string, format string, args ...any) {
	message := fmt.Sprintf("Warning [%s] (%s): %s", code, kind, fmt.Sprintf(format, args...))
	l.diags = append(l.diags, Diagnostic{
		Message:  message,
		Pos:      pos,
		Severity: SeverityWarning,
		Code:     code,
		Kind:     kind,
	})
}

func (l *linter) error(pos lexer.Position, code string, kind string, format string, args ...any) {
	if l.sourcePath != "" && l.currentSourcePath != "" && l.currentSourcePath != l.sourcePath {
		return
	}
	message := fmt.Sprintf("Error [%s] (%s): %s", code, kind, fmt.Sprintf(format, args...))
	l.diags = append(l.diags, Diagnostic{
		Message:  message,
		Pos:      pos,
		Severity: SeverityError,
		Code:     code,
		Kind:     kind,
	})
}

func enumMemberKey(enumName string, memberName string) string {
	return enumName + "." + memberName
}

func fieldKey(typeName string, fieldName string) string {
	return typeName + "." + fieldName
}
