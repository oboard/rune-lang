package checker

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/oboard/rune-lang/internal/tsdecl"
)

// Ambient DOM library — parsed once from lib.dom.d.ts at first analysis and
// cached for the lifetime of the process. The path is resolved via the
// RUNE_DOM_LIB_PATH override, falling back to the VS Code builtin location or
// a project-local copy. Loading is silent when no declaration file exists so
// headless CI and offline trees keep working.
var (
	domLibOnce sync.Once
	domLib     *tsdecl.DOM
	domFlat    map[string]*tsdecl.Flattened
)

// loadDOMLib lazily parses lib.dom.d.ts and builds the flat view.
func loadDOMLib() (*tsdecl.DOM, map[string]*tsdecl.Flattened) {
	domLibOnce.Do(func() {
		path := locateDOMLib()
		if path == "" {
			return
		}
		d, err := tsdecl.LoadFile(path)
		if err != nil {
			return
		}
		domLib = d
		domFlat = tsdecl.FlattenAll(d)
	})
	return domLib, domFlat
}

// SetDOMLibForTest overrides the ambient DOM with a custom parse (used by
// tests). Pass nil to clear.
func SetDOMLibForTest(d *tsdecl.DOM) {
	domLib = d
	if d == nil {
		domFlat = nil
	} else {
		domFlat = tsdecl.FlattenAll(d)
	}
}

// LocateDOMLib exposes the resolved path so tests and LSP can guard on
// availability. Returns "" when no lib.dom.d.ts is reachable.
func LocateDOMLib() string {
	return locateDOMLib()
}

// locateDOMLib searches for the lib.dom.d.ts declaration file.
//
// Search order:
//  1. RUNE_DOM_LIB_PATH — explicit absolute/relative path.
//  2. NEXT_AUTOCRAT — Visual Studio Code's bundled TypeScript lib.
//  3. lib.dom.d.ts beside the running executable (project layouts).
//  4. node_modules/typescript/lib/lib.dom.d.ts relative to the workspace.
func locateDOMLib() string {
	if env := os.Getenv("RUNE_DOM_LIB_PATH"); env != "" {
		if _, err := os.Stat(env); err == nil {
			return env
		}
	}
	candidates := []string{}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "lib.dom.d.ts"))
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "lib.dom.d.ts"),
			filepath.Join(cwd, "node_modules", "typescript", "lib", "lib.dom.d.ts"),
		)
	}
	candidates = append(candidates,
		"/Applications/Visual Studio Code.app/Contents/Resources/app/extensions/node_modules/typescript/lib/lib.dom.d.ts",
	)
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// collectDOMTypes attaches ambient DOM interfaces, globals, and functions to
// checker.Info so downstream inference can resolve them like user types.
// Types already populated by the core registry or user file win, so users
// can shadow DOM names (e.g. their own HTMLElement implementation).
func (c *checker) collectDOMTypes() {
	dom, flat := loadDOMLib()
	if dom == nil {
		return
	}
	for name := range dom.Interfaces {
		if c.info.Types[name] != nil {
			continue
		}
		c.info.Types[name] = &StructInfo{
			Name:       name,
			SourcePath: "",
			ByName:     map[string]FieldInfo{},
			Methods:    map[string]*FuncInfo{},
		}
	}
	for name, flatView := range flat {
		c.populateDOMInterface(name, flatView)
	}
	for name, v := range dom.Globals {
		if c.info.valuesByName[name] != nil {
			continue
		}
		c.info.ExternalValues = append(c.info.ExternalValues,
			&ExternalValueInfo{
				Name: name,
				Type: c.domToRuneType(v.Type),
			})
		c.info.valuesByName[name] = &ExternalValueInfo{
			Name: name,
			Type: c.domToRuneType(v.Type),
		}
	}
	// Group overloads by name so each DOM function is injected exactly once
	// with its preferred signature.
	byName := map[string][]*tsdecl.Function{}
	for _, fn := range dom.Functions {
		byName[fn.Name] = append(byName[fn.Name], fn)
	}
	for name, group := range byName {
		// Don't stomp user-defined functions of the same name.
		if existing := c.info.functionsByName[name]; len(existing) > 0 {
			continue
		}
		best := group[0]
		bestScore := overloadScore(best)
		for _, ov := range group[1:] {
			if s := overloadScore(ov); s > bestScore {
				best = ov
				bestScore = s
			}
		}
		info := c.domFunctionToFuncInfo(best)
		c.info.functionsByName[name] = append(c.info.functionsByName[name], info)
		if c.info.Functions[name] == nil {
			c.info.Functions[name] = info
		}
	}
}

// domFunctionToFuncInfo collapses TypeScript overloads into a single FuncInfo
// representative. Within a name group we prefer:
//  1. Overloads whose return type is NOT Any (since generic-param erasure
//     can degrade them) — this picks the user-friendly overload.
//  2. If several have a concrete return, the one with the most required
//     args is treated as the canonical signature.
func (c *checker) domFunctionToFuncInfo(fn *tsdecl.Function) *FuncInfo {
	best := fn
	bestScore := overloadScore(fn)
	for _, ov := range domOverloads(fn) {
		s := overloadScore(ov)
		if s > bestScore {
			best = ov
			bestScore = s
		}
	}
	info := &FuncInfo{
		Name:       best.Name,
		SourcePath: "",
		Return:     c.domToRuneType(best.Return),
	}
	required := 0
	for _, p := range best.Params {
		if p.Optional || p.Rest {
			break
		}
		required++
	}
	info.MinRequired = required
	for _, p := range best.Params {
		info.Params = append(info.Params, ParamInfo{
			Name: p.Name,
			Type: c.domToRuneType(p.Type),
			Rest: p.Rest,
		})
		if p.Rest {
			info.Variadic = true
		}
	}
	return info
}

// memberOverloadScore mirrors overloadScore for interface-method overloads —
// a Member stores its return in Type and params in Params, the Function
// analogue for globals.
func memberOverloadScore(mem *tsdecl.Member) int {
	score := 0
	if _, isAny := mem.Type.(tsdecl.Any); !isAny {
		score += 2
	}
	for _, p := range mem.Params {
		if p.Optional {
			break
		}
		if _, isAny := p.Type.(tsdecl.Any); isAny {
			return score
		}
		score++
	}
	return score
}

// overloadScore ranks a TS function for collapse. Higher is better.
//   +2 for a concrete (non-Any) return type
//   +1 if all required params are non-Any
func overloadScore(fn *tsdecl.Function) int {
	score := 0
	if _, isAny := fn.Return.(tsdecl.Any); !isAny {
		score += 2
	}
	for _, p := range fn.Params {
		if p.Optional {
			break
		}
		if _, isAny := p.Type.(tsdecl.Any); isAny {
			return score
		}
		score++
	}
	return score
}

// domOverloads returns all overload variants of a function in the DOM. The
// input fn is the first entry; the remaining variants live in dom.Functions.
// Since parseFunctions only captures the LAST overload per name, this is
// currently a no-op; kept for when overload support lands.
func domOverloads(fn *tsdecl.Function) []*tsdecl.Function {
	return []*tsdecl.Function{fn}
}

// countRequiredParams returns how many positional arguments must be supplied.
// In TS an optional parameter (marked '??') may be omitted; leading required
// params still count.
func countRequiredParams(fn *tsdecl.Function) int {
	count := 0
	for _, p := range fn.Params {
		if p.Optional {
			break
		}
		count++
	}
	return count
}

// populateDOMInterface fills one ambient StructInfo with flattened fields
// and methods. Subtypes shadow base members.
func (c *checker) populateDOMInterface(name string, flat *tsdecl.Flattened) {
	info := c.info.Types[name]
	if info == nil {
		return
	}
	// Group method overloads by name, then pick a single preferred signature
	// per name using overloadScore heuristics.
	seenGroups := map[string][]tsdecl.Member{}
	for _, mem := range flat.Members {
		if !mem.IsMethod {
			continue
		}
		seenGroups[mem.Name] = append(seenGroups[mem.Name], mem)
	}
	processed := map[string]bool{}
	for _, mem := range flat.Members {
		if info.ByName[mem.Name].Name != "" {
			continue
		}
		if processed[mem.Name] && mem.IsMethod {
			continue
		}
		if mem.IsMethod {
			processed[mem.Name] = true
			members := seenGroups[mem.Name]
			best := members[0]
			bestScore := memberOverloadScore(&best)
			for _, ov := range members[1:] {
				if s := memberOverloadScore(&ov); s > bestScore {
					best = ov
					bestScore = s
				}
			}
			mem = best
		}
		runType := c.domToRuneType(mem.Type)
		if mem.IsMethod {
			required := 0
			variadic := false
			for _, p := range mem.Params {
				if p.Optional || p.Rest {
					if p.Rest {
						variadic = true
					}
					break
				}
				required++
			}
			params := make([]ParamInfo, 0, len(mem.Params))
			for _, p := range mem.Params {
				params = append(params, ParamInfo{
					Name: p.Name,
					Type: c.domToRuneType(p.Type),
					Rest: p.Rest,
				})
			}
			info.Methods[mem.Name] = &FuncInfo{
				Name:        mem.Name,
				SourcePath:  "",
				Return:      runType,
				Params:      params,
				MinRequired: required,
				Variadic:    variadic,
			}
		} else {
			info.ByName[mem.Name] = FieldInfo{Name: mem.Name, Type: runType, SourcePath: ""}
			info.Fields = append(info.Fields, info.ByName[mem.Name])
		}
	}
}

// domToRuneType maps the tsdecl type model back to Rune checker.Type.
// Unknown DOM shapes degrade to Dynamic so errors stay contained.
func (c *checker) domToRuneType(t tsdecl.Type) Type {
	if t == nil {
		return Unknown
	}
	switch v := t.(type) {
	case tsdecl.Any:
		return Unknown
	case tsdecl.Primitive:
		switch v.Name {
		case "String":
			return String
		case "Bool":
			return Bool
		case "Double":
			return Double
		case "BigInt":
			return BigInt
		case "Void":
			return Void
		case "Null":
			return Null
		case "Object":
			return Object
		}
		return Unknown
	case tsdecl.Ref:
		// DOM interface refs become the Rune type name when registered.
		if c != nil && c.info != nil && c.info.Types[v.Name] != nil {
			return Type(v.Name)
		}
		return Unknown
	case tsdecl.Literal:
		return String
	case tsdecl.Array:
		return ArrayOf(c.domToRuneType(v.Elem))
	case tsdecl.Union:
		// Drop null/undefined leaves; single concrete alternative survives.
		var survivors []Type
		for _, opt := range v.Options {
			t := c.domToRuneType(opt)
			if t == Null {
				continue
			}
			survivors = append(survivors, t)
		}
		if len(survivors) == 1 {
			return survivors[0]
		}
		return Unknown
	case tsdecl.Intersect:
		// Take the first concrete option (e.g. `Window & typeof globalThis`).
		for _, opt := range v.Options {
			if t := c.domToRuneType(opt); t != Unknown {
				return t
			}
		}
		return Unknown
	case tsdecl.Callback:
		params := make([]Type, 0, len(v.Params))
		for _, p := range v.Params {
			params = append(params, c.domToRuneType(p.Type))
		}
		return AsyncFuncOfTypes(params, c.domToRuneType(v.Return))
	}
	return Unknown
}
