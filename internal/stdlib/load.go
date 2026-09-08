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
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Only the top-level files of a module directory belong to that
		// module. Nested directories are implementation-private and are not
		// scanned.
		moduleEntries, err := os.ReadDir(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		for _, moduleEntry := range moduleEntries {
			if moduleEntry.IsDir() || filepath.Ext(moduleEntry.Name()) != ".rn" {
				continue
			}
			sourcePath := filepath.Join(root, entry.Name(), moduleEntry.Name())
			data, err := os.ReadFile(sourcePath)
			if err != nil {
				return nil, err
			}
			sources[sourcePath] = string(data)
		}
	}
	return loadCoreSources(root, sources)
}

func loadCoreSources(root string, sources map[string]string) (*Registry, error) {
	reg := &Registry{Modules: map[string]*Module{}, Types: map[string]*Type{}, Traits: map[string]*Trait{}}
	paths := make([]string, 0, len(sources))
	for sourcePath := range sources {
		paths = append(paths, sourcePath)
	}
	sort.Strings(paths)
	for _, sourcePath := range paths {
		relativePath, err := filepath.Rel(root, sourcePath)
		if err != nil {
			return nil, err
		}
		parts := strings.Split(filepath.ToSlash(relativePath), "/")
		if len(parts) < 2 || filepath.Ext(sourcePath) != ".rn" {
			continue
		}
		if err := addModuleSource(reg, parts[0], sourcePath, sources[sourcePath]); err != nil {
			return nil, err
		}
	}
	return reg, nil
}

func LoadSources(sources map[string]string) (*Registry, error) {
	return loadSources(&Registry{Modules: map[string]*Module{}, Types: map[string]*Type{}, Traits: map[string]*Trait{}}, sources)
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
		if err := addModuleSource(reg, moduleName, sourcePath, sources[sourcePath]); err != nil {
			return nil, err
		}
	}
	return reg, nil
}

func addModuleSource(reg *Registry, moduleName string, sourcePath string, source string) error {
	parsed, err := parseModule(moduleName, sourcePath, source)
	if err != nil {
		return err
	}
	mod := reg.Modules[moduleName]
	if mod == nil {
		mod = &Module{
			Name:       moduleName,
			byName:     map[string]*Function{},
			byMacro:    map[string]*Function{},
			byReceiver: map[string]map[string]*Function{},
			byAlias:    map[string]*Function{},
		}
		reg.Modules[moduleName] = mod
	}
	for _, fn := range parsed.Functions {
		if err := addFunction(mod, moduleFunctionKeys(mod), fn); err != nil {
			return err
		}
	}
	for _, typ := range parsed.Types {
		mod.Types = append(mod.Types, typ)
		added := &mod.Types[len(mod.Types)-1]
		reg.Types[added.Name] = added
	}
	for _, trait := range parsed.Traits {
		mod.Traits = append(mod.Traits, trait)
		added := &mod.Traits[len(mod.Traits)-1]
		reg.Traits[added.Name] = added
	}
	indexModuleFunctions(mod)
	return nil
}

func moduleFunctionKeys(mod *Module) map[string]bool {
	seen := map[string]bool{}
	for _, fn := range mod.Functions {
		key := fn.Receiver + "." + fn.Name
		if fn.Macro {
			key = "macro:" + key
		}
		seen[key] = true
	}
	return seen
}

func indexModuleFunctions(mod *Module) {
	clear(mod.byName)
	clear(mod.byMacro)
	clear(mod.byReceiver)
	clear(mod.byAlias)
	for i := range mod.Functions {
		fn := &mod.Functions[i]
		if fn.Macro {
			mod.byMacro[fn.Name] = fn
		} else if fn.Receiver == "" {
			mod.byName[fn.Name] = fn
		} else {
			if _, exists := mod.byName[fn.Name]; !exists {
				mod.byName[fn.Name] = fn
			}
			methods := mod.byReceiver[fn.Receiver]
			if methods == nil {
				methods = map[string]*Function{}
				mod.byReceiver[fn.Receiver] = methods
			}
			methods[fn.Name] = fn
		}
		if fn.Alias != "" {
			mod.byAlias[fn.Alias] = fn
		}
	}
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
	if fn == nil {
		fn = mod.byMacro[functionName]
	}
	return fn, fn != nil
}

func (r *Registry) MacroFunction(moduleName string, functionName string) (*Function, bool) {
	if r == nil {
		return nil, false
	}
	mod := r.Modules[moduleName]
	if mod == nil {
		return nil, false
	}
	fn := mod.byMacro[functionName]
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
