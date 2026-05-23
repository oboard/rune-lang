package stdlib

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

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
	p := newStubParser(name, path, src)
	mod, err := p.parse()
	if err != nil {
		return nil, err
	}
	return mod, nil
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

func (r *Registry) ReceiverFunction(moduleName string, receiver string, functionName string) (*Function, bool) {
	if r == nil {
		return nil, false
	}
	mod := r.Modules[moduleName]
	if mod == nil {
		return nil, false
	}
	methods := mod.byReceiver[receiver]
	if methods == nil {
		return nil, false
	}
	fn := methods[functionName]
	return fn, fn != nil
}

func (r *Registry) FunctionByAlias(moduleName string, alias string) (*Function, bool) {
	if r == nil {
		return nil, false
	}
	mod := r.Modules[moduleName]
	if mod == nil {
		return nil, false
	}
	fn := mod.byAlias[alias]
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
