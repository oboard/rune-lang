package selfhostrunner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

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

func RunMainIR(file *ir.File) Result {
	payload, err := marshalSelfhostIR(file)
	if err != nil {
		return Result{Err: err}
	}
	return runSelfhostPayload("ir-main", payload, "")
}

func RunMainSource(source string) Result {
	return runSelfhost("main", source, "")
}

func selfhostInterpreterPath() (string, error) {
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "selfhost", "interpreter", "interpreter.rn"), nil
}

func repoRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot locate selfhost runner source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..")), nil
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
	dir, err := driverCacheDir()
	if err != nil {
		return "", err
	}
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	key, err := selfhostDriverCacheKey(root)
	if err != nil {
		return "", err
	}
	binPath := filepath.Join(dir, "driver-"+key)
	if _, err := os.Stat(binPath); err == nil {
		cleanupSelfhostDriverCache(dir, binPath)
		return binPath, nil
	}
	goCacheDir, cleanupGoCache, err := selfhostGoBuildCache(dir)
	if err != nil {
		return "", err
	}
	defer cleanupGoCache()
	buildEnv := append(os.Environ(), "GOMAXPROCS=1", "GOCACHE="+goCacheDir)
	interpreterPath, err := selfhostInterpreterPath()
	if err != nil {
		return "", err
	}
	goSource, err := generateSelfhostGoSource(interpreterPath, buildEnv)
	if err != nil {
		return "", err
	}
	goSource = addGoDriverImports(goSource)
	goSource += selfhostDriverSource()
	goFile, err := os.CreateTemp(dir, "driver-*.go")
	if err != nil {
		return "", err
	}
	goPath := goFile.Name()
	tmpBin := ""
	defer func() {
		_ = os.Remove(goPath)
		if tmpBin != "" {
			_ = os.Remove(tmpBin)
		}
	}()
	if _, err := goFile.Write([]byte(goSource)); err != nil {
		_ = goFile.Close()
		return "", err
	}
	if err := goFile.Close(); err != nil {
		return "", err
	}
	tmpFile, err := os.CreateTemp(dir, fmt.Sprintf("driver-%s-*.tmp", key))
	if err != nil {
		return "", err
	}
	tmpBin = tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return "", err
	}
	cmd := exec.Command("go", "build", "-p=1", "-gcflags=all=-l", "-o", tmpBin, goPath)
	cmd.Env = buildEnv
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("build selfhost interpreter: %w\n%s", err, out)
	}
	if err := os.Rename(tmpBin, binPath); err != nil {
		return "", err
	}
	cleanupSelfhostDriverCache(dir, binPath)
	return binPath, nil
}

func selfhostDriverCacheKey(root string) (string, error) {
	h := sha256.New()
	h.Write([]byte("rune-selfhost-driver-v2\n"))
	dirs := []string{
		"cmd/rune",
		"core",
		"internal",
		"selfhost",
	}
	for _, file := range []string{"go.mod", "go.sum"} {
		if err := hashCacheFile(h, root, file); err != nil {
			return "", err
		}
	}
	for _, dir := range dirs {
		if err := hashCacheDir(h, root, dir); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil))[:16], nil
}

func hashCacheDir(h hashWriter, root string, dir string) error {
	base := filepath.Join(root, dir)
	var files []string
	if err := filepath.WalkDir(base, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".rn" && ext != ".mod" && ext != ".sum" {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(files)
	for _, file := range files {
		if err := hashCacheFile(h, root, file); err != nil {
			return err
		}
	}
	return nil
}

type hashWriter interface {
	Write([]byte) (int, error)
}

func hashCacheFile(h hashWriter, root string, rel string) error {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return err
	}
	h.Write([]byte(rel))
	h.Write([]byte{0})
	h.Write(data)
	h.Write([]byte{0})
	return nil
}

func cleanupSelfhostDriverCache(dir string, keepPath string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	keepPath = filepath.Clean(keepPath)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "driver.go" || !strings.HasPrefix(name, "driver-") || strings.HasSuffix(name, ".tmp") || strings.HasSuffix(name, ".go") {
			continue
		}
		path := filepath.Join(dir, name)
		if filepath.Clean(path) == keepPath {
			continue
		}
		_ = os.Remove(path)
	}
}

func selfhostGoBuildCache(dir string) (string, func(), error) {
	goCacheDir, err := os.MkdirTemp(dir, "go-build-cache-*")
	if err != nil {
		return "", func() {}, err
	}
	return goCacheDir, func() {
		_ = os.RemoveAll(goCacheDir)
	}, nil
}

func generateSelfhostGoSource(interpreterPath string, env []string) (string, error) {
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("go", "run", "./cmd/rune", "go", interpreterPath)
	cmd.Dir = root
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("generate selfhost interpreter Go: %w\n%s", err, out)
	}
	return string(out), nil
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
	if mode == "ir-test" || mode == "ir-main" {
		var fileJSON __runeSelfhostIRFile
		if err := json.Unmarshal(input, &fileJSON); err != nil {
			println("__RUNE_SELFHOST_ERROR__" + err.Error())
			return
		}
		if mode == "ir-test" {
			result = __runTestIR(__runeSelfhostFile(fileJSON), name)
		} else {
			result = __runMainIR(__runeSelfhostFile(fileJSON))
		}
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
	Constants []__runeSelfhostIRConst   ` + "`json:\"constants\"`" + `
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
	JSONName string ` + "`json:\"jsonName\"`" + `
	JSONIgnore bool ` + "`json:\"jsonIgnore\"`" + `
	Line int ` + "`json:\"line\"`" + `
	Column int ` + "`json:\"column\"`" + `
}

type __runeSelfhostIREnumMember struct {
	Name string ` + "`json:\"name\"`" + `
	Private bool ` + "`json:\"private\"`" + `
	Value string ` + "`json:\"value\"`" + `
	Params []__runeSelfhostIRParam ` + "`json:\"params\"`" + `
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

type __runeSelfhostIRConst struct {
	Name string ` + "`json:\"name\"`" + `
	Private bool ` + "`json:\"private\"`" + `
	TypeName string ` + "`json:\"typeName\"`" + `
	Value __runeSelfhostIRExpr ` + "`json:\"value\"`" + `
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
	Generics []string ` + "`json:\"generics\"`" + `
	Members []__runeSelfhostIREnumMember ` + "`json:\"members\"`" + `
	Methods []__runeSelfhostIRFunc ` + "`json:\"methods\"`" + `
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
		__constants: __runeSelfhostConsts(in.Constants),
		__functions: __runeSelfhostFuncs(in.Functions),
		__tests: __runeSelfhostTests(in.Tests),
		__errors: __runeSelfhostErrors(in.Errors),
	}
}

func __runeSelfhostConsts(in []__runeSelfhostIRConst) []__IRConst {
	out := make([]__IRConst, 0, len(in))
	for _, item := range in {
		out = append(out, __IRConst{__name: item.Name, __private: item.Private, __typeName: item.TypeName, __value: __runeSelfhostExpr(item.Value), __line: item.Line, __column: item.Column})
	}
	return out
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
		out = append(out, __IRField{__name: item.Name, __private: item.Private, __typeName: item.TypeName, __jsonName: item.JSONName, __jsonIgnore: item.JSONIgnore, __line: item.Line, __column: item.Column})
	}
	return out
}

func __runeSelfhostEnumMembers(in []__runeSelfhostIREnumMember) []__IREnumMember {
	out := make([]__IREnumMember, 0, len(in))
	for _, item := range in {
		out = append(out, __IREnumMember{__name: item.Name, __private: item.Private, __value: item.Value, __params: __runeSelfhostParams(item.Params), __line: item.Line, __column: item.Column})
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
		out = append(out, __IREnumType{__name: item.Name, __private: item.Private, __generics: item.Generics, __members: __runeSelfhostEnumMembers(item.Members), __methods: __runeSelfhostFuncs(item.Methods), __line: item.Line, __column: item.Column})
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

func typeScriptRuntimeCommand(path string, args ...string) (*exec.Cmd, error) {
	return exec.Command(path, args...), nil
}
