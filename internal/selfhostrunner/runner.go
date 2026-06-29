package selfhostrunner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/ir"
)

type Result struct {
	Output string
	Err    error
}

var (
	driverOnce sync.Once
	driverPath string
	driverErr  error
)

func RunTestSource(source string, name string) Result {
	return runSelfhost("test", source, name)
}

func RunTestIR(file *ir.File, name string) Result {
	payload, err := marshalSelfhostIR(file)
	if err != nil {
		return Result{Err: err}
	}
	return runSelfhostPayload("ir-test", payload, name)
}

func RunMainSource(source string) Result {
	return runSelfhost("main", source, "")
}

func selfhostInterpreterPath() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot locate selfhost runner source")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "selfhost", "interpreter", "interpreter.rn"), nil
}

func runSelfhost(mode string, source string, name string) Result {
	return runSelfhostPayload(mode, []byte(source), name)
}

func runSelfhostPayload(mode string, payload []byte, name string) Result {
	dir, err := os.MkdirTemp("", "rune-selfhost-*")
	if err != nil {
		return Result{Err: err}
	}
	defer os.RemoveAll(dir)
	inputPath := filepath.Join(dir, "input")
	if err := os.WriteFile(inputPath, payload, 0o644); err != nil {
		return Result{Err: err}
	}
	driver, err := selfhostDriverPath()
	if err != nil {
		return Result{Err: err}
	}
	cmd, err := typeScriptRuntimeCommand(driver, mode, inputPath, name)
	if err != nil {
		return Result{Err: err}
	}
	out, err := cmd.CombinedOutput()
	output, runtimeErr := splitSelfhostError(string(out))
	if runtimeErr != nil {
		return Result{Output: output, Err: runtimeErr}
	}
	return Result{Output: output, Err: err}
}

func selfhostDriverPath() (string, error) {
	driverOnce.Do(func() {
		driverPath, driverErr = compileSelfhostDriver()
	})
	return driverPath, driverErr
}

func compileSelfhostDriver() (string, error) {
	interpreterPath, err := selfhostInterpreterPath()
	if err != nil {
		return "", err
	}
	goSource, diags := compiler.GenerateGoFile(interpreterPath)
	if len(diags) > 0 {
		return "", diagnosticsError(diags)
	}
	goSource = addGoDriverImports(goSource)
	goSource += selfhostDriverSource()
	dir, err := driverCacheDir()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(goSource))
	key := hex.EncodeToString(sum[:])[:16]
	binPath := filepath.Join(dir, "driver-"+key)
	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}
	goPath := filepath.Join(dir, "driver.go")
	if err := os.WriteFile(goPath, []byte(goSource), 0o644); err != nil {
		return "", err
	}
	tmpBin := filepath.Join(dir, fmt.Sprintf("driver-%s-%d.tmp", key, os.Getpid()))
	cmd := exec.Command("go", "build", "-o", tmpBin, goPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("build selfhost interpreter: %w\n%s", err, out)
	}
	if err := os.Rename(tmpBin, binPath); err != nil {
		return "", err
	}
	return binPath, nil
}

func driverCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "rune-lang", "selfhost-driver")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func selfhostDriverSource() string {
	return `
func main() {
	mode := "main"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	inputPath := ""
	if len(os.Args) > 2 {
		inputPath = os.Args[2]
	}
	name := ""
	if len(os.Args) > 3 {
		name = os.Args[3]
	}
	input, err := os.ReadFile(inputPath)
	if err != nil {
		println("__RUNE_SELFHOST_ERROR__" + err.Error())
		return
	}
	var result __InterpretResult
	if mode == "ir-test" {
		var fileJSON __runeSelfhostIRFile
		if err := json.Unmarshal(input, &fileJSON); err != nil {
			println("__RUNE_SELFHOST_ERROR__" + err.Error())
			return
		}
		result = __runTestIR(__runeSelfhostFile(fileJSON), name)
	} else if mode == "test" {
		result = __interpretTest(string(input), name)
	} else {
		result = __interpret(string(input))
	}
	for _, line := range result.__output {
		println(line)
	}
	if !result.__ok {
		println("__RUNE_SELFHOST_ERROR__" + result.__error)
	}
}

type __runeSelfhostIRFile struct {
	Imports   []__runeSelfhostIRImport  ` + "`json:\"imports\"`" + `
	Structs   []__runeSelfhostIRStruct  ` + "`json:\"structs\"`" + `
	Enums     []__runeSelfhostIREnum    ` + "`json:\"enums\"`" + `
	Functions []__runeSelfhostIRFunc    ` + "`json:\"functions\"`" + `
	Tests     []__runeSelfhostIRTest    ` + "`json:\"tests\"`" + `
	Errors    []__runeSelfhostParseErr  ` + "`json:\"errors\"`" + `
}

type __runeSelfhostIRImport struct {
	Path string ` + "`json:\"path\"`" + `
	Go bool ` + "`json:\"go\"`" + `
	Line int ` + "`json:\"line\"`" + `
	Column int ` + "`json:\"column\"`" + `
}

type __runeSelfhostIRParam struct {
	Name string ` + "`json:\"name\"`" + `
	TypeName string ` + "`json:\"typeName\"`" + `
	Line int ` + "`json:\"line\"`" + `
	Column int ` + "`json:\"column\"`" + `
}

type __runeSelfhostIRExpr struct {
	Kind int ` + "`json:\"kind\"`" + `
	Text string ` + "`json:\"text\"`" + `
	Name string ` + "`json:\"name\"`" + `
	Value string ` + "`json:\"value\"`" + `
	Op string ` + "`json:\"op\"`" + `
	Params []__runeSelfhostIRParam ` + "`json:\"params\"`" + `
	Children []__runeSelfhostIRExpr ` + "`json:\"children\"`" + `
	Line int ` + "`json:\"line\"`" + `
	Column int ` + "`json:\"column\"`" + `
}

type __runeSelfhostIRField struct {
	Name string ` + "`json:\"name\"`" + `
	Private bool ` + "`json:\"private\"`" + `
	TypeName string ` + "`json:\"typeName\"`" + `
	Line int ` + "`json:\"line\"`" + `
	Column int ` + "`json:\"column\"`" + `
}

type __runeSelfhostIREnumMember struct {
	Name string ` + "`json:\"name\"`" + `
	Private bool ` + "`json:\"private\"`" + `
	Value string ` + "`json:\"value\"`" + `
	Line int ` + "`json:\"line\"`" + `
	Column int ` + "`json:\"column\"`" + `
}

type __runeSelfhostIRFunc struct {
	Name string ` + "`json:\"name\"`" + `
	Private bool ` + "`json:\"private\"`" + `
	Routine bool ` + "`json:\"routine\"`" + `
	ReceiverType string ` + "`json:\"receiverType\"`" + `
	Generics []string ` + "`json:\"generics\"`" + `
	Params []__runeSelfhostIRParam ` + "`json:\"params\"`" + `
	ReturnType string ` + "`json:\"returnType\"`" + `
	Body __runeSelfhostIRExpr ` + "`json:\"body\"`" + `
	Line int ` + "`json:\"line\"`" + `
	Column int ` + "`json:\"column\"`" + `
}

type __runeSelfhostIRStruct struct {
	Name string ` + "`json:\"name\"`" + `
	Private bool ` + "`json:\"private\"`" + `
	Generics []string ` + "`json:\"generics\"`" + `
	Fields []__runeSelfhostIRField ` + "`json:\"fields\"`" + `
	Methods []__runeSelfhostIRFunc ` + "`json:\"methods\"`" + `
	Line int ` + "`json:\"line\"`" + `
	Column int ` + "`json:\"column\"`" + `
}

type __runeSelfhostIREnum struct {
	Name string ` + "`json:\"name\"`" + `
	Private bool ` + "`json:\"private\"`" + `
	Members []__runeSelfhostIREnumMember ` + "`json:\"members\"`" + `
	Line int ` + "`json:\"line\"`" + `
	Column int ` + "`json:\"column\"`" + `
}

type __runeSelfhostIRTest struct {
	Name string ` + "`json:\"name\"`" + `
	Body __runeSelfhostIRExpr ` + "`json:\"body\"`" + `
	Line int ` + "`json:\"line\"`" + `
	Column int ` + "`json:\"column\"`" + `
}

type __runeSelfhostParseErr struct {
	Message string ` + "`json:\"message\"`" + `
	Line int ` + "`json:\"line\"`" + `
	Column int ` + "`json:\"column\"`" + `
}

func __runeSelfhostFile(in __runeSelfhostIRFile) __IRFile {
	return __IRFile{
		__imports: __runeSelfhostImports(in.Imports),
		__structs: __runeSelfhostStructs(in.Structs),
		__enums: __runeSelfhostEnums(in.Enums),
		__functions: __runeSelfhostFuncs(in.Functions),
		__tests: __runeSelfhostTests(in.Tests),
		__errors: __runeSelfhostErrors(in.Errors),
	}
}

func __runeSelfhostImports(in []__runeSelfhostIRImport) []__IRImport {
	out := make([]__IRImport, 0, len(in))
	for _, item := range in {
		out = append(out, __IRImport{__path: item.Path, __go: item.Go, __line: item.Line, __column: item.Column})
	}
	return out
}

func __runeSelfhostParams(in []__runeSelfhostIRParam) []__IRParam {
	out := make([]__IRParam, 0, len(in))
	for _, item := range in {
		out = append(out, __IRParam{__name: item.Name, __typeName: item.TypeName, __line: item.Line, __column: item.Column})
	}
	return out
}

func __runeSelfhostExprs(in []__runeSelfhostIRExpr) []__IRExpr {
	out := make([]__IRExpr, 0, len(in))
	for _, item := range in {
		out = append(out, __runeSelfhostExpr(item))
	}
	return out
}

func __runeSelfhostExpr(in __runeSelfhostIRExpr) __IRExpr {
	return __IRExpr{
		__kind: __ExprKind(in.Kind),
		__text: in.Text,
		__name: in.Name,
		__value: in.Value,
		__op: in.Op,
		__params: __runeSelfhostParams(in.Params),
		__children: __runeSelfhostExprs(in.Children),
		__line: in.Line,
		__column: in.Column,
	}
}

func __runeSelfhostFields(in []__runeSelfhostIRField) []__IRField {
	out := make([]__IRField, 0, len(in))
	for _, item := range in {
		out = append(out, __IRField{__name: item.Name, __private: item.Private, __typeName: item.TypeName, __line: item.Line, __column: item.Column})
	}
	return out
}

func __runeSelfhostEnumMembers(in []__runeSelfhostIREnumMember) []__IREnumMember {
	out := make([]__IREnumMember, 0, len(in))
	for _, item := range in {
		out = append(out, __IREnumMember{__name: item.Name, __private: item.Private, __value: item.Value, __line: item.Line, __column: item.Column})
	}
	return out
}

func __runeSelfhostFunc(in __runeSelfhostIRFunc) __IRFunction {
	return __IRFunction{
		__name: in.Name,
		__private: in.Private,
		__routine: in.Routine,
		__receiverType: in.ReceiverType,
		__generics: in.Generics,
		__params: __runeSelfhostParams(in.Params),
		__returnType: in.ReturnType,
		__body: __runeSelfhostExpr(in.Body),
		__line: in.Line,
		__column: in.Column,
	}
}

func __runeSelfhostFuncs(in []__runeSelfhostIRFunc) []__IRFunction {
	out := make([]__IRFunction, 0, len(in))
	for _, item := range in {
		out = append(out, __runeSelfhostFunc(item))
	}
	return out
}

func __runeSelfhostStructs(in []__runeSelfhostIRStruct) []__IRStructType {
	out := make([]__IRStructType, 0, len(in))
	for _, item := range in {
		out = append(out, __IRStructType{__name: item.Name, __private: item.Private, __generics: item.Generics, __fields: __runeSelfhostFields(item.Fields), __methods: __runeSelfhostFuncs(item.Methods), __line: item.Line, __column: item.Column})
	}
	return out
}

func __runeSelfhostEnums(in []__runeSelfhostIREnum) []__IREnumType {
	out := make([]__IREnumType, 0, len(in))
	for _, item := range in {
		out = append(out, __IREnumType{__name: item.Name, __private: item.Private, __members: __runeSelfhostEnumMembers(item.Members), __line: item.Line, __column: item.Column})
	}
	return out
}

func __runeSelfhostTests(in []__runeSelfhostIRTest) []__IRTest {
	out := make([]__IRTest, 0, len(in))
	for _, item := range in {
		out = append(out, __IRTest{__name: item.Name, __body: __runeSelfhostExpr(item.Body), __line: item.Line, __column: item.Column})
	}
	return out
}

func __runeSelfhostErrors(in []__runeSelfhostParseErr) []__ParseError {
	out := make([]__ParseError, 0, len(in))
	for _, item := range in {
		out = append(out, __ParseError{__message: item.Message, __line: item.Line, __column: item.Column})
	}
	return out
}
`
}

func addGoDriverImports(source string) string {
	if strings.Contains(source, "import (\n") {
		source = strings.Replace(source, "import (\n", "import (\n\t\"encoding/json\"\n\t\"os\"\n", 1)
	} else {
		source = strings.Replace(source, "package main\n\n", "package main\n\nimport (\n\t\"encoding/json\"\n\t\"os\"\n)\n\n", 1)
	}
	return source
}

func marshalSelfhostIR(file *ir.File) ([]byte, error) {
	return json.Marshal(selfhostFile(file))
}

func splitSelfhostError(output string) (string, error) {
	const marker = "__RUNE_SELFHOST_ERROR__"
	idx := strings.Index(output, marker)
	if idx < 0 {
		return output, nil
	}
	before := output[:idx]
	message := strings.TrimSpace(output[idx+len(marker):])
	if message == "" {
		message = "selfhost interpreter failed"
	}
	return before, fmt.Errorf("%s", message)
}

func diagnosticsError(diags []compiler.Diagnostic) error {
	msg := ""
	for _, diag := range diags {
		if msg != "" {
			msg += "\n"
		}
		if diag.Path != "" && diag.Pos.Line > 0 {
			msg += fmt.Sprintf("%s:%d:%d: %s", diag.Path, diag.Pos.Line, diag.Pos.Column, diag.Message)
			continue
		}
		if diag.Path != "" {
			msg += fmt.Sprintf("%s: %s", diag.Path, diag.Message)
			continue
		}
		msg += diag.Message
	}
	return fmt.Errorf("%s", msg)
}

func typeScriptRuntimeCommand(path string, args ...string) (*exec.Cmd, error) {
	return exec.Command(path, args...), nil
}
