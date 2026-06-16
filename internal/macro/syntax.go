package macro

import (
	"fmt"
	"strconv"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/interpreter"
	"github.com/oboard/rune-lang/internal/lexer"
)

type syntaxRefs struct {
	annotations map[string]ast.Annotation
	bodies      map[string]ast.Expr
	positions   map[string]lexer.Position
	generated   int
}

func newSyntaxRefs() *syntaxRefs {
	return &syntaxRefs{
		annotations: map[string]ast.Annotation{},
		bodies:      map[string]ast.Expr{},
		positions:   map[string]lexer.Position{},
	}
}

func syntaxFileValue(file *ast.File, refs *syntaxRefs) interpreter.Value {
	types := &interpreter.Array{}
	for _, typ := range file.Types {
		types.Elements = append(types.Elements, syntaxStructValue(typ, refs))
	}
	enums := &interpreter.Array{}
	for _, enum := range file.Enums {
		enums.Elements = append(enums.Elements, syntaxEnumValue(enum, refs))
	}
	functions := &interpreter.Array{}
	for _, fn := range file.Functions {
		if fn.Macro {
			continue
		}
		functions.Elements = append(functions.Elements, syntaxFunctionValue(fn, refs))
	}
	return structValue("SyntaxFile",
		"types", types,
		"enums", enums,
		"functions", functions,
	)
}

func macroContextValue(target Target) interpreter.Value {
	return structValue("MacroContext",
		"targetID", targetSyntaxID(target),
		"sourcePath", target.SourcePath,
	)
}

func syntaxStructValue(typ *ast.StructType, refs *syntaxRefs) interpreter.Value {
	id := nodeID("struct", typ.SourcePath, typ.Pos)
	refs.positions[id] = typ.Pos
	fields := &interpreter.Array{}
	for i := range typ.Fields {
		fields.Elements = append(fields.Elements, syntaxFieldValue(&typ.Fields[i], typ.SourcePath, refs))
	}
	methods := &interpreter.Array{}
	for _, method := range typ.Methods {
		if method.Macro {
			continue
		}
		methods.Elements = append(methods.Elements, syntaxFunctionValue(method, refs))
	}
	return structValue("SyntaxStruct",
		"id", id,
		"name", typ.Name,
		"private", typ.Private,
		"generics", stringArray(typ.Generics),
		"annotations", syntaxAnnotationsValue(typ.Annotations, typ.SourcePath, refs),
		"fields", fields,
		"methods", methods,
		"sourcePath", typ.SourcePath,
	)
}

func syntaxFieldValue(field *ast.Field, sourcePath string, refs *syntaxRefs) interpreter.Value {
	id := nodeID("field", sourcePath, field.Pos)
	refs.positions[id] = field.Pos
	return structValue("SyntaxField",
		"id", id,
		"name", field.Name,
		"private", field.Private,
		"annotations", syntaxAnnotationsValue(field.Annotations, sourcePath, refs),
		"type", syntaxTypeValue(field.Type),
	)
}

func syntaxEnumValue(enum *ast.EnumType, refs *syntaxRefs) interpreter.Value {
	id := nodeID("enum", enum.SourcePath, enum.Pos)
	refs.positions[id] = enum.Pos
	members := &interpreter.Array{}
	for i := range enum.Members {
		member := &enum.Members[i]
		memberID := nodeID("enumMember", enum.SourcePath, member.Pos)
		refs.positions[memberID] = member.Pos
		params := &interpreter.Array{}
		for j := range member.Params {
			params.Elements = append(params.Elements, syntaxParamValue(&member.Params[j], enum.SourcePath, refs))
		}
		members.Elements = append(members.Elements, structValue("SyntaxEnumMember",
			"id", memberID,
			"name", member.Name,
			"private", member.Private,
			"annotations", syntaxAnnotationsValue(member.Annotations, enum.SourcePath, refs),
			"value", member.Value,
			"hasValue", member.HasValue,
			"params", params,
		))
	}
	return structValue("SyntaxEnum",
		"id", id,
		"name", enum.Name,
		"private", enum.Private,
		"generics", stringArray(enum.Generics),
		"annotations", syntaxAnnotationsValue(enum.Annotations, enum.SourcePath, refs),
		"members", members,
		"sourcePath", enum.SourcePath,
	)
}

func syntaxFunctionValue(fn *ast.Function, refs *syntaxRefs) interpreter.Value {
	id := nodeID("function", fn.SourcePath, fn.Pos)
	bodyID := id + "/body"
	refs.positions[id] = fn.Pos
	refs.bodies[bodyID] = fn.Body
	params := &interpreter.Array{}
	for i := range fn.Params {
		params.Elements = append(params.Elements, syntaxParamValue(&fn.Params[i], fn.SourcePath, refs))
	}
	return structValue("SyntaxFunction",
		"id", id,
		"name", fn.Name,
		"private", fn.Private,
		"static", fn.Static,
		"routine", fn.Routine,
		"generics", stringArray(fn.Generics),
		"annotations", syntaxAnnotationsValue(fn.Annotations, fn.SourcePath, refs),
		"receiverType", fn.ReceiverType,
		"params", params,
		"returnType", syntaxTypeValue(fn.ReturnType),
		"bodyID", bodyID,
		"generatedBody", interpreter.NullValue,
		"sourcePath", fn.SourcePath,
	)
}

func syntaxParamValue(param *ast.Param, sourcePath string, refs *syntaxRefs) interpreter.Value {
	id := nodeID("param", sourcePath, param.Pos)
	refs.positions[id] = param.Pos
	return structValue("SyntaxParam",
		"id", id,
		"name", param.Name,
		"type", syntaxTypeValue(param.Type),
	)
}

func syntaxAnnotationsValue(annotations []ast.Annotation, sourcePath string, refs *syntaxRefs) interpreter.Value {
	out := &interpreter.Array{}
	for i := range annotations {
		annotation := annotations[i]
		id := nodeID("annotation", sourcePath, annotation.Pos)
		refs.annotations[id] = annotation
		refs.positions[id] = annotation.Pos
		out.Elements = append(out.Elements, structValue("SyntaxAnnotation",
			"id", id,
			"module", annotation.Module,
			"name", annotation.Name,
			"hasParens", annotation.HasParens,
			"args", syntaxAnnotationArgsValue(annotation.Args),
		))
	}
	return out
}

func syntaxAnnotationArgsValue(args []ast.Expr) interpreter.Value {
	out := &interpreter.Array{}
	for _, arg := range args {
		kind := "unsupported"
		text := ""
		number := 0
		flag := false
		switch value := arg.(type) {
		case *ast.StringLiteral:
			kind = "string"
			text = value.Value
		case *ast.IntegerLiteral:
			kind = "int"
			number = value.Value
		case *ast.BoolLiteral:
			kind = "bool"
			flag = value.Value
		case *ast.NullLiteral:
			kind = "null"
		}
		out.Elements = append(out.Elements, structValue("SyntaxValue",
			"kind", kind,
			"stringValue", text,
			"intValue", number,
			"boolValue", flag,
		))
	}
	return out
}

func syntaxTypeValue(typ ast.Type) interpreter.Value {
	args := &interpreter.Array{}
	for _, arg := range typ.Args {
		args.Elements = append(args.Elements, syntaxTypeValue(arg))
	}
	params := &interpreter.Array{}
	for _, param := range typ.Params {
		params.Elements = append(params.Elements, structValue("SyntaxTypeParam",
			"name", param.Name,
			"optional", param.Optional,
			"type", syntaxTypeValue(param.Type),
		))
	}
	var returnType interpreter.Value = interpreter.NullValue
	if typ.Return != nil {
		returnType = syntaxTypeValue(*typ.Return)
	}
	var element interpreter.Value = interpreter.NullValue
	if typ.Elem != nil {
		element = syntaxTypeValue(*typ.Elem)
	}
	return structValue("SyntaxType",
		"kind", int(typ.Kind),
		"name", typ.Name,
		"module", typ.Module,
		"nullable", typ.Nullable,
		"args", args,
		"params", params,
		"returnType", returnType,
		"element", element,
		"raw", typ.Raw,
		"displayText", typ.DisplayText,
	)
}

func decodeSyntaxFile(value interpreter.Value, original *ast.File, refs *syntaxRefs) (*ast.File, error) {
	root, err := expectStruct(value, "SyntaxFile")
	if err != nil {
		return nil, err
	}
	out := &ast.File{
		Imports:   append([]ast.Import(nil), original.Imports...),
		GoImports: append([]ast.GoImport(nil), original.GoImports...),
		TSImports: append([]ast.TSImport(nil), original.TSImports...),
		Traits:    append([]*ast.TraitDecl(nil), original.Traits...),
		Tests:     append([]*ast.Test(nil), original.Tests...),
	}
	for _, fn := range original.Functions {
		if fn.Macro {
			out.Functions = append(out.Functions, fn)
		}
	}
	seen := map[string]bool{}
	typeValues, err := structArrayField(root, "types")
	if err != nil {
		return nil, err
	}
	for _, value := range typeValues {
		typ, err := decodeSyntaxStruct(value, refs, seen)
		if err != nil {
			return nil, err
		}
		out.Types = append(out.Types, typ)
	}
	enumValues, err := structArrayField(root, "enums")
	if err != nil {
		return nil, err
	}
	for _, value := range enumValues {
		enum, err := decodeSyntaxEnum(value, refs, seen)
		if err != nil {
			return nil, err
		}
		out.Enums = append(out.Enums, enum)
	}
	functionValues, err := structArrayField(root, "functions")
	if err != nil {
		return nil, err
	}
	for _, value := range functionValues {
		fn, err := decodeSyntaxFunction(value, refs, seen)
		if err != nil {
			return nil, err
		}
		out.Functions = append(out.Functions, fn)
	}
	return out, nil
}

func decodeSyntaxStruct(value interpreter.Value, refs *syntaxRefs, seen map[string]bool) (*ast.StructType, error) {
	node, err := expectStruct(value, "SyntaxStruct")
	if err != nil {
		return nil, err
	}
	id, pos, err := decodeIdentity(node, refs, seen)
	if err != nil {
		return nil, err
	}
	_ = id
	name, err := nonEmptyStringField(node, "name")
	if err != nil {
		return nil, err
	}
	private, err := boolField(node, "private")
	if err != nil {
		return nil, err
	}
	generics, err := stringArrayField(node, "generics")
	if err != nil {
		return nil, err
	}
	annotations, err := decodeAnnotations(node, refs, seen)
	if err != nil {
		return nil, err
	}
	sourcePath, err := stringField(node, "sourcePath")
	if err != nil {
		return nil, err
	}
	out := &ast.StructType{Name: name, Private: private, Generics: generics, Annotations: annotations, Pos: pos, NamePos: pos, SourcePath: sourcePath}
	fields, err := structArrayField(node, "fields")
	if err != nil {
		return nil, err
	}
	for _, value := range fields {
		field, err := decodeSyntaxField(value, sourcePath, refs, seen)
		if err != nil {
			return nil, err
		}
		out.Fields = append(out.Fields, field)
	}
	methods, err := structArrayField(node, "methods")
	if err != nil {
		return nil, err
	}
	for _, value := range methods {
		method, err := decodeSyntaxFunction(value, refs, seen)
		if err != nil {
			return nil, err
		}
		method.ReceiverType = name
		out.Methods = append(out.Methods, method)
	}
	return out, nil
}

func decodeSyntaxField(value interpreter.Value, sourcePath string, refs *syntaxRefs, seen map[string]bool) (ast.Field, error) {
	node, err := expectStruct(value, "SyntaxField")
	if err != nil {
		return ast.Field{}, err
	}
	_, pos, err := decodeIdentity(node, refs, seen)
	if err != nil {
		return ast.Field{}, err
	}
	name, err := nonEmptyStringField(node, "name")
	if err != nil {
		return ast.Field{}, err
	}
	private, err := boolField(node, "private")
	if err != nil {
		return ast.Field{}, err
	}
	annotations, err := decodeAnnotations(node, refs, seen)
	if err != nil {
		return ast.Field{}, err
	}
	typ, err := decodeSyntaxType(node.Fields["type"])
	if err != nil {
		return ast.Field{}, err
	}
	return ast.Field{Name: name, Private: private, Annotations: annotations, Type: typ, Pos: pos}, nil
}

func decodeSyntaxEnum(value interpreter.Value, refs *syntaxRefs, seen map[string]bool) (*ast.EnumType, error) {
	node, err := expectStruct(value, "SyntaxEnum")
	if err != nil {
		return nil, err
	}
	_, pos, err := decodeIdentity(node, refs, seen)
	if err != nil {
		return nil, err
	}
	name, err := nonEmptyStringField(node, "name")
	if err != nil {
		return nil, err
	}
	private, err := boolField(node, "private")
	if err != nil {
		return nil, err
	}
	generics, err := stringArrayField(node, "generics")
	if err != nil {
		return nil, err
	}
	annotations, err := decodeAnnotations(node, refs, seen)
	if err != nil {
		return nil, err
	}
	sourcePath, err := stringField(node, "sourcePath")
	if err != nil {
		return nil, err
	}
	out := &ast.EnumType{Name: name, Private: private, Generics: generics, Annotations: annotations, Pos: pos, NamePos: pos, SourcePath: sourcePath}
	members, err := structArrayField(node, "members")
	if err != nil {
		return nil, err
	}
	for _, value := range members {
		memberNode, err := expectStruct(value, "SyntaxEnumMember")
		if err != nil {
			return nil, err
		}
		_, memberPos, err := decodeIdentity(memberNode, refs, seen)
		if err != nil {
			return nil, err
		}
		memberName, err := nonEmptyStringField(memberNode, "name")
		if err != nil {
			return nil, err
		}
		memberPrivate, err := boolField(memberNode, "private")
		if err != nil {
			return nil, err
		}
		memberAnnotations, err := decodeAnnotations(memberNode, refs, seen)
		if err != nil {
			return nil, err
		}
		value, err := intField(memberNode, "value")
		if err != nil {
			return nil, err
		}
		hasValue, err := boolField(memberNode, "hasValue")
		if err != nil {
			return nil, err
		}
		member := ast.EnumMember{Name: memberName, Private: memberPrivate, Annotations: memberAnnotations, Value: value, HasValue: hasValue, Pos: memberPos}
		params, err := structArrayField(memberNode, "params")
		if err != nil {
			return nil, err
		}
		for _, value := range params {
			param, err := decodeSyntaxParam(value, refs, seen)
			if err != nil {
				return nil, err
			}
			member.Params = append(member.Params, param)
		}
		out.Members = append(out.Members, member)
	}
	return out, nil
}

func decodeSyntaxFunction(value interpreter.Value, refs *syntaxRefs, seen map[string]bool) (*ast.Function, error) {
	node, err := expectStruct(value, "SyntaxFunction")
	if err != nil {
		return nil, err
	}
	_, pos, err := decodeIdentity(node, refs, seen)
	if err != nil {
		return nil, err
	}
	name, err := nonEmptyStringField(node, "name")
	if err != nil {
		return nil, err
	}
	private, err := boolField(node, "private")
	if err != nil {
		return nil, err
	}
	static, err := boolField(node, "static")
	if err != nil {
		return nil, err
	}
	routine, err := boolField(node, "routine")
	if err != nil {
		return nil, err
	}
	generics, err := stringArrayField(node, "generics")
	if err != nil {
		return nil, err
	}
	annotations, err := decodeAnnotations(node, refs, seen)
	if err != nil {
		return nil, err
	}
	receiverType, err := stringField(node, "receiverType")
	if err != nil {
		return nil, err
	}
	returnType, err := decodeSyntaxType(node.Fields["returnType"])
	if err != nil {
		return nil, err
	}
	bodyID, err := stringField(node, "bodyID")
	if err != nil {
		return nil, err
	}
	var body ast.Expr
	if generated := node.Fields["generatedBody"]; !isNull(generated) {
		body, err = decodeSyntaxExpr(generated)
		if err != nil {
			return nil, err
		}
	} else {
		var ok bool
		body, ok = refs.bodies[bodyID]
		if !ok {
			return nil, fmt.Errorf("SyntaxFunction.bodyID %q does not reference an existing function body", bodyID)
		}
	}
	sourcePath, err := stringField(node, "sourcePath")
	if err != nil {
		return nil, err
	}
	out := &ast.Function{
		Name: name, Private: private, Static: static, Routine: routine, Generics: generics,
		Annotations: annotations, ReceiverType: receiverType, ReturnType: returnType,
		Body: body, Pos: pos, NamePos: pos, SourcePath: sourcePath,
	}
	params, err := structArrayField(node, "params")
	if err != nil {
		return nil, err
	}
	for _, value := range params {
		param, err := decodeSyntaxParam(value, refs, seen)
		if err != nil {
			return nil, err
		}
		out.Params = append(out.Params, param)
	}
	return out, nil
}

func decodeSyntaxExpr(value interpreter.Value) (ast.Expr, error) {
	node, err := expectStruct(value, "SyntaxExpr")
	if err != nil {
		return nil, err
	}
	kind, err := stringField(node, "kind")
	if err != nil {
		return nil, err
	}
	name, err := stringField(node, "name")
	if err != nil {
		return nil, err
	}
	text, err := stringField(node, "stringValue")
	if err != nil {
		return nil, err
	}
	flag, err := boolField(node, "boolValue")
	if err != nil {
		return nil, err
	}
	args, err := structArrayField(node, "args")
	if err != nil {
		return nil, err
	}
	switch kind {
	case "identifier":
		return &ast.Identifier{Name: name}, nil
	case "module":
		return &ast.AtExpr{Name: name}, nil
	case "string":
		return &ast.StringLiteral{Value: text}, nil
	case "bool":
		return &ast.BoolLiteral{Value: flag}, nil
	case "null":
		return &ast.NullLiteral{}, nil
	case "selector":
		if len(args) != 1 {
			return nil, fmt.Errorf("selector SyntaxExpr expects one receiver")
		}
		receiver, err := decodeSyntaxExpr(args[0])
		if err != nil {
			return nil, err
		}
		return &ast.SelectorExpr{Receiver: receiver, Name: name}, nil
	case "staticSelector":
		if len(args) != 1 {
			return nil, fmt.Errorf("static selector SyntaxExpr expects one receiver")
		}
		receiver, err := decodeSyntaxExpr(args[0])
		if err != nil {
			return nil, err
		}
		return &ast.SelectorExpr{Receiver: receiver, Name: name, Static: true}, nil
	case "call":
		if len(args) == 0 {
			return nil, fmt.Errorf("call SyntaxExpr expects a callee")
		}
		callee, err := decodeSyntaxExpr(args[0])
		if err != nil {
			return nil, err
		}
		call := &ast.CallExpr{Callee: callee}
		for _, value := range args[1:] {
			arg, err := decodeSyntaxExpr(value)
			if err != nil {
				return nil, err
			}
			call.Args = append(call.Args, arg)
		}
		return call, nil
	case "struct":
		lit := &ast.StructLiteral{TypeName: name}
		fields, err := structArrayField(node, "fields")
		if err != nil {
			return nil, err
		}
		for _, value := range fields {
			field, err := expectStruct(value, "SyntaxExprField")
			if err != nil {
				return nil, err
			}
			fieldName, err := nonEmptyStringField(field, "name")
			if err != nil {
				return nil, err
			}
			fieldValue, err := decodeSyntaxExpr(field.Fields["value"])
			if err != nil {
				return nil, err
			}
			lit.Fields = append(lit.Fields, ast.FieldValue{Name: fieldName, Value: fieldValue})
		}
		return lit, nil
	case "block":
		block := &ast.BlockExpr{}
		statements, err := structArrayField(node, "statements")
		if err != nil {
			return nil, err
		}
		for _, value := range statements {
			statement, err := expectStruct(value, "SyntaxStatement")
			if err != nil {
				return nil, err
			}
			statementKind, err := stringField(statement, "kind")
			if err != nil {
				return nil, err
			}
			statementName, err := stringField(statement, "name")
			if err != nil {
				return nil, err
			}
			statementValue, err := decodeSyntaxExpr(statement.Fields["value"])
			if err != nil {
				return nil, err
			}
			switch statementKind {
			case "let":
				block.Statements = append(block.Statements, &ast.LetStmt{Name: statementName, Value: statementValue})
			case "expr":
				block.Statements = append(block.Statements, &ast.ExprStmt{Expr: statementValue})
			default:
				return nil, fmt.Errorf("unsupported SyntaxStatement kind %q", statementKind)
			}
		}
		return block, nil
	default:
		return nil, fmt.Errorf("unsupported SyntaxExpr kind %q", kind)
	}
}

func decodeSyntaxParam(value interpreter.Value, refs *syntaxRefs, seen map[string]bool) (ast.Param, error) {
	node, err := expectStruct(value, "SyntaxParam")
	if err != nil {
		return ast.Param{}, err
	}
	_, pos, err := decodeIdentity(node, refs, seen)
	if err != nil {
		return ast.Param{}, err
	}
	name, err := nonEmptyStringField(node, "name")
	if err != nil {
		return ast.Param{}, err
	}
	typ, err := decodeSyntaxType(node.Fields["type"])
	if err != nil {
		return ast.Param{}, err
	}
	return ast.Param{Name: name, Type: typ, Pos: pos}, nil
}

func decodeAnnotations(node *interpreter.Struct, refs *syntaxRefs, seen map[string]bool) ([]ast.Annotation, error) {
	values, err := structArrayField(node, "annotations")
	if err != nil {
		return nil, err
	}
	out := make([]ast.Annotation, 0, len(values))
	for _, value := range values {
		annotationNode, err := expectStruct(value, "SyntaxAnnotation")
		if err != nil {
			return nil, err
		}
		id, _, err := decodeIdentity(annotationNode, refs, seen)
		if err != nil {
			return nil, err
		}
		annotation, ok := refs.annotations[id]
		if id != "" && !ok {
			return nil, fmt.Errorf("SyntaxAnnotation.id %q does not reference an existing annotation", id)
		}
		module, err := stringField(annotationNode, "module")
		if err != nil {
			return nil, err
		}
		name, err := nonEmptyStringField(annotationNode, "name")
		if err != nil {
			return nil, err
		}
		hasParens, err := boolField(annotationNode, "hasParens")
		if err != nil {
			return nil, err
		}
		annotation.Module = module
		annotation.Name = name
		annotation.HasParens = hasParens
		out = append(out, annotation)
	}
	return out, nil
}

func decodeSyntaxType(value interpreter.Value) (ast.Type, error) {
	node, err := expectStruct(value, "SyntaxType")
	if err != nil {
		return ast.Type{}, err
	}
	kind, err := intField(node, "kind")
	if err != nil {
		return ast.Type{}, err
	}
	name, err := stringField(node, "name")
	if err != nil {
		return ast.Type{}, err
	}
	module, err := stringField(node, "module")
	if err != nil {
		return ast.Type{}, err
	}
	nullable, err := boolField(node, "nullable")
	if err != nil {
		return ast.Type{}, err
	}
	raw, err := stringField(node, "raw")
	if err != nil {
		return ast.Type{}, err
	}
	displayText, err := stringField(node, "displayText")
	if err != nil {
		return ast.Type{}, err
	}
	out := ast.Type{Kind: ast.TypeKind(kind), Name: name, Module: module, Nullable: nullable, Raw: raw, DisplayText: displayText}
	args, err := structArrayField(node, "args")
	if err != nil {
		return ast.Type{}, err
	}
	for _, value := range args {
		arg, err := decodeSyntaxType(value)
		if err != nil {
			return ast.Type{}, err
		}
		out.Args = append(out.Args, arg)
	}
	params, err := structArrayField(node, "params")
	if err != nil {
		return ast.Type{}, err
	}
	for _, value := range params {
		paramNode, err := expectStruct(value, "SyntaxTypeParam")
		if err != nil {
			return ast.Type{}, err
		}
		paramName, err := stringField(paramNode, "name")
		if err != nil {
			return ast.Type{}, err
		}
		optional, err := boolField(paramNode, "optional")
		if err != nil {
			return ast.Type{}, err
		}
		paramType, err := decodeSyntaxType(paramNode.Fields["type"])
		if err != nil {
			return ast.Type{}, err
		}
		out.Params = append(out.Params, ast.TypeParam{Name: paramName, Optional: optional, Type: paramType})
	}
	if value := node.Fields["returnType"]; !isNull(value) {
		typ, err := decodeSyntaxType(value)
		if err != nil {
			return ast.Type{}, err
		}
		out.Return = &typ
	}
	if value := node.Fields["element"]; !isNull(value) {
		typ, err := decodeSyntaxType(value)
		if err != nil {
			return ast.Type{}, err
		}
		out.Elem = &typ
	}
	return out, nil
}

func targetSyntaxID(target Target) string {
	return nodeID(string(target.Kind), target.SourcePath, target.Pos)
}

func nodeID(kind string, sourcePath string, pos lexer.Position) string {
	return kind + ":" + sourcePath + ":" + strconv.Itoa(pos.Offset)
}

func decodeIdentity(node *interpreter.Struct, refs *syntaxRefs, seen map[string]bool) (string, lexer.Position, error) {
	id, err := stringField(node, "id")
	if err != nil {
		return "", lexer.Position{}, err
	}
	if id == "" {
		if refs.generated == 0 {
			refs.generated = -1
			for _, pos := range refs.positions {
				if pos.Offset <= refs.generated {
					refs.generated = pos.Offset - 1
				}
			}
		} else {
			refs.generated--
		}
		return "", lexer.Position{Offset: refs.generated}, nil
	}
	if seen[id] {
		return "", lexer.Position{}, fmt.Errorf("duplicate syntax id %q", id)
	}
	seen[id] = true
	pos, ok := refs.positions[id]
	if !ok {
		return "", lexer.Position{}, fmt.Errorf("%s.id %q does not reference an existing syntax node", node.TypeName, id)
	}
	return id, pos, nil
}

func structValue(typeName string, fields ...interpreter.Value) *interpreter.Struct {
	out := &interpreter.Struct{TypeName: typeName, Fields: map[string]interpreter.Value{}}
	for i := 0; i < len(fields); i += 2 {
		name := fields[i].(string)
		out.Fields[name] = fields[i+1]
		out.Order = append(out.Order, name)
	}
	return out
}

func stringArray(values []string) *interpreter.Array {
	out := &interpreter.Array{Elements: make([]interpreter.Value, 0, len(values))}
	for _, value := range values {
		out.Elements = append(out.Elements, value)
	}
	return out
}

func expectStruct(value interpreter.Value, typeName string) (*interpreter.Struct, error) {
	node, ok := value.(*interpreter.Struct)
	if !ok || node.TypeName != typeName {
		return nil, fmt.Errorf("expected %s, got %T", typeName, value)
	}
	return node, nil
}

func structArrayField(node *interpreter.Struct, name string) ([]interpreter.Value, error) {
	value, ok := node.Fields[name]
	if !ok {
		return nil, fmt.Errorf("%s.%s is missing", node.TypeName, name)
	}
	array, ok := value.(*interpreter.Array)
	if !ok {
		return nil, fmt.Errorf("%s.%s must be an Array", node.TypeName, name)
	}
	return array.Elements, nil
}

func stringArrayField(node *interpreter.Struct, name string) ([]string, error) {
	values, err := structArrayField(node, name)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("%s.%s must contain only Strings", node.TypeName, name)
		}
		out = append(out, text)
	}
	return out, nil
}

func stringField(node *interpreter.Struct, name string) (string, error) {
	value, ok := node.Fields[name]
	if !ok {
		return "", fmt.Errorf("%s.%s is missing", node.TypeName, name)
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s.%s must be a String", node.TypeName, name)
	}
	return text, nil
}

func nonEmptyStringField(node *interpreter.Struct, name string) (string, error) {
	value, err := stringField(node, name)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", fmt.Errorf("%s.%s must not be empty", node.TypeName, name)
	}
	return value, nil
}

func boolField(node *interpreter.Struct, name string) (bool, error) {
	value, ok := node.Fields[name]
	if !ok {
		return false, fmt.Errorf("%s.%s is missing", node.TypeName, name)
	}
	flag, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("%s.%s must be a Bool", node.TypeName, name)
	}
	return flag, nil
}

func intField(node *interpreter.Struct, name string) (int, error) {
	value, ok := node.Fields[name]
	if !ok {
		return 0, fmt.Errorf("%s.%s is missing", node.TypeName, name)
	}
	number, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("%s.%s must be an Int", node.TypeName, name)
	}
	return number, nil
}

func isNull(value interpreter.Value) bool {
	return value == interpreter.NullValue
}
