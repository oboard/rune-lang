package compiler

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/interpreter"
	"github.com/oboard/rune-lang/internal/ir"
)

func TestAnalyzeCoreSourceDoesNotDuplicateOwnTypes(t *testing.T) {
	_, diags := AnalyzeFile(filepath.Join("..", "..", "core", "set", "set.rn"))
	for _, diag := range diags {
		if strings.Contains(diag.Message, `duplicate type "Set"`) ||
			strings.Contains(diag.Message, `duplicate type "WeakSet"`) {
			t.Fatalf("AnalyzeFile() diagnostics include own stdlib type duplicate: %#v", diags)
		}
	}
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
}

func TestAnalyzeCoreIterSource(t *testing.T) {
	_, diags := AnalyzeFile(filepath.Join("..", "..", "core", "iter", "iter.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
}

func TestAnalyzeCoreCompressSource(t *testing.T) {
	_, diags := AnalyzeFile(filepath.Join("..", "..", "core", "compress", "compress.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
}

func TestAnalyzeCoreCliSource(t *testing.T) {
	_, diags := AnalyzeFile(filepath.Join("..", "..", "core", "cli", "cli.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
}

func TestAnalyzeCoreSources(t *testing.T) {
	root := filepath.Join("..", "..", "core")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".rn" {
			return nil
		}
		_, diags := AnalyzeFile(path)
		if len(diags) > 0 {
			t.Fatalf("AnalyzeFile(%s) diagnostics = %#v", path, diags)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk core sources: %v", err)
	}
}

func TestAnalyzeSourceRequiresSyntaxImport(t *testing.T) {
	_, diags := AnalyzeSource("syntax_import.rn", `makeExpr(name: String) -> SyntaxExpr => IdentifierExpr(name)
`)
	for _, diag := range diags {
		if strings.Contains(diag.Message, `unknown return type "SyntaxExpr"`) {
			return
		}
	}
	t.Fatalf("AnalyzeSource() diagnostics = %#v, want missing SyntaxExpr import", diags)
}

func TestAnalyzeFileLoadsRuneImports(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "math.rn"), `+ inc(value: Int) -> Int => value + 1
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"math.rn"

main() => @io.println(inc(41))
`)

	prog, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
	if prog.Info.Functions["inc"] == nil {
		t.Fatalf("imported function inc was not collected")
	}

	var out bytes.Buffer
	runner := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := runner.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "42" {
		t.Fatalf("output = %q, want 42", got)
	}
}

func TestAnalyzeFileLoadsRuneImportNamespaces(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "helper.rn"), `+ greeting(name: String) -> String => privateGreeting(name)

privateGreeting(name: String) -> String => "hello, " + name
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `main() => {
  io := @io
  helper := @"helper.rn"
  io.println(helper.greeting("Alice"))
}
`)

	prog, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}

	var out bytes.Buffer
	runner := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := runner.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "hello, Alice" {
		t.Fatalf("output = %q, want hello, Alice", got)
	}
}

func TestGenerateGoFileImportsRuneNamespaceReferences(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "helper.rn"), `+ greeting(name: String) -> String => "hello, " + name
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `main() => {
  io := @io
  helper := @"helper.rn"
  io.println(helper.greeting("Alice"))
}
`)

	got, diags := GenerateGoFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("GenerateGoFile() diagnostics = %#v", diags)
	}
	if !strings.Contains(got, `fmt.Println(__greeting("Alice"))`) {
		t.Fatalf("generated Go does not call imported namespace function:\n%s", got)
	}
	if strings.Contains(got, `__io := struct{}{}`) || strings.Contains(got, `__helper := struct{}{}`) {
		t.Fatalf("generated Go should not emit namespace placeholders:\n%s", got)
	}
}

func TestAnalyzeFileLoadsTypeScriptImports(t *testing.T) {
	dir := t.TempDir()
	writeTextFile(t, filepath.Join(dir, "greet.ts"), "export function greet(name: string): string {\n  return `Hello, ${name}!`;\n}\n\nexport function score(value: number): number { return value }\n\nexport function flag(value: boolean): boolean { return value }\n\nexport function size(value: bigint): bigint { return value }\n\nexport function noop(): void {}\n\nexport function mystery(value: any): unknown { return value }\n\nexport const version: string = \"1.0.0\"\n")
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"greet.ts"

main() => {
  @io.println(greet("Rune"))
  @io.println(version)
}
`)

	prog, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
	fn := prog.Info.Functions["greet"]
	if fn == nil || !fn.External {
		t.Fatalf("greet function = %#v, want external TypeScript function", fn)
	}
	if len(fn.Params) != 1 || fn.Params[0].Name != "name" || fn.Params[0].Type != checker.String {
		t.Fatalf("greet params = %#v, want name: String", fn.Params)
	}
	if fn.Return != checker.String {
		t.Fatalf("greet return = %s, want String", fn.Return)
	}
	score := prog.Info.Functions["score"]
	if score == nil || len(score.Params) != 1 || score.Params[0].Type != checker.Double || score.Return != checker.Double {
		t.Fatalf("score function = %#v, want (Double) -> Double", score)
	}
	flag := prog.Info.Functions["flag"]
	if flag == nil || len(flag.Params) != 1 || flag.Params[0].Type != checker.Bool || flag.Return != checker.Bool {
		t.Fatalf("flag function = %#v, want (Bool) -> Bool", flag)
	}
	size := prog.Info.Functions["size"]
	if size == nil || len(size.Params) != 1 || size.Params[0].Type != checker.BigInt || size.Return != checker.BigInt {
		t.Fatalf("size function = %#v, want (BigInt) -> BigInt", size)
	}
	noop := prog.Info.Functions["noop"]
	if noop == nil || noop.Return != checker.Void {
		t.Fatalf("noop function = %#v, want Void return", noop)
	}
	mystery := prog.Info.Functions["mystery"]
	if mystery == nil || len(mystery.Params) != 1 || mystery.Params[0].Type != checker.Unknown || mystery.Return != checker.Unknown {
		t.Fatalf("mystery function = %#v, want Unknown types", mystery)
	}
	if len(prog.Info.ExternalValues) != 1 {
		t.Fatalf("external values = %#v, want version", prog.Info.ExternalValues)
	}
	version := prog.Info.ExternalValues[0]
	if version.Name != "version" || version.Type != checker.String {
		t.Fatalf("external value = %#v, want version: String", version)
	}
}

func TestGenerateTypeScriptFileImportsTypeScriptFunctions(t *testing.T) {
	dir := t.TempDir()
	tsPath := filepath.Join(dir, "greet.ts")
	writeTextFile(t, tsPath, "export function greet(name: string): string {\n  return `Hello, ${name}!`;\n}\n\nexport const version: string = \"1.0.0\"\n")
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"greet.ts"

main() => {
  @io.println(greet("Rune"))
  @io.println(version)
}
`)

	got, diags := GenerateTypeScriptFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("GenerateTypeScriptFile() diagnostics = %#v", diags)
	}
	wantImport := `import { greet as __greet, version as __version } from "greet.ts";`
	if !strings.Contains(got, wantImport) {
		t.Fatalf("generated TypeScript missing %q:\n%s", wantImport, got)
	}
	if !strings.Contains(got, `console.log(__greet("Rune"));`) {
		t.Fatalf("generated TypeScript does not call imported alias:\n%s", got)
	}
	if !strings.Contains(got, `console.log(__version);`) {
		t.Fatalf("generated TypeScript does not read imported value alias:\n%s", got)
	}
	if strings.Contains(got, "function __greet") {
		t.Fatalf("generated TypeScript should not emit imported function body:\n%s", got)
	}
}

func TestGenerateTypeScriptFileImportsNamespaceReferences(t *testing.T) {
	dir := t.TempDir()
	writeTextFile(t, filepath.Join(dir, "greet.ts"), "export function greet(name: string): string {\n  return `Hello, ${name}!`;\n}\n\nexport const version: string = \"1.0.0\"\n")
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `main() => {
  io := @io
  helper := @"greet.ts"
  io.println(helper.greet("Rune"))
  io.println(helper.version)
}
`)

	got, diags := GenerateTypeScriptFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("GenerateTypeScriptFile() diagnostics = %#v", diags)
	}
	wantImport := `import { greet as __greet, version as __version } from "greet.ts";`
	if !strings.Contains(got, wantImport) {
		t.Fatalf("generated TypeScript missing %q:\n%s", wantImport, got)
	}
	if !strings.Contains(got, `console.log(__greet("Rune"));`) {
		t.Fatalf("generated TypeScript does not call imported namespace alias:\n%s", got)
	}
	if !strings.Contains(got, `console.log(__version);`) {
		t.Fatalf("generated TypeScript does not read imported namespace alias:\n%s", got)
	}
}

func TestGenerateGoFileRejectsTypeScriptImports(t *testing.T) {
	dir := t.TempDir()
	writeTextFile(t, filepath.Join(dir, "greet.ts"), "export function greet(name: string): string { return name }\n")
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"greet.ts"

main() => @io.println(greet("Rune"))
`)

	_, diags := GenerateGoFile(filepath.Join(dir, "main.rn"))
	for _, diag := range diags {
		if strings.Contains(diag.Message, "Go backend does not support TypeScript imports") {
			return
		}
	}
	t.Fatalf("GenerateGoFile() diagnostics = %#v, want TypeScript import backend diagnostic", diags)
}

func TestGenerateMoonBitFile(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `main() => {
  @io.println("Rune")
}
`)

	got, diags := GenerateMoonBitFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("GenerateMoonBitFile() diagnostics = %#v", diags)
	}
	if !strings.Contains(got, "fn main {\n") || !strings.Contains(got, `println("Rune")`) {
		t.Fatalf("generated MoonBit =\n%s", got)
	}
}

func TestGenerateMoonBitFileRejectsTypeScriptImports(t *testing.T) {
	dir := t.TempDir()
	writeTextFile(t, filepath.Join(dir, "greet.ts"), "export function greet(name: string): string { return name }\n")
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"greet.ts"

main() => @io.println(greet("Rune"))
`)

	_, diags := GenerateMoonBitFile(filepath.Join(dir, "main.rn"))
	for _, diag := range diags {
		if strings.Contains(diag.Message, "MoonBit backend does not support TypeScript imports") {
			return
		}
	}
	t.Fatalf("GenerateMoonBitFile() diagnostics = %#v, want TypeScript import backend diagnostic", diags)
}

func TestAnalyzeSourceSupportsUnicodeIdentifiers(t *testing.T) {
	src := `问候: {
  名字💡: String

  欢迎👋() -> String => .名字💡 + "!"
}

计算✅(数值🐉: Int) -> Int => {
  增量📈 := 1
  数值🐉 + 增量📈
}

main() => {
  用户👩‍💻 := 问候 { 名字💡: "Rune" }
  @io.println(用户👩‍💻.欢迎👋())
  @io.println(计算✅(41))
}
`
	prog, diags := AnalyzeSource("unicode_identifiers.rn", src)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}

	var out bytes.Buffer
	runner := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := runner.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "Rune!\n42" {
		t.Fatalf("output = %q, want Rune!\\n42", got)
	}
}

func TestAnalyzeFileLoadsTransitiveRuneImports(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "base.rn"), `+ base() -> Int => 40
`)
	writeRuneFile(t, filepath.Join(dir, "math.rn"), `@"base.rn"

+ inc2(value: Int) -> Int => value + 2
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"math.rn"

main() => @io.println(inc2(base()))
`)

	prog, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
	if prog.Info.Functions["base"] == nil || prog.Info.Functions["inc2"] == nil {
		t.Fatalf("transitive imports were not collected: %#v", prog.Info.Functions)
	}
}

func TestAnalyzeFileKeepsPrivateImportsAvailableToOwner(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "math.rn"), `add(a: Int, b: Int) -> Int => a + b

+ sum(a: Int, b: Int) -> Int => add(a, b)
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"math.rn"

main() => @io.println(sum(20, 22))
`)

	prog, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
	if prog.Info.Functions["add"] == nil || !prog.Info.Functions["add"].Private {
		t.Fatalf("private function add was not collected: %#v", prog.Info.Functions["add"])
	}

	var out bytes.Buffer
	runner := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := runner.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "42" {
		t.Fatalf("output = %q, want 42", got)
	}
}

func TestAnalyzeFileRejectsPrivateImportedFunctionUse(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "math.rn"), `add(a: Int, b: Int) -> Int => a + b
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"math.rn"

main() => add(20, 22)
`)

	_, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	for _, diag := range diags {
		if strings.Contains(diag.Message, `function "add" is private`) {
			return
		}
	}
	t.Fatalf("AnalyzeFile() diagnostics = %#v, want private function diagnostic", diags)
}

func TestAnalyzeFileAllowsLocalFunctionToShadowImportedPrivate(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "helper.rn"), `private() => "this is private function"
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"helper.rn"

private() => "this is real function"

main() => @io.println(private())
`)

	prog, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
	var out bytes.Buffer
	runner := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := runner.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "this is real function" {
		t.Fatalf("output = %q, want local function", got)
	}
}

func TestAnalyzeFileRejectsPrivateImportedMembers(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "user.rn"), `+ User: {
  name: String
  - token: String

  - secret() -> String => .token
}

+ Status: {
  + Ready = 0
  Hidden = 1
}
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"user.rn"

main() => {
  user := User { name: "Rune", token: "secret" }
  @io.println(user.token)
  @io.println(user.secret())
  Status.Hidden
}
`)

	_, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	want := []string{
		`field "User.token" is private`,
		`method "User.secret" is private`,
	}
	for _, msg := range want {
		found := false
		for _, diag := range diags {
			if strings.Contains(diag.Message, msg) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("AnalyzeFile() diagnostics = %#v, want %s", diags, msg)
		}
	}
}

func TestAnalyzeFileRequiresRuneImportExtension(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "math.rn"), `inc(value: Int) -> Int => value + 1
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"math"

main() => inc(1)
`)

	_, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	for _, diag := range diags {
		if strings.Contains(diag.Message, "must include a file extension") {
			return
		}
	}
	t.Fatalf("AnalyzeFile() diagnostics = %#v, want missing extension diagnostic", diags)
}

func TestAnalyzeSourceExpandsRuneMacroBeforeLowering(t *testing.T) {
	prog, diags := AnalyzeSource("macro_expand.rn", `#macro.renameDeclaration("FinalArgs")
Args: {
  verbose: Bool
}
`)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	if len(prog.File.Types) != 1 || prog.File.Types[0].Name != "FinalArgs" {
		t.Fatalf("expanded AST types = %#v", prog.File.Types)
	}
	if len(prog.IR.Types) != 1 || prog.IR.Types[0].Name != "FinalArgs" {
		t.Fatalf("expanded IR types = %#v", prog.IR.Types)
	}
	if len(prog.Macros) != 1 || prog.Macros[0].Annotation.Name != "renameDeclaration" {
		t.Fatalf("macro plan = %#v", prog.Macros)
	}
}

func TestAnalyzeSourceEvaluatesCompileTimeExpression(t *testing.T) {
	prog, diags := AnalyzeSource("consteval.rn", `double(value: Int) -> Int => value * 2

main() => (double(21))'
`)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	lit, ok := prog.File.Functions[1].Body.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("main body = %T, want IntegerLiteral", prog.File.Functions[1].Body)
	}
	if lit.Value != 42 {
		t.Fatalf("literal value = %d, want 42", lit.Value)
	}
	if len(prog.IR.Functions) != 1 || prog.IR.Functions[0].Name != "main" {
		t.Fatalf("IR functions = %#v, want only main", prog.IR.Functions)
	}
	if got := prog.IR.Functions[0].Body.(*ir.IntegerLiteral).Value; got != 42 {
		t.Fatalf("IR literal value = %d, want 42", got)
	}
}

func TestAnalyzeSourceEvaluatesCompileTimeProcessPlatform(t *testing.T) {
	prog, diags := AnalyzeSource("consteval_process.rn", `main() => (@process.platform())'
`)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	lit, ok := prog.File.Functions[0].Body.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("main body = %T, want StringLiteral", prog.File.Functions[0].Body)
	}
	if lit.Value == "" {
		t.Fatalf("platform literal should not be empty")
	}
}

func TestAnalyzeSourceKeepsRuntimeUseOfCompileTimeFunction(t *testing.T) {
	prog, diags := AnalyzeSource("consteval.rn", `double(value: Int) -> Int => value * 2

main() => double(1) + (double(21))'
`)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	names := []string{}
	for _, fn := range prog.IR.Functions {
		names = append(names, fn.Name)
	}
	if !strings.Contains(strings.Join(names, ","), "double") {
		t.Fatalf("IR functions = %#v, want double retained", names)
	}
}

func TestAnalyzeSourceExpandsJSONFromJsonMethod(t *testing.T) {
	src := `#json.object
Document: {
  name: String
}

main() => {
  document := @json.parse("{\"name\":\"Rune\"}") : Document
  document
}
`
	prog, diags := AnalyzeSource("main.rn", src)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %v", diags)
	}
	if len(prog.File.Types) != 1 || len(prog.File.Types[0].Methods) != 1 {
		t.Fatalf("expanded type methods = %#v", prog.File.Types)
	}
	method := prog.File.Types[0].Methods[0]
	if method.Name != "fromJson" || !method.Static || method.ReturnType.Canonical() != "Document" {
		t.Fatalf("generated method = %#v", method)
	}
}

func TestAnalyzeSourceExpandsMultipleJSONFromJsonMethods(t *testing.T) {
	src := `#json.object
FirstDocument: {
  #json.name("first_name")
  name: String
}

#json.object
SecondDocument: {
  #json.ignore
  secret: String
}

main() => null
`
	prog, diags := AnalyzeSource("main.rn", src)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %v", diags)
	}
	for _, typ := range prog.File.Types {
		if len(typ.Methods) != 1 || typ.Methods[0].Name != "fromJson" || !typ.Methods[0].Static {
			t.Fatalf("expanded %s methods = %#v", typ.Name, typ.Methods)
		}
	}
}

func TestAnalyzeJSONFixtureExpandsFromJsonMethods(t *testing.T) {
	prog, diags := AnalyzeFile(filepath.Join("..", "..", "tests", "json.rn"))
	if len(diags) > 0 {
		methods := map[string][]string{}
		for _, typ := range prog.File.Types {
			for _, method := range typ.Methods {
				methods[typ.Name] = append(methods[typ.Name], method.Name)
			}
		}
		t.Fatalf("AnalyzeFile() diagnostics = %v, methods = %v", diags, methods)
	}
}

func TestAnalyzeSourceReportsCompileTimeSideEffects(t *testing.T) {
	_, diags := AnalyzeSource("macro_io.rn", `@syntax

#bad(tree: SyntaxFile, context: MacroContext) -> SyntaxFile => {
  @io.println("side effect")
  tree
}

#bad
Args: {
  verbose: Bool
}
	`)
	for _, diag := range diags {
		if strings.Contains(diag.Message, "not pure") && strings.Contains(diag.Message, "@io.println") {
			return
		}
	}
	t.Fatalf("AnalyzeSource() diagnostics = %#v, want compile-time side-effect error", diags)
}

func TestAnalyzeSourceLowersReturnedSyntaxTree(t *testing.T) {
	prog, diags := AnalyzeSource("syntax_macro.rn", `@syntax

#renameFirst(
  tree: SyntaxFile,
  context: MacroContext,
  name: String
) -> SyntaxFile => {
  current := tree.types[0]
  selectedName := context.targetID == current.id ? name : current.name
	  renamed := SyntaxStruct {
	    id: current.id,
	    name: selectedName,
	    private: current.private,
	    generics: current.generics,
	    annotations: current.annotations,
	    fields: current.fields,
	    methods: current.methods,
	    sourcePath: current.sourcePath
	  }
	  SyntaxFile {
	    types: [renamed],
	    enums: tree.enums,
	    functions: tree.functions
	  }
}

#renameFirst("Generated")
Original: {
  value: Int
}
`)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	if len(prog.File.Types) != 1 || prog.File.Types[0].Name != "Generated" {
		t.Fatalf("expanded AST types = %#v", prog.File.Types)
	}
	if len(prog.IR.Types) != 1 || prog.IR.Types[0].Name != "Generated" {
		t.Fatalf("expanded IR types = %#v", prog.IR.Types)
	}
}

func TestAnalyzeSourceExpandsStructLiteralSpreadInMacro(t *testing.T) {
	prog, diags := AnalyzeSource("syntax_macro_spread.rn", `@syntax

#renameFirst(
  tree: SyntaxFile,
  context: MacroContext,
  name: String
) -> SyntaxFile => {
  current := tree.types[0]
	  renamed := SyntaxStruct {
	    ...current,
	    name: context.targetID == current.id ? name : current.name
	  }
	  SyntaxFile {
	    types: [renamed],
	    enums: tree.enums,
	    functions: tree.functions
	  }
}

#renameFirst("Generated")
Original: {
  value: Int
}
`)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	if len(prog.File.Types) != 1 || prog.File.Types[0].Name != "Generated" {
		t.Fatalf("expanded AST types = %#v", prog.File.Types)
	}
	if len(prog.File.Types[0].Fields) != 1 || prog.File.Types[0].Fields[0].Name != "value" {
		t.Fatalf("expanded AST fields = %#v", prog.File.Types[0].Fields)
	}
}

func TestAnalyzeSourceGeneratesTypedCliEntry(t *testing.T) {
	prog, diags := AnalyzeSource("cli_macro.rn", `#cli.command("ship", "Ship an artifact", "1.0.0")
Args: {
  #cli.flag("v", "enable verbose output")
  verbose: Bool

  #cli.option("o", "FILE", "write output", "dist/app")
  output: String

  #cli.arg("target name")
  target: String
}

#cli.main
main(args: Args) => args.target
`)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	if len(prog.File.Functions) != 2 {
		t.Fatalf("expanded functions = %#v, want handler and entry", prog.File.Functions)
	}
	handler := prog.File.Functions[0]
	if handler.Name != "__cliMain" || len(handler.Params) != 1 || handler.Params[0].Type.Name != "Args" {
		t.Fatalf("handler = %#v, want __cliMain(args: Args)", handler)
	}
	entry := prog.File.Functions[1]
	if entry.Name != "main" || len(entry.Params) != 0 {
		t.Fatalf("entry = %#v, want main()", entry)
	}
	if _, ok := entry.Body.(*ast.BlockExpr); !ok {
		t.Fatalf("entry body = %T, want generated block", entry.Body)
	}
}

func writeRuneFile(t *testing.T, path string, src string) {
	t.Helper()
	writeTextFile(t, path, src)
}

func writeTextFile(t *testing.T, path string, src string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
