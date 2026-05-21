package stdlib

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/parser"
)

type Registry struct {
	Modules map[string]*Module
}

type Module struct {
	Name      string
	Functions []Function

	byName map[string]*Function
}

type Function struct {
	Name         string
	Params       []string
	Return       string
	Variadic     bool
	TopLevelOnly bool
	Intrinsic    string
	Go           *GoBinding
}

type GoBinding struct {
	Import string
	Symbol string
}

func LoadDefault() (*Registry, error) {
	root, err := findCoreRoot()
	if err != nil {
		return nil, err
	}
	return Load(root)
}

func Load(root string) (*Registry, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("load core: %w", err)
	}

	reg := &Registry{Modules: map[string]*Module{}}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		path := filepath.Join(root, name, name+".rn")
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		mod, err := parseModule(name, path, string(data))
		if err != nil {
			return nil, err
		}
		reg.Modules[mod.Name] = mod
	}
	return reg, nil
}

func parseModule(name string, path string, src string) (*Module, error) {
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		var messages []string
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
		return nil, fmt.Errorf("%s: %s", path, strings.Join(messages, "; "))
	}
	if len(file.GoImports) > 0 || len(file.Types) > 0 {
		return nil, fmt.Errorf("%s: core stubs may only declare functions", path)
	}

	mod := &Module{Name: name, byName: map[string]*Function{}}
	seen := map[string]bool{}
	for _, decl := range file.Functions {
		fn, err := parseFunction(name, decl)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if seen[fn.Name] {
			return nil, fmt.Errorf("%s: duplicate function %s.%s", path, name, fn.Name)
		}
		seen[fn.Name] = true
		mod.Functions = append(mod.Functions, fn)
	}
	for i := range mod.Functions {
		mod.byName[mod.Functions[i].Name] = &mod.Functions[i]
	}
	return mod, nil
}

func parseFunction(moduleName string, decl *ast.Function) (Function, error) {
	body, ok := decl.Body.(*ast.StringLiteral)
	if !ok {
		return Function{}, fmt.Errorf("%s.%s body must be an intrinsic string", moduleName, decl.Name)
	}

	fn := Function{Name: decl.Name, Return: decl.ReturnType}
	for _, param := range decl.Params {
		fn.Params = append(fn.Params, param.Type)
	}

	spec := body.Value
	if !strings.HasPrefix(spec, "%") {
		return Function{}, fmt.Errorf("%s.%s intrinsic must start with %%", moduleName, decl.Name)
	}
	if err := applyIntrinsicSpec(moduleName, &fn, strings.TrimPrefix(spec, "%")); err != nil {
		return Function{}, err
	}
	if fn.Return == "" {
		fn.Return = inferredReturn(fn)
	}
	return fn, nil
}

func applyIntrinsicSpec(moduleName string, fn *Function, spec string) error {
	if strings.HasPrefix(spec, "go:") {
		binding, err := parseGoBinding(strings.TrimPrefix(spec, "go:"))
		if err != nil {
			return fmt.Errorf("%s.%s: %w", moduleName, fn.Name, err)
		}
		fn.Go = binding
		if moduleName == "io" {
			fn.Variadic = true
		}
		return nil
	}

	fn.Intrinsic = spec
	if spec == "go.import" {
		fn.TopLevelOnly = true
	}
	return nil
}

func parseGoBinding(spec string) (*GoBinding, error) {
	if spec == "" {
		return nil, fmt.Errorf("empty Go binding")
	}
	if importPath, symbol, ok := strings.Cut(spec, ":"); ok {
		if importPath == "" || symbol == "" {
			return nil, fmt.Errorf("invalid Go binding %q", spec)
		}
		return &GoBinding{Import: importPath, Symbol: symbol}, nil
	}
	lastSlash := strings.LastIndexByte(spec, '/')
	dot := strings.IndexByte(spec[lastSlash+1:], '.')
	if dot < 0 {
		return nil, fmt.Errorf("Go binding %q must include a selector", spec)
	}
	importPath := spec[:lastSlash+1+dot]
	return &GoBinding{Import: importPath, Symbol: spec[lastSlash+1:]}, nil
}

func inferredReturn(fn Function) string {
	switch fn.Intrinsic {
	case "go.expr":
		return "Dynamic"
	default:
		return "Void"
	}
}

func (r *Registry) Function(moduleName string, functionName string) (*Function, bool) {
	if r == nil {
		return nil, false
	}
	mod := r.Modules[moduleName]
	if mod == nil {
		return nil, false
	}
	fn := mod.byName[functionName]
	return fn, fn != nil
}

func (r *Registry) ModuleNames() []string {
	if r == nil {
		return nil
	}
	names := make([]string, 0, len(r.Modules))
	for name := range r.Modules {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func findCoreRoot() (string, error) {
	if root := os.Getenv("RUNE_ROOT"); root != "" {
		return filepath.Join(root, "core"), nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "core")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("core directory not found; run from the Rune repo or set RUNE_ROOT")
}
