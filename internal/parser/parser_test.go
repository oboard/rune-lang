package parser

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func TestGenericTypeDeclRequiresColon(t *testing.T) {
	file, errs := Parse(`Array[T] {
	    length[T]() -> Int => "%array.len"
}
`)
	if len(errs) == 0 {
		t.Fatalf("Parse() accepted generic type declaration without ':'")
	}
	if len(file.Types) != 0 {
		t.Fatalf("Parse() produced %d types for invalid declaration", len(file.Types))
	}
}

func TestParseTraitDeclarationAndReference(t *testing.T) {
	file, errs := Parse(`&ToJson: {
  name: String
  toJson(pretty: Bool) -> Self
}

encode(value: &ToJson) -> &ToJson => value
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Traits) != 1 {
		t.Fatalf("traits = %#v, want one trait", file.Traits)
	}
	trait := file.Traits[0]
	if trait.Name != "ToJson" || len(trait.Fields) != 1 || len(trait.Methods) != 1 {
		t.Fatalf("trait = %#v", trait)
	}
	if got := file.Functions[0].Params[0].Type.Canonical(); got != "&ToJson" {
		t.Fatalf("parameter type = %q, want &ToJson", got)
	}
}

func TestParseStaticTraitAndStructMethods(t *testing.T) {
	file, errs := Parse(`&FromJson: {
  ::fromJson(text: String) -> Self
}

User: {
  name: String
  ::fromJson(text: String) -> User => User { name: text }
}

main() => User::fromJson("{}")
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if !file.Traits[0].Methods[0].Static {
		t.Fatal("trait method should be static")
	}
	if !file.Types[0].Methods[0].Static {
		t.Fatal("struct method should be static")
	}
	call := file.Functions[0].Body.(*ast.CallExpr)
	selector := call.Callee.(*ast.SelectorExpr)
	if !selector.Static {
		t.Fatal("static method call selector should use ::")
	}
}

func TestParseTypedBinding(t *testing.T) {
	file, errs := Parse(`main() => {
  user := @json.parse(text) : User
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	body := file.Functions[0].Body.(*ast.BlockExpr)
	let := body.Statements[0].(*ast.LetStmt)
	if got := let.Type.Canonical(); got != "User" {
		t.Fatalf("binding type = %q, want User", got)
	}
}

func TestParseStructLiteralSpreadField(t *testing.T) {
	file, errs := Parse(`User: {
  name: String
  age: Int
}

main(existing: User) => User {
  ...existing,
  age: 42,
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	lit := file.Functions[0].Body.(*ast.StructLiteral)
	if len(lit.Fields) != 2 {
		t.Fatalf("fields = %#v, want spread and age", lit.Fields)
	}
	if !lit.Fields[0].Spread {
		t.Fatalf("first field = %#v, want spread", lit.Fields[0])
	}
	if _, ok := lit.Fields[0].Value.(*ast.Identifier); !ok {
		t.Fatalf("spread value = %T, want Identifier", lit.Fields[0].Value)
	}
	if lit.Fields[1].Name != "age" {
		t.Fatalf("second field = %#v, want age", lit.Fields[1])
	}
}

func TestParseStructLiteralRequiresCommaBetweenFields(t *testing.T) {
	file, errs := Parse(`User: {
  name: String
  age: Int
}

main() => User {
  name: "oboard"
  age: 42
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	lit := file.Functions[0].Body.(*ast.StructLiteral)
	if len(lit.Fields) != 2 || !lit.Fields[0].MissingComma {
		t.Fatalf("fields = %#v, want first field marked missing comma", lit.Fields)
	}
}

func TestGenericTypeDeclWithColon(t *testing.T) {
	file, errs := Parse(`Array[T]: {
    length[T]() -> Int => "%array.len"
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Types) != 1 {
		t.Fatalf("Parse() produced %d types, want 1", len(file.Types))
	}
	if got := file.Types[0].Generics; len(got) != 1 || got[0] != "T" {
		t.Fatalf("type generics = %v, want [T]", got)
	}
}

func TestParseGenericConstraints(t *testing.T) {
	file, errs := Parse(`Box[T: &Named]: {
  value: T
}

add[T: Number](a: T, b: T) -> T => a + b
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if got := file.Types[0].Generics; len(got) != 1 || got[0] != "T" {
		t.Fatalf("type generics = %v, want [T]", got)
	}
	if got := file.Types[0].GenericConstraints["T"].Canonical(); got != "&Named" {
		t.Fatalf("type generic constraint = %q, want &Named", got)
	}
	if got := file.Functions[0].Generics; len(got) != 1 || got[0] != "T" {
		t.Fatalf("function generics = %v, want [T]", got)
	}
	if got := file.Functions[0].GenericConstraints["T"].Canonical(); got != "Number" {
		t.Fatalf("function generic constraint = %q, want Number", got)
	}
	if got := file.Functions[0].Signature(); got != "add[T: Number](a: T, b: T)" {
		t.Fatalf("signature = %q, want constrained generic signature", got)
	}
}

func TestRuneImportDecl(t *testing.T) {
	file, errs := Parse(`@"./helper.rn"

main() => helper()
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Imports) != 1 {
		t.Fatalf("imports = %d, want 1", len(file.Imports))
	}
	if file.Imports[0].Path != "./helper.rn" {
		t.Fatalf("import path = %q, want ./helper.rn", file.Imports[0].Path)
	}
	if len(file.Functions) != 1 || file.Functions[0].Name != "main" {
		t.Fatalf("functions = %#v, want main", file.Functions)
	}
}

func TestDeclarationAnnotationsPreserveCalls(t *testing.T) {
	file, errs := Parse(`#cli.command("ship")
Args: {
  #cli.flag("v", "show verbose output")
  verbose: Bool
}

#derive() => null
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Types) != 1 || len(file.Types[0].Annotations) != 1 {
		t.Fatalf("type annotations = %#v, want one annotation", file.Types)
	}
	command := file.Types[0].Annotations[0]
	if command.Module != "cli" || command.Name != "command" || len(command.Args) != 1 {
		t.Fatalf("command annotation = %#v", command)
	}
	if len(file.Types[0].Fields) != 1 || len(file.Types[0].Fields[0].Annotations) != 1 {
		t.Fatalf("field annotations = %#v, want one annotation", file.Types[0].Fields)
	}
	flag := file.Types[0].Fields[0].Annotations[0]
	if flag.Module != "cli" || flag.Name != "flag" || len(flag.Args) != 2 {
		t.Fatalf("flag annotation = %#v", flag)
	}
	if len(file.Functions) != 1 || !file.Functions[0].Macro {
		t.Fatalf("functions = %#v, want one macro function", file.Functions)
	}
	if len(file.Functions[0].Annotations) != 0 {
		t.Fatalf("macro bootstrap marker leaked into annotations: %#v", file.Functions[0].Annotations)
	}
}

func TestMacroDefinitionRequiresParameterList(t *testing.T) {
	file, errs := Parse(`#derive => null
`)
	if len(errs) == 0 {
		t.Fatal("Parse() accepted a macro definition without parameter parentheses")
	}
	if len(file.Functions) != 0 {
		t.Fatalf("functions = %#v, want no parsed macro definition", file.Functions)
	}
}

func TestAnnotatedEnumMembersRemainEnum(t *testing.T) {
	file, errs := Parse(`#cli.subcommands
Command: {
  #cli.command("clone")
  Clone(options: CloneOptions)
  #cli.command("status")
  Status
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Enums) != 1 || len(file.Enums[0].Members) != 2 {
		t.Fatalf("enums = %#v, want annotated Command enum", file.Enums)
	}
	for _, member := range file.Enums[0].Members {
		if len(member.Annotations) != 1 || member.Annotations[0].Module != "cli" {
			t.Fatalf("member annotation = %#v", member.Annotations)
		}
	}
}

func TestDeclarationVisibility(t *testing.T) {
	file, errs := Parse(`Secret: {
  exposed: Int
  - value: Int

  get() -> Int => .exposed
  - reveal() -> Int => .value
}

+ Status: {
  Ready = 0
  + Hidden = 1
}

+ add(a: Int, b: Int) -> Int => a + b
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Types) != 1 || !file.Types[0].Private {
		t.Fatalf("types = %#v, want one private type", file.Types)
	}
	if len(file.Types[0].Fields) != 2 || file.Types[0].Fields[0].Private || !file.Types[0].Fields[1].Private {
		t.Fatalf("fields = %#v, want public exposed and private value", file.Types[0].Fields)
	}
	if len(file.Types[0].Methods) != 2 || file.Types[0].Methods[0].Private || !file.Types[0].Methods[1].Private {
		t.Fatalf("methods = %#v, want public get and private reveal", file.Types[0].Methods)
	}
	if len(file.Enums) != 1 || file.Enums[0].Private {
		t.Fatalf("enums = %#v, want one public enum", file.Enums)
	}
	if len(file.Enums[0].Members) != 2 || file.Enums[0].Members[0].Private || file.Enums[0].Members[1].Private {
		t.Fatalf("members = %#v, want public enum members by default", file.Enums[0].Members)
	}
	if len(file.Functions) != 1 || file.Functions[0].Private {
		t.Fatalf("functions = %#v, want public add", file.Functions)
	}
}

func TestEnumDeclWithIntegerMembers(t *testing.T) {
	file, errs := Parse(`Status: {
  Completed = 0
  Fail = 1
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Enums) != 1 {
		t.Fatalf("Parse() produced %d enums, want 1", len(file.Enums))
	}
	enum := file.Enums[0]
	if enum.Name != "Status" || len(enum.Members) != 2 {
		t.Fatalf("enum = %#v, want Status with two members", enum)
	}
	if enum.Members[0].Name != "Completed" || enum.Members[0].Value != 0 {
		t.Fatalf("first enum member = %#v, want Completed = 0", enum.Members[0])
	}
	if enum.Members[1].Name != "Fail" || enum.Members[1].Value != 1 {
		t.Fatalf("second enum member = %#v, want Fail = 1", enum.Members[1])
	}
}

func TestEnumDeclWithBareMembers(t *testing.T) {
	file, errs := Parse(`TokenKind: {
  EOF
  Ident
  Int
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Enums) != 1 {
		t.Fatalf("Parse() produced %d enums, want 1", len(file.Enums))
	}
	enum := file.Enums[0]
	if enum.Name != "TokenKind" || len(enum.Members) != 3 {
		t.Fatalf("enum = %#v, want TokenKind with three members", enum)
	}
	if enum.Members[0].Name != "EOF" || enum.Members[0].HasValue {
		t.Fatalf("first enum member = %#v, want bare EOF", enum.Members[0])
	}
	if enum.Members[2].Name != "Int" || enum.Members[2].HasValue {
		t.Fatalf("third enum member = %#v, want bare Int", enum.Members[2])
	}
}

func TestGenericEnumDeclWithConstructors(t *testing.T) {
	file, errs := Parse(`Result[T, E]: {
  Ok(value: T)
  Err(error: E)
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Enums) != 1 {
		t.Fatalf("Parse() produced %d enums, want 1", len(file.Enums))
	}
	enum := file.Enums[0]
	if enum.Name != "Result" || len(enum.Generics) != 2 || enum.Generics[0] != "T" || enum.Generics[1] != "E" {
		t.Fatalf("enum = %#v, want generic Result[T, E]", enum)
	}
	if len(enum.Members) != 2 {
		t.Fatalf("members = %#v, want Ok and Err", enum.Members)
	}
	if enum.Members[0].HasValue || enum.Members[0].Name != "Ok" || len(enum.Members[0].Params) != 1 || enum.Members[0].Params[0].Type.Canonical() != "T" {
		t.Fatalf("first member = %#v, want Ok(value: T)", enum.Members[0])
	}
	if enum.Members[1].HasValue || enum.Members[1].Name != "Err" || len(enum.Members[1].Params) != 1 || enum.Members[1].Params[0].Type.Canonical() != "E" {
		t.Fatalf("second member = %#v, want Err(error: E)", enum.Members[1])
	}
}

func TestFunctionDeclRequiresFatArrow(t *testing.T) {
	file, errs := Parse(`main() {
}
`)
	if len(errs) == 0 {
		t.Fatalf("Parse() accepted function declaration without '=>'")
	}
	if len(file.Functions) != 0 {
		t.Fatalf("Parse() produced %d functions for invalid declaration", len(file.Functions))
	}
}

func TestLambdaRequiresParenthesizedParams(t *testing.T) {
	_, errs := Parse(`main() => {
    values.map(value => value + 1)
}
`)
	if len(errs) == 0 {
		t.Fatalf("Parse() accepted lambda without parenthesized params")
	}
}

func TestLambdaWithParenthesizedParams(t *testing.T) {
	_, errs := Parse(`main() => {
    values.map((value) => value + 1)
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
}

func TestLambdaWithAnnotatedParams(t *testing.T) {
	file, errs := Parse(`Input: {
    a: Int
}

main() => {
    f := (value: Input) => value.a + 1
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 1 {
		t.Fatalf("main body = %#v, want one statement", file.Functions[0].Body)
	}
	let, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("statement = %T, want LetStmt", block.Statements[0])
	}
	lambda, ok := let.Value.(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("let value = %T, want LambdaExpr", let.Value)
	}
	if len(lambda.ParamTypes) != 1 || lambda.ParamTypes[0].Canonical() != "Input" {
		t.Fatalf("lambda param types = %v, want [Input]", lambda.ParamTypes)
	}
}

func TestRoutineDeclAndResultUnwrap(t *testing.T) {
	file, errs := Parse(`~ read() => {
    file := @fs.readFile("1.txt")?
    file
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Functions) != 1 || !file.Functions[0].Routine {
		t.Fatalf("function = %#v, want one routine function", file.Functions)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 2 {
		t.Fatalf("routine body = %#v, want two statements", file.Functions[0].Body)
	}
	let, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("first statement = %T, want LetStmt", block.Statements[0])
	}
	if _, ok := let.Value.(*ast.ResultUnwrapExpr); !ok {
		t.Fatalf("let value = %T, want ResultUnwrapExpr", let.Value)
	}
}

func TestRoutineDeclWithQualifiedGenericReturn(t *testing.T) {
	file, errs := Parse(`~ gzip(data: @io.Data) -> Result[@io.Data, Error] => "%compress.gzip"`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Functions) != 1 {
		t.Fatalf("Parse() functions = %d, want 1", len(file.Functions))
	}
	fn := file.Functions[0]
	if !fn.Routine || fn.Name != "gzip" {
		t.Fatalf("function = %#v, want routine gzip", fn)
	}
	if len(fn.Params) != 1 || fn.Params[0].Type.Canonical() != "Data" || fn.Params[0].Type.Display() != "@io.Data" {
		t.Fatalf("params = %#v, want data: @io.Data", fn.Params)
	}
	if fn.Params[0].Type.Kind != ast.TypeName || fn.Params[0].Type.Module != "io" || fn.Params[0].Type.Name != "Data" {
		t.Fatalf("param type node = %#v, want qualified Data type", fn.Params[0].Type)
	}
	if fn.ReturnType.Canonical() != "Result[Data,Error]" || fn.ReturnType.Display() != "Result[@io.Data, Error]" {
		t.Fatalf("return = %q/%q, want Result[@io.Data, Error]", fn.ReturnType.Canonical(), fn.ReturnType.Display())
	}
	if fn.ReturnType.Kind != ast.TypeName || fn.ReturnType.Name != "Result" || len(fn.ReturnType.Args) != 2 || fn.ReturnType.Args[0].Module != "io" {
		t.Fatalf("return type node = %#v, want Result with qualified Data argument", fn.ReturnType)
	}
}

func TestConstructorPatternMatch(t *testing.T) {
	file, errs := Parse(`main() => readUser() {
    Ok(user) => user.name
    Err(e) => e.message
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Functions) != 1 {
		t.Fatalf("functions = %d, want 1", len(file.Functions))
	}
	match, ok := file.Functions[0].Body.(*ast.MatchExpr)
	if !ok || len(match.Branches) != 2 {
		t.Fatalf("body = %#v, want constructor pattern match", file.Functions[0].Body)
	}
	first, ok := match.Branches[0].Pattern.(*ast.ConstructorPattern)
	if !ok || first.Name != "Ok" || first.Binding != "user" {
		t.Fatalf("first pattern = %#v, want Ok(user)", match.Branches[0].Pattern)
	}
	second, ok := match.Branches[1].Pattern.(*ast.ConstructorPattern)
	if !ok || second.Name != "Err" || second.Binding != "e" {
		t.Fatalf("second pattern = %#v, want Err(e)", match.Branches[1].Pattern)
	}
}

func TestRangePatternMatch(t *testing.T) {
	file, errs := Parse(`isDigit(ch: Char) -> Bool => ch {
  '0'..='9' => true
  _ => false
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	match, ok := file.Functions[0].Body.(*ast.MatchExpr)
	if !ok || len(match.Branches) != 2 {
		t.Fatalf("body = %#v, want range pattern match", file.Functions[0].Body)
	}
	pattern, ok := match.Branches[0].Pattern.(*ast.RangePattern)
	if !ok {
		t.Fatalf("first pattern = %#v, want RangePattern", match.Branches[0].Pattern)
	}
	if _, ok := pattern.Start.(*ast.CharLiteral); !ok {
		t.Fatalf("range start = %#v, want CharLiteral", pattern.Start)
	}
	if _, ok := pattern.End.(*ast.CharLiteral); !ok {
		t.Fatalf("range end = %#v, want CharLiteral", pattern.End)
	}
}

func TestOrPatternMatch(t *testing.T) {
	file, errs := Parse(`typeLabel(name: String) -> String => name {
  "" | "Void" => "void"
  "Int" | "Double" => "number"
  _ => "other"
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	match, ok := file.Functions[0].Body.(*ast.MatchExpr)
	if !ok || len(match.Branches) != 3 {
		t.Fatalf("body = %#v, want or pattern match", file.Functions[0].Body)
	}
	pattern, ok := match.Branches[0].Pattern.(*ast.OrPattern)
	if !ok || len(pattern.Alternatives) != 2 {
		t.Fatalf("first pattern = %#v, want two-way OrPattern", match.Branches[0].Pattern)
	}
}

func TestObjectAndMapPatternMatch(t *testing.T) {
	file, errs := Parse(`pointScore(point) => point {
  { x, y: yy, .. } => x + yy
  _ => 0
}

mapScore(values) => values {
  { "a": value, .. } => value
  _ => 0
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	pointMatch, ok := file.Functions[0].Body.(*ast.MatchExpr)
	if !ok || len(pointMatch.Branches) != 2 {
		t.Fatalf("point body = %#v, want object pattern match", file.Functions[0].Body)
	}
	objectPattern, ok := pointMatch.Branches[0].Pattern.(*ast.ObjectPattern)
	if !ok || !objectPattern.Rest || len(objectPattern.Fields) != 2 {
		t.Fatalf("point pattern = %#v, want object pattern with rest", pointMatch.Branches[0].Pattern)
	}
	if binding, ok := objectPattern.Fields[0].Pattern.(*ast.BindingPattern); !ok || binding.Name != "x" {
		t.Fatalf("first object field = %#v, want x binding", objectPattern.Fields[0].Pattern)
	}
	if binding, ok := objectPattern.Fields[1].Pattern.(*ast.BindingPattern); !ok || binding.Name != "yy" {
		t.Fatalf("second object field = %#v, want yy binding", objectPattern.Fields[1].Pattern)
	}

	mapMatch, ok := file.Functions[1].Body.(*ast.MatchExpr)
	if !ok || len(mapMatch.Branches) != 2 {
		t.Fatalf("map body = %#v, want map pattern match", file.Functions[1].Body)
	}
	mapPattern, ok := mapMatch.Branches[0].Pattern.(*ast.MapPattern)
	if !ok || !mapPattern.Rest || len(mapPattern.Entries) != 1 {
		t.Fatalf("map pattern = %#v, want map pattern with rest", mapMatch.Branches[0].Pattern)
	}
	if binding, ok := mapPattern.Entries[0].Pattern.(*ast.BindingPattern); !ok || binding.Name != "value" {
		t.Fatalf("map value pattern = %#v, want value binding", mapPattern.Entries[0].Pattern)
	}
}

func TestNullCoalesceExpression(t *testing.T) {
	file, errs := Parse(`fallback(value: String?) -> String => value ?? "missing"
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	expr, ok := file.Functions[0].Body.(*ast.BinaryExpr)
	if !ok || expr.Op != lexer.QuestionQuestion {
		t.Fatalf("body = %#v, want ?? binary expression", file.Functions[0].Body)
	}
	if _, ok := expr.Left.(*ast.Identifier); !ok {
		t.Fatalf("left = %#v, want identifier", expr.Left)
	}
	if _, ok := expr.Right.(*ast.StringLiteral); !ok {
		t.Fatalf("right = %#v, want string literal", expr.Right)
	}
}

func TestTemplateLiteral(t *testing.T) {
	file, errs := Parse("greet(name: String) -> String => `Hello, \\(name)`\n")
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	lit, ok := file.Functions[0].Body.(*ast.TemplateLiteral)
	if !ok {
		t.Fatalf("body = %T, want TemplateLiteral", file.Functions[0].Body)
	}
	if len(lit.Parts) != 2 {
		t.Fatalf("parts = %#v, want text and expression", lit.Parts)
	}
	if lit.Parts[0].Text != "Hello, " {
		t.Fatalf("first part = %#v, want text", lit.Parts[0])
	}
	ident, ok := lit.Parts[1].Expr.(*ast.Identifier)
	if !ok || ident.Name != "name" {
		t.Fatalf("second part = %#v, want name identifier", lit.Parts[1])
	}
	if ident.Pos.Line != 1 || ident.Pos.Column != 44 {
		t.Fatalf("template identifier position = %s, want 1:44", ident.Pos)
	}
}

func TestMultilineTemplateLiteral(t *testing.T) {
	file, errs := Parse("message() -> String => `Hello\nRune`\n")
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	lit, ok := file.Functions[0].Body.(*ast.TemplateLiteral)
	if !ok {
		t.Fatalf("body = %T, want TemplateLiteral", file.Functions[0].Body)
	}
	if len(lit.Parts) != 1 || lit.Parts[0].Text != "Hello\nRune" {
		t.Fatalf("parts = %#v, want multiline text", lit.Parts)
	}
}

func TestTemplateLiteralExpressionPositionsAcrossLines(t *testing.T) {
	file, errs := Parse("greet(name: String) -> String => `Hello\n\\(name)`\n")
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	lit, ok := file.Functions[0].Body.(*ast.TemplateLiteral)
	if !ok || len(lit.Parts) != 2 {
		t.Fatalf("body = %#v, want template with expression", file.Functions[0].Body)
	}
	ident, ok := lit.Parts[1].Expr.(*ast.Identifier)
	if !ok || ident.Name != "name" {
		t.Fatalf("second part = %#v, want name identifier", lit.Parts[1])
	}
	if ident.Pos.Line != 2 || ident.Pos.Column != 3 {
		t.Fatalf("multiline template identifier position = %s, want 2:3", ident.Pos)
	}
}

func TestRegexLiteral(t *testing.T) {
	file, errs := Parse(`main() => {
    re := /rune\s+(\d+)/ig
    value := 10 / 2
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 2 {
		t.Fatalf("main body = %#v, want two statements", file.Functions[0].Body)
	}
	let, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("first statement = %T, want LetStmt", block.Statements[0])
	}
	regex, ok := let.Value.(*ast.RegexLiteral)
	if !ok {
		t.Fatalf("let value = %T, want RegexLiteral", let.Value)
	}
	if regex.Pattern != `rune\s+(\d+)` || regex.Flags != "ig" {
		t.Fatalf("regex = /%s/%s, want pattern and flags", regex.Pattern, regex.Flags)
	}
}

func TestRegexLiteralWithSlashInClassAndEscape(t *testing.T) {
	file, errs := Parse(`main() => /[\/a-z]+/g`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	regex, ok := file.Functions[0].Body.(*ast.RegexLiteral)
	if !ok {
		t.Fatalf("body = %T, want RegexLiteral", file.Functions[0].Body)
	}
	if regex.Pattern != `[\/a-z]+` || regex.Flags != "g" {
		t.Fatalf("regex = /%s/%s, want escaped slash pattern", regex.Pattern, regex.Flags)
	}
}

func TestFunctionTypeDisplayPreservesNamedParams(t *testing.T) {
	file, errs := Parse(`Array[T]: {
    each(callbackfn: (value: T, index?: Int, array?: Array[T]) -> Void) => "%array.each"
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Types) != 1 || len(file.Types[0].Methods) != 1 {
		t.Fatalf("parsed file = %#v, want one array method", file)
	}
	param := file.Types[0].Methods[0].Params[0]
	if param.Type.Canonical() != "Func[T,Int,Array[T],Void]" {
		t.Fatalf("param type = %q, want canonical Func type", param.Type.Canonical())
	}
	if param.Type.Kind != ast.TypeFunction || len(param.Type.Params) != 3 || param.Type.Params[1].Name != "index" || !param.Type.Params[1].Optional {
		t.Fatalf("param type node = %#v, want function type with named optional parameter", param.Type)
	}
	wantDisplay := "(value: T, index?: Int, array?: Array[T]) -> Void"
	if param.Type.Display() != wantDisplay {
		t.Fatalf("param type display = %q, want %q", param.Type.Display(), wantDisplay)
	}
}

func TestBlockStartingWithCallIsNotObjectLiteral(t *testing.T) {
	file, errs := Parse(`nestedMatch() => {
}

main() => {
    nestedMatch()
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Functions) != 2 {
		t.Fatalf("functions = %d, want 2", len(file.Functions))
	}
	block, ok := file.Functions[1].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 1 {
		t.Fatalf("main body = %#v, want one block statement", file.Functions[1].Body)
	}
	stmt, ok := block.Statements[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("statement = %T, want ExprStmt", block.Statements[0])
	}
	if _, ok := stmt.Expr.(*ast.CallExpr); !ok {
		t.Fatalf("expression = %T, want CallExpr", stmt.Expr)
	}
}

func TestObjectMethodDependsOnObjectContext(t *testing.T) {
	file, errs := Parse(`main() => {
    obj := {
        greet() => @io.println("hi")
    }
    obj.greet()
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 2 {
		t.Fatalf("main body = %#v, want two block statements", file.Functions[0].Body)
	}
	let, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("first statement = %T, want LetStmt", block.Statements[0])
	}
	obj, ok := let.Value.(*ast.AnonymousObjectLiteral)
	if !ok || len(obj.Fields) != 1 {
		t.Fatalf("let value = %#v, want object with one method", let.Value)
	}
	if _, ok := obj.Fields[0].Value.(*ast.LambdaExpr); !ok {
		t.Fatalf("object field = %T, want LambdaExpr method", obj.Fields[0].Value)
	}
}

func TestObjectDestructureDeclaration(t *testing.T) {
	file, errs := Parse(`main() => {
  { state: nextState, ch } := step
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	stmt, ok := block.Statements[0].(*ast.ObjectDestructureStmt)
	if !ok {
		t.Fatalf("statement = %T, want ObjectDestructureStmt", block.Statements[0])
	}
	if len(stmt.Fields) != 2 {
		t.Fatalf("fields = %#v, want 2 fields", stmt.Fields)
	}
	if stmt.Fields[0].Field != "state" || stmt.Fields[0].Name != "nextState" {
		t.Fatalf("first field = %#v, want state: nextState", stmt.Fields[0])
	}
	if stmt.Fields[1].Field != "ch" || stmt.Fields[1].Name != "ch" {
		t.Fatalf("second field = %#v, want ch", stmt.Fields[1])
	}
}

func TestAnonymousObjectMethodMembers(t *testing.T) {
	file, errs := Parse(`main() => {
    obj := {
        name: "Alice"
        nextAge() => .age + 1
        - greetText() => "Hello, " + .name
    }
}

`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Functions) != 1 {
		t.Fatalf("Parse() functions = %d, want 1", len(file.Functions))
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 1 {
		t.Fatalf("main body = %#v, want one block statement", file.Functions[0].Body)
	}
	let, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("statement = %T, want LetStmt", block.Statements[0])
	}
	obj, ok := let.Value.(*ast.AnonymousObjectLiteral)
	if !ok {
		t.Fatalf("let value = %T, want AnonymousObjectLiteral", let.Value)
	}
	if len(obj.Fields) != 3 {
		t.Fatalf("object fields = %d, want 3", len(obj.Fields))
	}
	method, ok := obj.Fields[1].Value.(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("method field = %T, want LambdaExpr", obj.Fields[1].Value)
	}
	if obj.Fields[1].Name != "nextAge" || len(method.Params) != 0 {
		t.Fatalf("method field = %s params=%v, want nextAge()", obj.Fields[1].Name, method.Params)
	}
	privateMethod, ok := obj.Fields[2].Value.(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("private method field = %T, want LambdaExpr", obj.Fields[2].Value)
	}
	if obj.Fields[2].Name != "greetText" || !obj.Fields[2].Private || len(privateMethod.Params) != 0 {
		t.Fatalf("private method field = %s private=%v params=%v, want private greetText()", obj.Fields[2].Name, obj.Fields[2].Private, privateMethod.Params)
	}
}

func TestParseMapLiteral(t *testing.T) {
	file, errs := Parse(`main() => {
    values := {
        "a": 1,
        "b": 2
    }
    values["b"] = 3
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 2 {
		t.Fatalf("main body = %#v, want two statements", file.Functions[0].Body)
	}
	let, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("first statement = %T, want LetStmt", block.Statements[0])
	}
	lit, ok := let.Value.(*ast.MapLiteral)
	if !ok || len(lit.Entries) != 2 {
		t.Fatalf("let value = %#v, want map with two entries", let.Value)
	}
	if _, ok := lit.Entries[0].Key.(*ast.StringLiteral); !ok {
		t.Fatalf("first key = %T, want StringLiteral", lit.Entries[0].Key)
	}
	stmt, ok := block.Statements[1].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("second statement = %T, want ExprStmt", block.Statements[1])
	}
	assign, ok := stmt.Expr.(*ast.AssignExpr)
	if !ok {
		t.Fatalf("second expr = %T, want AssignExpr", stmt.Expr)
	}
	if _, ok := assign.Target.(*ast.IndexExpr); !ok {
		t.Fatalf("assign target = %T, want IndexExpr", assign.Target)
	}
}

func TestMapLiteralWithNonStringKeys(t *testing.T) {
	file, errs := Parse(`main() => {
    values := {
        1: 2,
        2: 4
    }
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	let := block.Statements[0].(*ast.LetStmt)
	lit, ok := let.Value.(*ast.MapLiteral)
	if !ok || len(lit.Entries) != 2 {
		t.Fatalf("let value = %#v, want map with two entries", let.Value)
	}
	if _, ok := lit.Entries[0].Key.(*ast.IntegerLiteral); !ok {
		t.Fatalf("first key = %T, want IntegerLiteral", lit.Entries[0].Key)
	}
}

func TestAnonymousObjectLiteralStillUsesIdentifierKeys(t *testing.T) {
	file, errs := Parse(`main() => {
    obj := {
        name: "Alice"
    }
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	let := block.Statements[0].(*ast.LetStmt)
	if _, ok := let.Value.(*ast.AnonymousObjectLiteral); !ok {
		t.Fatalf("let value = %T, want AnonymousObjectLiteral", let.Value)
	}
}

func TestParseXMLElementWithEmbeddedExpressions(t *testing.T) {
	file, errs := Parse(`render() -> HTMLElement => {
    count $= 0

    <div>
        <h1>Counter Example</h1>
        <p>Count: {count}</p>
        <button @click={count++}>Click Me</button>
    </div>
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 2 {
		t.Fatalf("render body = %#v, want two statements", file.Functions[0].Body)
	}
	stmt, ok := block.Statements[1].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("second statement = %T, want ExprStmt", block.Statements[1])
	}
	root, ok := stmt.Expr.(*ast.XMLElement)
	if !ok {
		t.Fatalf("expression = %T, want XMLElement", stmt.Expr)
	}
	if root.Tag != "div" || len(root.Children) != 3 {
		t.Fatalf("root = <%s> children=%d, want div with 3 children", root.Tag, len(root.Children))
	}
	button, ok := root.Children[2].Expr.(*ast.XMLElement)
	if !ok || button.Tag != "button" || len(button.Attrs) != 1 {
		t.Fatalf("button child = %#v, want button with one event", root.Children[2].Expr)
	}
	if !button.Attrs[0].Event || button.Attrs[0].Name != "click" {
		t.Fatalf("button attr = %#v, want @click", button.Attrs[0])
	}
	if _, ok := button.Attrs[0].Value.(*ast.PostfixExpr); !ok {
		t.Fatalf("button attr value = %T, want PostfixExpr", button.Attrs[0].Value)
	}
}

func TestParseReactiveLiterals(t *testing.T) {
	file, errs := Parse(`main() => {
    list := $["Item 1"]
    obj := ${name: "Alice"}
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 2 {
		t.Fatalf("main body = %#v, want two statements", file.Functions[0].Body)
	}
	list, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("first statement = %T, want LetStmt", block.Statements[0])
	}
	if _, ok := list.Value.(*ast.ReactiveLiteral); !ok {
		t.Fatalf("list value = %T, want ReactiveLiteral", list.Value)
	}
	obj, ok := block.Statements[1].(*ast.LetStmt)
	if !ok {
		t.Fatalf("second statement = %T, want LetStmt", block.Statements[1])
	}
	if _, ok := obj.Value.(*ast.ReactiveLiteral); !ok {
		t.Fatalf("obj value = %T, want ReactiveLiteral", obj.Value)
	}
}

func TestParseSignalPrefixSyntax(t *testing.T) {
	file, errs := Parse(`main() => {
    $count := 0
    $double := $count * 2
    {
        @io.println($count)
        @io.println($double)
    }
    $count = $count + 1
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 4 {
		t.Fatalf("main body = %#v, want four statements", file.Functions[0].Body)
	}
	count, ok := block.Statements[0].(*ast.LetStmt)
	if !ok || !count.Signal || count.Name != "count" {
		t.Fatalf("first statement = %#v, want signal count binding", block.Statements[0])
	}
	double, ok := block.Statements[1].(*ast.LetStmt)
	if !ok || !double.Signal || double.Name != "double" {
		t.Fatalf("second statement = %#v, want signal double binding", block.Statements[1])
	}
	bin, ok := double.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("double value = %T, want BinaryExpr", double.Value)
	}
	ref, ok := bin.Left.(*ast.Identifier)
	if !ok || ref.Name != "count" || !ref.SignalPrefix {
		t.Fatalf("double dependency = %#v, want $count identifier", bin.Left)
	}
	effect, ok := block.Statements[2].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("third statement = %#v, want effect scope expression statement", block.Statements[2])
	}
	if _, ok := effect.Expr.(*ast.BlockExpr); !ok {
		t.Fatalf("third statement expr = %T, want BlockExpr", effect.Expr)
	}
	assign, ok := block.Statements[3].(*ast.AssignStmt)
	if !ok || assign.Name != "count" {
		t.Fatalf("fourth statement = %#v, want count assignment", block.Statements[3])
	}
}

func TestParseArraySpread(t *testing.T) {
	file, errs := Parse(`main() => {
    next := [...items, "New Item"]
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 1 {
		t.Fatalf("main body = %#v, want one statement", file.Functions[0].Body)
	}
	let, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("statement = %T, want LetStmt", block.Statements[0])
	}
	array, ok := let.Value.(*ast.ArrayLiteral)
	if !ok || len(array.Elements) != 2 {
		t.Fatalf("let value = %#v, want array with two elements", let.Value)
	}
	if _, ok := array.Elements[0].(*ast.SpreadExpr); !ok {
		t.Fatalf("first element = %T, want SpreadExpr", array.Elements[0])
	}
}

func TestParseAssignmentExpression(t *testing.T) {
	file, errs := Parse(`render() => {
    list $= ["Item 1"]
    <button @click={list = [...list, "New Item"]}>Add Item</button>
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 2 {
		t.Fatalf("render body = %#v, want two statements", file.Functions[0].Body)
	}
	stmt, ok := block.Statements[1].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("second statement = %T, want ExprStmt", block.Statements[1])
	}
	button, ok := stmt.Expr.(*ast.XMLElement)
	if !ok || len(button.Attrs) != 1 {
		t.Fatalf("expression = %#v, want button with one attr", stmt.Expr)
	}
	if _, ok := button.Attrs[0].Value.(*ast.AssignExpr); !ok {
		t.Fatalf("event value = %T, want AssignExpr", button.Attrs[0].Value)
	}
}
