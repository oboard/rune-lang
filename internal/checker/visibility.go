package checker

import "github.com/oboard/rune-lang/internal/lexer"

func (c *checker) withSourcePath(path string, fn func()) {
	prev := c.currentSourcePath
	if normalized := normalizeSourcePath(path); normalized != "" {
		c.currentSourcePath = normalized
	}
	fn()
	c.currentSourcePath = prev
}

func (c *checker) canAccessPrivate(private bool, sourcePath string) bool {
	if !private {
		return true
	}
	declPath := normalizeSourcePath(sourcePath)
	usePath := normalizeSourcePath(c.currentSourcePath)
	return declPath == "" || usePath == "" || declPath == usePath
}

func (c *checker) checkPrivateAccess(kind string, name string, private bool, sourcePath string, pos lexer.Position) bool {
	if c.canAccessPrivate(private, sourcePath) {
		return true
	}
	c.errorf(pos, "%s %q is private", kind, name)
	return false
}

func sameSourcePath(left string, right string) bool {
	return normalizeSourcePath(left) == normalizeSourcePath(right)
}

func (c *checker) resolveFunction(name string, pos lexer.Position) (*FuncInfo, bool) {
	var public *FuncInfo
	var inaccessible *FuncInfo
	for _, fn := range c.info.functionsByName[name] {
		if fn.Private {
			if c.canAccessPrivate(true, fn.SourcePath) {
				return fn, true
			}
			if inaccessible == nil {
				inaccessible = fn
			}
			continue
		}
		if public == nil {
			public = fn
		}
	}
	if public != nil {
		return public, true
	}
	if inaccessible != nil {
		c.checkPrivateAccess("function", name, true, inaccessible.SourcePath, pos)
		return nil, true
	}
	return nil, false
}

func (c *checker) resolveExternalValue(name string) *ExternalValueInfo {
	if c.info == nil {
		return nil
	}
	return c.info.valuesByName[name]
}

func (c *checker) reportUnknownOrPrivateType(pos lexer.Position, name string) {
	if privateName, ok := c.inaccessibleTypeName(name); ok {
		c.errorf(pos, "type %q is private", privateName)
		return
	}
	c.errorf(pos, "unknown type %q", name)
}

func (c *checker) inaccessibleTypeName(name string) (string, bool) {
	if info := c.info.Types[name]; info != nil && !c.canAccessPrivate(info.Private, info.SourcePath) {
		return name, true
	}
	if info := c.info.Enums[name]; info != nil && !c.canAccessPrivate(info.Private, info.SourcePath) {
		return name, true
	}
	if inner, ok := parseNullableType(name); ok {
		return c.inaccessibleTypeName(inner)
	}
	if elem, ok := parseArrayType(name); ok {
		return c.inaccessibleTypeName(elem)
	}
	if base, args, ok := parseGenericType(name); ok {
		if privateName, ok := c.inaccessibleTypeName(base); ok {
			return privateName, true
		}
		for _, arg := range args {
			if privateName, ok := c.inaccessibleTypeName(arg); ok {
				return privateName, true
			}
		}
	}
	if params, ret, ok := parseFuncType(name); ok {
		for _, param := range params {
			if privateName, ok := c.inaccessibleTypeName(param); ok {
				return privateName, true
			}
		}
		return c.inaccessibleTypeName(ret)
	}
	return "", false
}
