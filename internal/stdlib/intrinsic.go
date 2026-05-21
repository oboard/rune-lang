package stdlib

import (
	"fmt"
	"strings"
)

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
