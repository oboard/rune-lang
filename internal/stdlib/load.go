package stdlib

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var (
	defaultRegistryMu sync.Mutex
	defaultRegistry   *Registry
)

func SetDefault(reg *Registry) {
	defaultRegistryMu.Lock()
	defer defaultRegistryMu.Unlock()
	defaultRegistry = reg
}

func LoadDefault() (*Registry, error) {
	defaultRegistryMu.Lock()
	if defaultRegistry != nil {
		reg := defaultRegistry
		defaultRegistryMu.Unlock()
		return reg, nil
	}
	defaultRegistryMu.Unlock()

	root, err := findCoreRoot()
	if err != nil {
		return nil, err
	}
	reg, err := Load(root)
	if err != nil {
		return nil, err
	}

	defaultRegistryMu.Lock()
	defer defaultRegistryMu.Unlock()
	if defaultRegistry != nil {
		return defaultRegistry, nil
	}
	defaultRegistry = reg
	return defaultRegistry, nil
}

func Load(root string) (*Registry, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("load core: %w", err)
	}

	sources := map[string]string{}
	reg := &Registry{Modules: map[string]*Module{}, Types: map[string]*Type{}}
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
		sources[path] = string(data)
	}
	return loadSources(reg, sources)
}

func LoadSources(sources map[string]string) (*Registry, error) {
	return loadSources(&Registry{Modules: map[string]*Module{}, Types: map[string]*Type{}}, sources)
}

func loadSources(reg *Registry, sources map[string]string) (*Registry, error) {
	paths := make([]string, 0, len(sources))
	for sourcePath := range sources {
		paths = append(paths, sourcePath)
	}
	sort.Strings(paths)
	for _, sourcePath := range paths {
		moduleName, ok := moduleNameFromSourcePath(sourcePath)
		if !ok {
			continue
		}
		mod, err := parseModule(moduleName, sourcePath, sources[sourcePath])
		if err != nil {
			return nil, err
		}
		reg.Modules[mod.Name] = mod
		for i := range mod.Types {
			typ := &mod.Types[i]
			reg.Types[typ.Name] = typ
		}
	}
	return reg, nil
}

func moduleNameFromSourcePath(sourcePath string) (string, bool) {
	clean := path.Clean(strings.ReplaceAll(sourcePath, "\\", "/"))
	base := path.Base(clean)
	if path.Ext(base) != ".rn" {
		return "", false
	}
	name := strings.TrimSuffix(base, ".rn")
	if name == "" || path.Base(path.Dir(clean)) != name {
		return "", false
	}
	return name, true
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

func (r *Registry) Type(name string) (*Type, bool) {
	if r == nil {
		return nil, false
	}
	typ := r.Types[name]
	return typ, typ != nil
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
