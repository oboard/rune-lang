// Package libenv loads TypeScript declaration libraries (`lib.*.d.ts`) and
// exposes them as one merged environment for the checker, LSP, and code
// generators.
//
// It is the foundation for RFC 0001: Rune does not define a standard library
// of its own, so TypeScript declaration files are the single source of truth
// for built-in APIs. The loader is deliberately tolerant; declaration shapes
// that the lightweight tsdecl model cannot represent degrade to Dynamic.
package libenv

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/oboard/rune-lang/internal/tsdecl"
)

// Capability identifies the host surface a declaration belongs to.
type Capability string

const (
	// CapESCore is available on every backend (Array, JSON, Math, ...).
	CapESCore Capability = "es.core"
	// CapDOM is only available on the TypeScript/DOM backend.
	CapDOM Capability = "dom"
	// CapNode is available on the TypeScript and Go/MoonBit Node shims.
	CapNode Capability = "node"
)

// Set is the merged view of one or more declaration libraries.
type Set struct {
	Version      string
	Files        []string
	Interfaces   map[string]*tsdecl.Interface
	Aliases      map[string]*tsdecl.Alias
	Globals      map[string]*tsdecl.Value
	Functions    map[string][]*tsdecl.Function
	Order        []string
	Capabilities map[string]Capability
}

// New returns an empty environment.
func New() *Set {
	return &Set{
		Interfaces:   map[string]*tsdecl.Interface{},
		Aliases:      map[string]*tsdecl.Alias{},
		Globals:      map[string]*tsdecl.Value{},
		Functions:    map[string][]*tsdecl.Function{},
		Capabilities: map[string]Capability{},
	}
}

// Load parses the given declaration files and merges them in order. Later
// files win on conflicting single-valued declarations; interface members are
// unioned so an extended lib can add APIs without dropping the base surface.
func Load(paths ...string) (*Set, error) {
	set := New()
	for _, path := range paths {
		dom, err := tsdecl.LoadFile(path)
		if err != nil {
			return nil, fmt.Errorf("libenv: load %s: %w", path, err)
		}
		capability := capabilityForFile(path)
		set.merge(dom, capability)
		set.Files = append(set.Files, path)
		if set.Version == "" {
			set.Version = detectVersion(path)
		}
	}
	return set, nil
}

// LoadDefault loads the pinned/vendored library set when available, falling
// back to a local TypeScript installation. It loads lib.es5.d.ts and, when
// present, lib.dom.d.ts. Missing optional files are ignored.
func LoadDefault() (*Set, error) {
	dir := DefaultLibDir()
	if dir == "" {
		return nil, fmt.Errorf("libenv: no TypeScript lib directory found; set RUNE_TS_LIB_DIR or vendor tools/ts-lib")
	}
	required := filepath.Join(dir, "lib.es5.d.ts")
	if _, err := os.Stat(required); err != nil {
		return nil, fmt.Errorf("libenv: %s not found in %s", filepath.Base(required), dir)
	}
	paths := []string{required}
	optional := []string{
		filepath.Join(dir, "lib.dom.d.ts"),
		filepath.Join(dir, "lib.es2015.d.ts"),
	}
	for _, path := range optional {
		if _, err := os.Stat(path); err == nil {
			paths = append(paths, path)
		}
	}
	return Load(paths...)
}

// DefaultLibDir returns the first existing TypeScript lib directory. The
// RUNE_TS_LIB_DIR override exists for tests and for users with a custom
// TypeScript checkout.
func DefaultLibDir() string {
	if dir := strings.TrimSpace(os.Getenv("RUNE_TS_LIB_DIR")); dir != "" {
		return dir
	}
	candidates := []string{
		"tools/ts-lib",
		"node_modules/typescript/lib",
		"/Applications/Visual Studio Code.app/Contents/Resources/app/extensions/node_modules/typescript/lib",
		"/usr/local/lib/node_modules/typescript/lib",
		"/usr/lib/node_modules/typescript/lib",
	}
	if cwd, err := os.Getwd(); err == nil {
		for {
			for _, candidate := range candidates {
				path := candidate
				if !filepath.IsAbs(candidate) {
					path = filepath.Join(cwd, candidate)
				}
				if info, err := os.Stat(path); err == nil && info.IsDir() {
					if _, err := os.Stat(filepath.Join(path, "lib.es5.d.ts")); err != nil {
						continue
					}
					return path
				}
			}
			parent := filepath.Dir(cwd)
			if parent == cwd {
				break
			}
			cwd = parent
		}
	}
	return ""
}

// Interface returns the named interface declaration, or nil.
func (s *Set) Interface(name string) *tsdecl.Interface {
	if s == nil {
		return nil
	}
	return s.Interfaces[name]
}

// Alias returns the named type alias declaration, or nil.
func (s *Set) Alias(name string) *tsdecl.Alias {
	if s == nil {
		return nil
	}
	return s.Aliases[name]
}

// Global returns the named global value declaration, or nil.
func (s *Set) Global(name string) *tsdecl.Value {
	if s == nil {
		return nil
	}
	return s.Globals[name]
}

// Function returns the first overload for the named global function, or nil.
func (s *Set) Function(name string) *tsdecl.Function {
	if s == nil {
		return nil
	}
	overloads := s.Functions[name]
	if len(overloads) == 0 {
		return nil
	}
	return overloads[0]
}

// CapabilityOf returns the capability assigned to a declaration name. Unknown
// names default to CapESCore so callers can stay permissive while the
// capability table is filled in.
func (s *Set) CapabilityOf(name string) Capability {
	if s == nil {
		return CapESCore
	}
	if capability, ok := s.Capabilities[name]; ok {
		return capability
	}
	return CapESCore
}

// Names returns the merged declaration order with duplicates removed.
func (s *Set) Names() []string {
	if s == nil {
		return nil
	}
	return append([]string(nil), s.Order...)
}

func (s *Set) merge(dom *tsdecl.DOM, capability Capability) {
	if dom == nil {
		return
	}
	for name, iface := range dom.Interfaces {
		if existing := s.Interfaces[name]; existing != nil {
			existing.Members = mergeMembers(existing.Members, iface.Members)
			existing.Bases = mergeStrings(existing.Bases, iface.Bases)
		} else {
			s.Interfaces[name] = iface
		}
		s.Capabilities[name] = capability
	}
	for name, alias := range dom.Aliases {
		s.Aliases[name] = alias
		s.Capabilities[name] = capability
	}
	for name, value := range dom.Globals {
		s.Globals[name] = value
		s.Capabilities[name] = capability
	}
	for _, fn := range dom.Functions {
		s.Functions[fn.Name] = append(s.Functions[fn.Name], fn)
		if _, ok := s.Capabilities[fn.Name]; !ok || capability == CapDOM {
			s.Capabilities[fn.Name] = capability
		}
	}
	s.Order = append(s.Order, dom.Order...)
}

func mergeMembers(base, extra []tsdecl.Member) []tsdecl.Member {
	seen := make(map[string]bool, len(base))
	for _, member := range base {
		seen[member.Name+"\x00"+member.Signature] = true
	}
	out := append([]tsdecl.Member(nil), base...)
	for _, member := range extra {
		key := member.Name + "\x00" + member.Signature
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, member)
	}
	return out
}

func mergeStrings(base, extra []string) []string {
	seen := make(map[string]bool, len(base))
	for _, value := range base {
		seen[value] = true
	}
	out := append([]string(nil), base...)
	for _, value := range extra {
		if seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func capabilityForFile(path string) Capability {
	name := strings.ToLower(filepath.Base(path))
	switch {
	case strings.Contains(name, "dom"):
		return CapDOM
	case strings.Contains(name, "node"):
		return CapNode
	default:
		return CapESCore
	}
}

func detectVersion(path string) string {
	dir := filepath.Dir(path)
	for depth := 0; depth < 3; depth++ {
		if data, err := os.ReadFile(filepath.Join(dir, "VERSION")); err == nil {
			if version := strings.TrimSpace(string(data)); version != "" {
				return version
			}
		}
		if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
			var pkg struct {
				Version string `json:"version"`
			}
			if json.Unmarshal(data, &pkg) == nil && pkg.Version != "" {
				return pkg.Version
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// SortedNames is a deterministic helper for diagnostics and tests.
func (s *Set) SortedNames() []string {
	names := s.Names()
	sort.Strings(names)
	return names
}
