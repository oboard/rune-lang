// Package tsfront converts a conservative TypeScript subset into the Rune AST.
//
// It is the first step of RFC 0001's TS -> Rune -> Go/MoonBit pipeline. The
// converter intentionally rejects constructs that Rune cannot represent yet
// (classes, unions, try/catch, dynamic features) instead of emitting a lossy
// approximation.
package tsfront

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/kdy1/go-typescript-eslint/pkg/typescriptestree"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

// Diagnostic reports a TypeScript construct the Rune frontend cannot convert.
type Diagnostic struct {
	Message string
	// Code is the stable RFC 0001 diagnostic code (`RUNE-DYN-001`, ...).
	// Empty for generic conversion errors that are not part of the profile
	// contract.
	Code   string
	Line   int
	Column int
	Path   string
}

// ConvertSource parses TypeScript source and returns the equivalent Rune AST.
func ConvertSource(path, source string) (*ast.File, []Diagnostic) {
	if diags := scanDeclareStatements(source, path); len(diags) > 0 {
		return nil, diags
	}
	source = preprocessASISemicolons(source)
	opts := typescriptestree.NewBuilder().
		WithSourceType(typescriptestree.SourceTypeModule).
		WithFilePath("input.ts").
		WithRange(true).
		WithLoc(true).
		MustBuild()
	result, err := typescriptestree.Parse(source, opts)
	if err != nil {
		if code := parseErrorCode(err.Error()); code != "" {
			return nil, []Diagnostic{{Code: code, Message: "parse TypeScript: " + err.Error(), Path: path}}
		}
		return nil, []Diagnostic{{Message: "parse TypeScript: " + err.Error(), Path: path}}
	}
	data, err := json.Marshal(result.AST.Body)
	if err != nil {
		return nil, []Diagnostic{{Message: "read TypeScript AST: " + err.Error(), Path: path}}
	}
	var statements []map[string]any
	if err := json.Unmarshal(data, &statements); err != nil {
		return nil, []Diagnostic{{Message: "read TypeScript AST: " + err.Error(), Path: path}}
	}
	c := &converter{
		path:       path,
		classes:    map[string]bool{},
		declared:   map[string]bool{},
		globalAny:  map[string]bool{},
		classNodes: map[string]map[string]any{},
		classInfos: map[string]*classInfo{},
		converting: map[string]bool{},
	}
	collectTopLevelDeclarations(statements, c.declared)
	for _, stmt := range statements {
		candidate := stmt
		if stringField(stmt, "type") == "ExportNamedDeclaration" {
			if decl, ok := mapField(stmt, "declaration"); ok {
				candidate = decl
			}
		}
		if stringField(candidate, "type") == "ClassDeclaration" {
			if id, ok := mapField(candidate, "id"); ok {
				if name := stringField(id, "name"); name != "" {
					c.classes[name] = true
					c.classNodes[name] = candidate
				}
			}
		}
	}
	if diags := scanProfileViolations(statements, path); len(diags) > 0 {
		// Dynamic constructs are rejected before conversion so callers always
		// see the stable D7 code instead of a secondary "unsupported node".
		return nil, diags
	}
	file := &ast.File{}
	c.file = file
	for _, stmt := range statements {
		c.topLevel(stmt, file)
	}
	return file, c.diags
}

type converter struct {
	path       string
	diags      []Diagnostic
	classes    map[string]bool
	declared   map[string]bool
	scopes     []map[string]bool
	anyScopes  []map[string]bool
	varTypes   []map[string]string
	classStack []string
	globalAny  map[string]bool
	classNodes map[string]map[string]any
	classInfos map[string]*classInfo
	converting map[string]bool
	file       *ast.File
}

func (c *converter) errorf(node map[string]any, format string, args ...any) {
	line, column := nodePosition(node)
	c.diags = append(c.diags, Diagnostic{
		Message: fmt.Sprintf(format, args...),
		Line:    line,
		Column:  column,
		Path:    c.path,
	})
}

func (c *converter) errorCodef(node map[string]any, code string, format string, args ...any) {
	line, column := nodePosition(node)
	c.diags = append(c.diags, Diagnostic{
		Code:    code,
		Message: "[" + code + "] " + fmt.Sprintf(format, args...),
		Line:    line,
		Column:  column,
		Path:    c.path,
	})
}

func (c *converter) topLevel(node map[string]any, file *ast.File) {
	kind := stringField(node, "type")
	if kind == "ExportNamedDeclaration" {
		decl, ok := mapField(node, "declaration")
		if !ok {
			return
		}
		node = decl
		kind = stringField(node, "type")
	}
	switch kind {
	case "EmptyStatement":
		return
	case "ImportDeclaration":
		source, ok := mapField(node, "source")
		if !ok {
			return
		}
		path := stringField(source, "value")
		if path == "" {
			return
		}
		file.Imports = append(file.Imports, ast.Import{Path: convertModulePath(path)})
	case "FunctionDeclaration":
		if fn, ok := c.function(node, false); ok {
			file.Functions = append(file.Functions, fn)
		}
	case "ClassDeclaration":
		c.class(node, file)
	case "VariableDeclaration":
		for _, decl := range c.variableDeclaration(node, false) {
			file.Constants = append(file.Constants, decl)
		}
	case "ExpressionStatement":
		// Top-level calls are not representable as Rune declarations; wrap them
		// in a synthetic main so the pipeline can still execute the module.
		if len(file.Functions) > 0 {
			c.errorf(node, "top-level expression statements are only supported in main")
			return
		}
		expr, ok := c.expressionField(node, "expression")
		if !ok {
			return
		}
		file.Functions = append(file.Functions, &ast.Function{
			Name:       "main",
			ReturnType: ast.NamedType("Void"),
			Body:       &ast.BlockExpr{Statements: []ast.Stmt{&ast.ExprStmt{Expr: expr}}},
		})
	case "":
		return
	default:
		c.errorf(node, "unsupported top-level TypeScript node %q", kind)
	}
}

func convertModulePath(path string) string {
	for _, extension := range []string{".tsx", ".ts", ".jsx", ".js"} {
		if strings.HasSuffix(path, extension) {
			return strings.TrimSuffix(path, extension) + ".rune"
		}
	}
	if !strings.HasSuffix(path, ".rune") {
		return path + ".rune"
	}
	return path
}

func (c *converter) function(node map[string]any, private bool) (*ast.Function, bool) {
	id, ok := mapField(node, "id")
	if !ok {
		c.errorf(node, "anonymous TypeScript functions are not supported")
		return nil, false
	}
	name := stringField(id, "name")
	params := make([]ast.Param, 0)
	paramInfos := make([]astParamInfo, 0)
	for _, raw := range arrayField(node, "params") {
		param, ok := asObject(raw)
		if !ok || stringField(param, "type") != "Identifier" {
			c.errorf(node, "only identifier parameters are supported")
			return nil, false
		}
		paramType, any, ok := c.typedParam(param)
		if !ok {
			return nil, false
		}
		paramName := stringField(param, "name")
		params = append(params, ast.Param{Name: paramName, Type: paramType})
		paramInfos = append(paramInfos, astParamInfo{name: paramName, any: any, class: c.classNameOf(paramType)})
	}
	returnType := ast.NamedType("Void")
	if annotation, ok := mapField(node, "returnType"); ok {
		if converted, ok := c.typeAnnotation(annotation); ok {
			returnType = converted
		} else {
			return nil, false
		}
	}
	bodyNode, ok := mapField(node, "body")
	if !ok {
		return nil, false
	}
	restore := c.pushScope(paramInfos)
	defer restore()
	block, ok := c.block(bodyNode)
	if !ok {
		return nil, false
	}
	return &ast.Function{
		Name:       name,
		Private:    private,
		Params:     params,
		ReturnType: returnType,
		Body:       block,
	}, true
}

func (c *converter) statementField(node map[string]any, key string) (ast.Stmt, bool) {
	raw, ok := mapField(node, key)
	if !ok {
		return nil, false
	}
	return c.statement(raw)
}

func (c *converter) statement(node map[string]any) (ast.Stmt, bool) {
	switch stringField(node, "type") {
	case "BlockStatement":
		block, ok := c.block(node)
		if !ok {
			return nil, false
		}
		return &ast.ExprStmt{Expr: block}, true
	case "ReturnStatement":
		stmt := &ast.ReturnStmt{}
		if arg, ok := mapField(node, "argument"); ok {
			expr, ok := c.expression(arg)
			if !ok {
				return nil, false
			}
			stmt.Value = expr
		}
		return stmt, true
	case "ThrowStatement":
		arg, ok := mapField(node, "argument")
		if !ok {
			c.errorf(node, "throw statement requires a value")
			return nil, false
		}
		expr, ok := c.expression(arg)
		if !ok {
			return nil, false
		}
		return &ast.ThrowStmt{Value: expr}, true
	case "TryStatement":
		blockNode, ok := mapField(node, "block")
		if !ok {
			return nil, false
		}
		body, ok := c.block(blockNode)
		if !ok {
			return nil, false
		}
		stmt := &ast.TryStmt{Body: body}
		if handler, ok := mapField(node, "handler"); ok {
			if param, ok := mapField(handler, "param"); ok {
				stmt.CatchName = stringField(param, "name")
			}
			catchNode, ok := mapField(handler, "body")
			if !ok {
				return nil, false
			}
			stmt.Catch, ok = c.block(catchNode)
			if !ok {
				return nil, false
			}
		}
		if finalizer, ok := mapField(node, "finalizer"); ok {
			stmt.Finally, ok = c.block(finalizer)
			if !ok {
				return nil, false
			}
		}
		return stmt, true
	case "VariableDeclaration":
		decls := arrayField(node, "declarations")
		if len(decls) != 1 {
			c.errorf(node, "only single-declarator variable statements are supported")
			return nil, false
		}
		decl, ok := asObject(decls[0])
		if !ok {
			return nil, false
		}
		id, ok := mapField(decl, "id")
		if !ok {
			return nil, false
		}
		value, ok := c.expressionField(decl, "init")
		if !ok {
			c.errorf(node, "variable declarations without initializer are not supported")
			return nil, false
		}
		stmt := &ast.LetStmt{
			Name:    stringField(id, "name"),
			Mutable: stringField(node, "kind") != "const",
			Value:   value,
		}
		c.declare(stmt.Name)
		if annotation, ok := mapField(id, "typeAnnotation"); ok {
			if converted, ok := c.typeAnnotation(annotation); ok {
				stmt.Type = converted
				if isDynamicType(converted) {
					c.markAny(stmt.Name)
				}
			}
		}
		class := c.classOfValue(value)
		if stmt.Type.Kind == ast.TypeName && c.classes[stmt.Type.Name] {
			class = stmt.Type.Name
		}
		c.declareClassVar(stmt.Name, class)
		return stmt, true
	case "ExpressionStatement":
		expr, ok := c.expressionField(node, "expression")
		if !ok {
			return nil, false
		}
		return &ast.ExprStmt{Expr: expr}, true
	case "IfStatement":
		cond, ok := c.expressionField(node, "test")
		if !ok {
			return nil, false
		}
		thenStmt, ok := c.statementField(node, "consequent")
		if !ok {
			return nil, false
		}
		var elseExpr ast.Expr
		if alt, ok := mapField(node, "alternate"); ok {
			altStmt, ok := c.statement(alt)
			if !ok {
				return nil, false
			}
			elseExpr = statementExpr(altStmt)
		}
		return &ast.ExprStmt{Expr: &ast.IfExpr{Cond: cond, Then: statementExpr(thenStmt), Else: elseExpr}}, true
	case "BreakStatement":
		return &ast.BreakStmt{}, true
	case "ContinueStatement":
		return &ast.ContinueStmt{}, true
	case "WhileStatement":
		cond, ok := c.expressionField(node, "test")
		if !ok {
			return nil, false
		}
		body, ok := c.statementField(node, "body")
		if !ok {
			return nil, false
		}
		return &ast.ExprStmt{Expr: &ast.WhileExpr{Cond: cond, Body: statementExpr(body)}}, true
	case "ForOfStatement":
		return c.forOfStatement(node)
	case "ForStatement":
		return c.forStatement(node)
	case "DoWhileStatement":
		return c.doWhileStatement(node)
	default:
		c.errorf(node, "unsupported TypeScript statement %q", stringField(node, "type"))
		return nil, false
	}
}

func (c *converter) forOfStatement(node map[string]any) (ast.Stmt, bool) {
	left, ok := mapField(node, "left")
	if !ok {
		c.errorf(node, "for-of left-hand side must be a variable declaration")
		return nil, false
	}
	name := ""
	if stringField(left, "type") == "VariableDeclaration" {
		decls := arrayField(left, "declarations")
		if len(decls) == 1 {
			if decl, ok := asObject(decls[0]); ok {
				if id, ok := mapField(decl, "id"); ok {
					name = stringField(id, "name")
				}
			}
		}
	} else {
		name = stringField(left, "name")
	}
	if name == "" {
		c.errorf(node, "for-of loop variable is required")
		return nil, false
	}
	iterable, ok := c.expressionField(node, "right")
	if !ok {
		return nil, false
	}
	body, ok := c.statementField(node, "body")
	if !ok {
		return nil, false
	}
	lambda := &ast.LambdaExpr{Params: []string{name}, Body: statementExpr(body)}
	call := &ast.CallExpr{
		Callee: &ast.SelectorExpr{Receiver: iterable, Name: "each"},
		Args:   []ast.Expr{lambda},
	}
	return &ast.ExprStmt{Expr: call}, true
}

func statementExpr(stmt ast.Stmt) ast.Expr {
	if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
		return exprStmt.Expr
	}
	return &ast.BlockExpr{Statements: []ast.Stmt{stmt}}
}

func (c *converter) block(node map[string]any) (*ast.BlockExpr, bool) {
	block := &ast.BlockExpr{}
	for _, raw := range arrayField(node, "body") {
		stmtNode, ok := asObject(raw)
		if !ok {
			return nil, false
		}
		stmt, ok := c.statement(stmtNode)
		if !ok {
			return nil, false
		}
		block.Statements = append(block.Statements, stmt)
	}
	return block, true
}

func (c *converter) variableDeclaration(node map[string]any, mutable bool) []*ast.ConstDecl {
	kind := stringField(node, "kind")
	mutable = mutable || kind == "let" || kind == "var"
	out := make([]*ast.ConstDecl, 0)
	for _, raw := range arrayField(node, "declarations") {
		decl, ok := asObject(raw)
		if !ok {
			continue
		}
		id, ok := mapField(decl, "id")
		if !ok {
			continue
		}
		name := stringField(id, "name")
		if name == "" {
			continue
		}
		value, ok := c.expressionField(decl, "init")
		if !ok {
			continue
		}
		if annotation, ok := mapField(id, "typeAnnotation"); ok {
			if converted, ok := c.typeAnnotation(annotation); ok && isDynamicType(converted) {
				c.markAny(name)
			}
		}
		_ = mutable // Rune constants are immutable; let is represented as a local when inside blocks.
		out = append(out, &ast.ConstDecl{Name: name, Value: value})
	}
	return out
}

func (c *converter) expressionField(node map[string]any, key string) (ast.Expr, bool) {
	raw, ok := mapField(node, key)
	if !ok {
		return nil, false
	}
	return c.expression(raw)
}

func (c *converter) expression(node map[string]any) (ast.Expr, bool) {
	switch stringField(node, "type") {
	case "Identifier":
		name := stringField(node, "name")
		if name == "undefined" {
			return &ast.UndefinedLiteral{}, true
		}
		return &ast.Identifier{Name: name}, true
	case "Literal":
		return c.literal(node)
	case "BinaryExpression":
		op, ok := binaryOp(stringField(node, "operator"))
		if !ok {
			c.errorf(node, "unsupported binary operator %q", stringField(node, "operator"))
			return nil, false
		}
		left, ok := c.expressionField(node, "left")
		if !ok {
			return nil, false
		}
		right, ok := c.expressionField(node, "right")
		if !ok {
			return nil, false
		}
		if op == lexer.Plus {
			if literal, ok := left.(*ast.StringLiteral); ok {
				return &ast.TemplateLiteral{Parts: []ast.TemplatePart{{Text: literal.Value, Expr: right}}}, true
			}
		}
		return &ast.BinaryExpr{Op: op, Left: left, Right: right}, true
	case "UnaryExpression":
		if stringField(node, "operator") == "typeof" {
			value, ok := c.expressionField(node, "argument")
			if !ok {
				return nil, false
			}
			return &ast.TypeOfExpr{Argument: value}, true
		}
		op, ok := unaryOp(stringField(node, "operator"))
		if !ok {
			c.errorf(node, "unsupported unary operator %q", stringField(node, "operator"))
			return nil, false
		}
		value, ok := c.expressionField(node, "argument")
		if !ok {
			return nil, false
		}
		return &ast.UnaryExpr{Op: op, Expr: value}, true
	case "ParenthesizedExpression":
		return c.expressionField(node, "expression")
	case "ThisExpression":
		return &ast.ThisExpr{}, true
	case "Super":
		c.errorCodef(node, CodeProfileClass, "super member access is not supported yet")
		return nil, false
	case "TSAsExpression":
		return c.expressionField(node, "expression")
	case "CallExpression":
		return c.call(node)
	case "NewExpression":
		return c.newExpression(node)
	case "ArrowFunctionExpression", "FunctionExpression":
		return c.lambda(node)
	case "ArrayExpression":
		lit := &ast.ArrayLiteral{}
		for _, raw := range arrayField(node, "elements") {
			element, ok := asObject(raw)
			if !ok {
				continue
			}
			if stringField(element, "type") == "SpreadElement" {
				value, ok := c.expressionField(element, "argument")
				if !ok {
					return nil, false
				}
				lit.Elements = append(lit.Elements, &ast.SpreadExpr{Expr: value})
				continue
			}
			value, ok := c.expression(element)
			if !ok {
				return nil, false
			}
			lit.Elements = append(lit.Elements, value)
		}
		return lit, true
	case "ObjectExpression":
		lit := &ast.AnonymousObjectLiteral{}
		for _, raw := range arrayField(node, "properties") {
			property, ok := asObject(raw)
			if !ok {
				continue
			}
			if stringField(property, "type") == "SpreadElement" {
				value, ok := c.expressionField(property, "argument")
				if !ok {
					return nil, false
				}
				lit.Fields = append(lit.Fields, ast.FieldValue{Spread: true, Value: value})
				continue
			}
			key, ok := mapField(property, "key")
			if !ok {
				c.errorf(property, "object property key is required")
				return nil, false
			}
			name := stringField(key, "name")
			if name == "" {
				if value, ok := key["value"].(string); ok {
					name = value
				}
			}
			if name == "" {
				c.errorf(property, "computed object keys are not supported")
				return nil, false
			}
			value, ok := c.expressionField(property, "value")
			if !ok {
				return nil, false
			}
			lit.Fields = append(lit.Fields, ast.FieldValue{Name: name, Value: value})
		}
		return lit, true
	case "IndexExpression":
		receiver, ok := c.expressionField(node, "object")
		if !ok {
			return nil, false
		}
		indexNode, ok := mapField(node, "property")
		if !ok {
			c.errorf(node, "computed index expression requires a property")
			return nil, false
		}
		if computed, _ := node["computed"].(bool); computed {
			index, ok := c.indexExpression(indexNode)
			if !ok {
				return nil, false
			}
			return &ast.IndexExpr{Receiver: receiver, Index: index}, true
		}
		name := memberName(node)
		if name == "" {
			return nil, false
		}
		return &ast.SelectorExpr{Receiver: receiver, Name: name}, true
	case "LogicalExpression":
		op, ok := binaryOp(stringField(node, "operator"))
		if !ok {
			c.errorf(node, "unsupported logical operator %q", stringField(node, "operator"))
			return nil, false
		}
		left, ok := c.expressionField(node, "left")
		if !ok {
			return nil, false
		}
		right, ok := c.expressionField(node, "right")
		if !ok {
			return nil, false
		}
		return &ast.BinaryExpr{Op: op, Left: left, Right: right}, true
	case "ConditionalExpression":
		cond, ok := c.expressionField(node, "test")
		if !ok {
			return nil, false
		}
		consequence, ok := c.expressionField(node, "consequent")
		if !ok {
			return nil, false
		}
		alternative, ok := c.expressionField(node, "alternate")
		if !ok {
			return nil, false
		}
		return &ast.TernaryExpr{Condition: cond, Consequence: consequence, Alternative: alternative}, true
	case "TemplateLiteral":
		return c.templateLiteral(node)
	case "UpdateExpression":
		argument, ok := c.expressionField(node, "argument")
		if !ok {
			return nil, false
		}
		if stringField(node, "operator") != "++" {
			c.errorf(node, "only postfix ++ is supported")
			return nil, false
		}
		return &ast.PostfixExpr{Op: lexer.PlusPlus, Expr: argument}, true
	case "MemberExpression":
		if object, ok := mapField(node, "object"); ok && stringField(object, "type") == "Identifier" {
			className := stringField(object, "name")
			if c.classes[className] {
				if constName, ok := c.staticFieldName(className, memberName(node)); ok {
					return &ast.Identifier{Name: constName}, true
				}
			}
		}
		if object, ok := mapField(node, "object"); ok && stringField(object, "type") == "Super" {
			name := memberName(node)
			if name == "" {
				c.errorf(node, "computed super member expressions are not supported")
				return nil, false
			}
			if base := c.classBase(c.currentClass()); base != "" {
				if get, _ := c.classAccessor(base, name); get {
					return &ast.CallExpr{Callee: &ast.SelectorExpr{Receiver: &ast.ThisExpr{}, Name: "__super_get_" + name}}, true
				}
			}
			return &ast.SelectorExpr{Receiver: &ast.ThisExpr{}, Name: name}, true
		}
		if object, ok := mapField(node, "object"); ok && stringField(object, "type") == "Identifier" {
			if name := stringField(object, "name"); c.isAny(name) {
				c.errorCodef(node, CodeDynamicAnyMember, "member access on `any` value %q is not supported", name)
				return nil, false
			}
		}
		receiver, ok := c.expressionField(node, "object")
		if !ok {
			return nil, false
		}
		if computed, _ := node["computed"].(bool); computed {
			prop, ok := mapField(node, "property")
			if !ok {
				return nil, false
			}
			index, ok := c.indexExpression(prop)
			if !ok {
				return nil, false
			}
			return &ast.IndexExpr{Receiver: receiver, Index: index}, true
		}
		name := memberName(node)
		if name == "" {
			c.errorf(node, "computed member expressions are not supported")
			return nil, false
		}
		receiverClass := ""
		if object, ok := mapField(node, "object"); ok {
			switch stringField(object, "type") {
			case "ThisExpression":
				receiverClass = c.currentClass()
			case "Identifier":
				receiverClass = c.varClass(stringField(object, "name"))
			}
		}
		if receiverClass != "" {
			if get, _ := c.classAccessor(receiverClass, name); get {
				return &ast.CallExpr{Callee: &ast.SelectorExpr{Receiver: receiver, Name: "get_" + name}}, true
			}
		}
		return &ast.SelectorExpr{Receiver: receiver, Name: name}, true
	case "AssignmentExpression":
		if rawLeft, ok := mapField(node, "left"); ok && stringField(rawLeft, "type") == "MemberExpression" {
			name := memberName(rawLeft)
			object, _ := mapField(rawLeft, "object")
			receiverClass := ""
			isSuper := false
			if stringField(object, "type") == "Super" {
				isSuper = true
				receiverClass = c.classBase(c.currentClass())
			} else if stringField(object, "type") == "ThisExpression" {
				receiverClass = c.currentClass()
			} else if stringField(object, "type") == "Identifier" {
				receiverClass = c.varClass(stringField(object, "name"))
			}
			if name != "" && receiverClass != "" {
				get, set := c.classAccessor(receiverClass, name)
				if set {
					receiver := ast.Expr(&ast.ThisExpr{})
					if !isSuper {
						converted, ok := c.expression(object)
						if !ok {
							return nil, false
						}
						receiver = converted
					}
					value, ok := c.expressionField(node, "right")
					if !ok {
						return nil, false
					}
					setter := "set_" + name
					if isSuper {
						setter = "__super_set_" + name
					}
					return &ast.CallExpr{Callee: &ast.SelectorExpr{Receiver: receiver, Name: setter}, Args: []ast.Expr{value}}, true
				}
				if get && !set {
					c.errorCodef(node, CodeProfileClass, "property %q has no setter", name)
					return nil, false
				}
			}
		}
		left, ok := c.expressionField(node, "left")
		if !ok {
			return nil, false
		}
		value, ok := c.expressionField(node, "right")
		if !ok {
			return nil, false
		}
		ident, ok := left.(*ast.Identifier)
		if !ok {
			c.errorf(node, "only identifier assignment is supported")
			return nil, false
		}
		if !c.isDeclared(ident.Name) {
			c.errorCodef(node, CodeDynamicImplicitAny, "assignment to undeclared global %q", ident.Name)
			return nil, false
		}
		return &ast.AssignExpr{Name: ident.Name, Value: value}, true
	default:
		c.errorf(node, "unsupported TypeScript expression %q", stringField(node, "type"))
		return nil, false
	}
}

func (c *converter) lambda(node map[string]any) (ast.Expr, bool) {
	params := make([]string, 0)
	paramTypes := make([]ast.Type, 0)
	paramInfos := make([]astParamInfo, 0)
	for _, raw := range arrayField(node, "params") {
		param, ok := asObject(raw)
		if !ok || stringField(param, "type") != "Identifier" {
			c.errorf(node, "only identifier lambda parameters are supported")
			return nil, false
		}
		params = append(params, stringField(param, "name"))
		paramType := ast.Type{}
		if annotation, ok := mapField(param, "typeAnnotation"); ok {
			converted, ok := c.typeAnnotation(annotation)
			if !ok {
				return nil, false
			}
			paramType = converted
		}
		paramTypes = append(paramTypes, paramType)
		paramInfos = append(paramInfos, astParamInfo{
			name:  stringField(param, "name"),
			any:   isDynamicType(paramType),
			class: c.classNameOf(paramType),
		})
	}
	returnType := ast.Type{}
	if annotation, ok := mapField(node, "returnType"); ok {
		converted, ok := c.typeAnnotation(annotation)
		if !ok {
			return nil, false
		}
		returnType = converted
	}
	body, ok := mapField(node, "body")
	if !ok {
		return nil, false
	}
	restore := c.pushScope(paramInfos)
	defer restore()
	var bodyExpr ast.Expr
	if stringField(body, "type") == "BlockStatement" {
		block, ok := c.block(body)
		if !ok {
			return nil, false
		}
		bodyExpr = block
	} else {
		expr, ok := c.expression(body)
		if !ok {
			return nil, false
		}
		bodyExpr = expr
	}
	return &ast.LambdaExpr{
		Params:     params,
		ParamTypes: paramTypes,
		ReturnType: returnType,
		Body:       bodyExpr,
	}, true
}

func (c *converter) indexExpression(node map[string]any) (ast.Expr, bool) {
	if stringField(node, "type") == "Literal" {
		if raw := stringField(node, "raw"); raw != "" {
			if number, err := strconv.ParseFloat(raw, 64); err == nil {
				return &ast.IntegerLiteral{Value: int(number)}, true
			}
		}
	}
	return c.expression(node)
}

func (c *converter) templateLiteral(node map[string]any) (ast.Expr, bool) {
	quasis := arrayField(node, "quasis")
	expressions := arrayField(node, "expressions")
	lit := &ast.TemplateLiteral{}
	for index, raw := range quasis {
		quasi, ok := asObject(raw)
		if !ok {
			continue
		}
		value, _ := mapField(quasi, "value")
		text := stringField(value, "cooked")
		if text == "" {
			text = stringField(value, "raw")
		}
		part := ast.TemplatePart{Text: text}
		if index < len(expressions) {
			exprNode, ok := asObject(expressions[index])
			if !ok {
				return nil, false
			}
			expr, ok := c.expression(exprNode)
			if !ok {
				return nil, false
			}
			part.Expr = expr
		}
		lit.Parts = append(lit.Parts, part)
	}
	return lit, true
}

func (c *converter) call(node map[string]any) (ast.Expr, bool) {
	callee, ok := mapField(node, "callee")
	if !ok {
		return nil, false
	}
	args := make([]ast.Expr, 0)
	for _, raw := range arrayField(node, "arguments") {
		arg, ok := asObject(raw)
		if !ok {
			continue
		}
		expr, ok := c.expression(arg)
		if !ok {
			return nil, false
		}
		args = append(args, expr)
	}
	if stringField(callee, "type") == "MemberExpression" {
		object, _ := mapField(callee, "object")
		name := memberName(callee)
		if stringField(object, "type") == "Super" {
			return &ast.CallExpr{
				Callee: &ast.SelectorExpr{Receiver: &ast.ThisExpr{}, Name: "__super_" + name},
				Args:   args,
			}, true
		}
		if stringField(object, "type") == "Identifier" && stringField(object, "name") == "console" && name == "log" {
			return &ast.CallExpr{
				Callee: &ast.SelectorExpr{Receiver: &ast.AtExpr{Name: "io"}, Name: "println"},
				Args:   args,
			}, true
		}
		if stringField(object, "type") == "Identifier" && c.classes[stringField(object, "name")] {
			return &ast.CallExpr{
				Callee: &ast.Identifier{Name: stringField(object, "name") + "_" + name},
				Args:   args,
			}, true
		}
	}
	fn, ok := c.expression(callee)
	if !ok {
		return nil, false
	}
	return &ast.CallExpr{Callee: fn, Args: args}, true
}

func (c *converter) newExpression(node map[string]any) (ast.Expr, bool) {
	callee, ok := mapField(node, "callee")
	if !ok || stringField(callee, "type") != "Identifier" {
		c.errorCodef(node, CodeProfileClass, "only `new ClassName(...)` is supported (callee type %q)", stringField(callee, "type"))
		return nil, false
	}
	name := stringField(callee, "name")
	if !c.classes[name] {
		c.errorCodef(node, CodeProfileClass, "`new` is only supported for classes declared in this module")
		return nil, false
	}
	args := make([]ast.Expr, 0)
	for _, raw := range arrayField(node, "arguments") {
		arg, ok := asObject(raw)
		if !ok {
			continue
		}
		expr, ok := c.expression(arg)
		if !ok {
			return nil, false
		}
		args = append(args, expr)
	}
	return &ast.CallExpr{Callee: &ast.Identifier{Name: name + "_new"}, Args: args}, true
}

func (c *converter) literal(node map[string]any) (ast.Expr, bool) {
	value, ok := node["value"]
	if !ok {
		return nil, false
	}
	switch v := value.(type) {
	case float64:
		raw := stringField(node, "raw")
		return &ast.DoubleLiteral{Value: v, Raw: raw}, true
	case string:
		raw := stringField(node, "raw")
		if raw != "" {
			if number, err := strconv.ParseFloat(raw, 64); err == nil {
				return &ast.DoubleLiteral{Value: number, Raw: raw}, true
			}
			if raw == "true" {
				return &ast.BoolLiteral{Value: true}, true
			}
			if raw == "false" {
				return &ast.BoolLiteral{Value: false}, true
			}
		}
		return &ast.StringLiteral{Value: v}, true
	case bool:
		return &ast.BoolLiteral{Value: v}, true
	case nil:
		return &ast.NullLiteral{}, true
	default:
		c.errorf(node, "unsupported literal value %T", value)
		return nil, false
	}
}

func (c *converter) typeAnnotation(node map[string]any) (ast.Type, bool) {
	inner, ok := mapField(node, "typeAnnotation")
	if !ok {
		return ast.NamedType("Dynamic"), true
	}
	switch stringField(inner, "type") {
	case "TSNumberKeyword":
		return ast.NamedType("Double"), true
	case "TSStringKeyword":
		return ast.NamedType("String"), true
	case "TSBooleanKeyword":
		return ast.NamedType("Bool"), true
	case "TSVoidKeyword":
		return ast.NamedType("Void"), true
	case "TSNullKeyword":
		return ast.NamedType("Null"), true
	case "TSUndefinedKeyword":
		return ast.NamedType("Undefined"), true
	case "TSAnyKeyword", "TSUnknownKeyword":
		return ast.NamedType("Dynamic"), true
	case "TSArrayType":
		elem, ok := mapField(inner, "elementType")
		if !ok {
			return ast.RawType("Dynamic"), true
		}
		elemType, ok := c.typeNode(elem)
		if !ok {
			return ast.RawType(""), false
		}
		return ast.NamedType("Array").WithArgs([]ast.Type{elemType}), true
	case "TSTypeReference":
		name := stringField(inner, "typeName")
		if nameNode, ok := mapField(inner, "typeName"); ok {
			name = stringField(nameNode, "name")
			if name == "" {
				name = stringField(nameNode, "left")
			}
		}
		switch name {
		case "Array":
			if params := arrayField(inner, "typeParameters"); len(params) > 0 {
				if param, ok := asObject(params[0]); ok {
					if elem, ok := c.typeNode(param); ok {
						return ast.NamedType("Array").WithArgs([]ast.Type{elem}), true
					}
				}
			}
		}
		if name == "" {
			c.errorf(inner, "unsupported TypeScript type reference")
			return ast.RawType(""), false
		}
		return ast.NamedType(name), true
	case "TSUnionType":
		members := make([]ast.Type, 0)
		for _, raw := range arrayField(inner, "types") {
			member, ok := asObject(raw)
			if !ok {
				continue
			}
			converted, ok := c.typeNode(member)
			if !ok {
				return ast.RawType(""), false
			}
			members = append(members, converted)
		}
		if len(members) == 0 {
			c.errorf(inner, "empty TypeScript union type")
			return ast.RawType(""), false
		}
		return ast.UnionType(members), true
	default:
		c.errorf(inner, "unsupported TypeScript type %q", stringField(inner, "type"))
		return ast.RawType(""), false
	}
}

func (c *converter) typeNode(node map[string]any) (ast.Type, bool) {
	return c.typeAnnotation(map[string]any{"typeAnnotation": node})
}

func binaryOp(operator string) (lexer.Kind, bool) {
	switch operator {
	case "+":
		return lexer.Plus, true
	case "-":
		return lexer.Minus, true
	case "*":
		return lexer.Star, true
	case "/":
		return lexer.Slash, true
	case "%":
		return lexer.Percent, true
	case "==", "===":
		return lexer.EqualEqual, true
	case "!=", "!==":
		return lexer.BangEqual, true
	case "<":
		return lexer.Less, true
	case "<=":
		return lexer.LessEqual, true
	case ">":
		return lexer.Greater, true
	case ">=":
		return lexer.GreaterEqual, true
	case "&&":
		return lexer.AndAnd, true
	case "||":
		return lexer.OrOr, true
	case "&":
		return lexer.BitAnd, true
	case "|":
		return lexer.BitOr, true
	case "^":
		return lexer.BitXor, true
	case "<<":
		return lexer.ShiftLeft, true
	case ">>":
		return lexer.ShiftRight, true
	case ">>>":
		return lexer.UnsignedShiftRight, true
	default:
		return lexer.Illegal, false
	}
}

// preprocessASISemicolons works around the TypeScript parser's incomplete ASI
// handling for `return`/`break`/`continue` at the start of a line. The Rune
// profile requires those statements to occupy a single line.
func preprocessASISemicolons(source string) string {
	lines := strings.Split(source, "\n")
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasSuffix(trimmed, ";") {
			continue
		}
		if trimmed == "return" || strings.HasPrefix(trimmed, "return ") ||
			trimmed == "break" || trimmed == "continue" {
			if strings.Contains(trimmed, "//") {
				continue
			}
			lines[index] = line + ";"
		}
	}
	return strings.Join(lines, "\n")
}

func unaryOp(operator string) (lexer.Kind, bool) {
	switch operator {
	case "-":
		return lexer.Minus, true
	case "!":
		return lexer.Bang, true
	default:
		return lexer.Illegal, false
	}
}

func memberName(node map[string]any) string {
	prop, ok := mapField(node, "property")
	if !ok {
		return ""
	}
	if computed, _ := node["computed"].(bool); computed {
		return ""
	}
	return stringField(prop, "name")
}

func mapField(node map[string]any, key string) (map[string]any, bool) {
	value, ok := node[key]
	if !ok {
		return nil, false
	}
	return asObject(value)
}

func asObject(value any) (map[string]any, bool) {
	object, ok := value.(map[string]any)
	return object, ok
}

func arrayField(node map[string]any, key string) []any {
	value, ok := node[key]
	if !ok {
		return nil
	}
	array, _ := value.([]any)
	return array
}

func stringField(node map[string]any, key string) string {
	value, _ := node[key].(string)
	return value
}

func nodePosition(node map[string]any) (int, int) {
	if loc, ok := mapField(node, "loc"); ok {
		start, _ := mapField(loc, "start")
		line, _ := start["line"].(float64)
		column, _ := start["column"].(float64)
		return int(line), int(column)
	}
	return 0, 0
}

// FormatDiagnostics renders converter diagnostics in the CLI style.
func FormatDiagnostics(path string, diags []Diagnostic) string {
	var b strings.Builder
	for _, diag := range diags {
		prefix := ""
		if diag.Code != "" && !strings.Contains(diag.Message, diag.Code) {
			prefix = "[" + diag.Code + "] "
		}
		fmt.Fprintf(&b, "%s:%d:%d: %s%s\n", path, diag.Line, diag.Column, prefix, diag.Message)
	}
	return b.String()
}

// ParseNumberLiteral is a helper used by tests and future callers.
func ParseNumberLiteral(raw string) (float64, bool) {
	value, err := strconv.ParseFloat(raw, 64)
	return value, err == nil
}
