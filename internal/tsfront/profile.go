package tsfront

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Stable RFC 0001 D7 diagnostic codes. The profile rejects dynamic behaviour
// at conversion time; it never lowers it to a best-effort approximation.
const (
	CodeDynamicEval             = "RUNE-DYN-001"
	CodeDynamicFunction         = "RUNE-DYN-002"
	CodeDynamicWith             = "RUNE-DYN-003"
	CodeDynamicProxy            = "RUNE-DYN-004"
	CodeDynamicPrototype        = "RUNE-DYN-005"
	CodeDynamicImport           = "RUNE-DYN-006"
	CodeDynamicArguments        = "RUNE-DYN-007"
	CodeDynamicImplicitAny      = "RUNE-DYN-008"
	CodeDynamicAnyMember        = "RUNE-DYN-009"
	CodeDynamicGlobalReflection = "RUNE-DYN-010"
	CodeDynamicDebugger         = "RUNE-DYN-011"
	CodeDynamicRequire          = "RUNE-DYN-012"
	CodeDynamicThisBinding      = "RUNE-DYN-013"
	CodeProfileSyntax           = "RUNE-PROFILE-001"
	CodeProfileClass            = "RUNE-PROFILE-002"
)

// scanProfileViolations walks the raw ESTree JSON and reports every construct
// that RFC 0001 D7 keeps out of the language. Running this before conversion
// guarantees that users see the documented code even when the surrounding
// construct is otherwise unsupported.
func scanProfileViolations(statements []map[string]any, path string) []Diagnostic {
	scanner := &profileScanner{path: path}
	for _, stmt := range statements {
		scanner.node(stmt)
	}
	sort.SliceStable(scanner.diags, func(i, j int) bool {
		if scanner.diags[i].Line != scanner.diags[j].Line {
			return scanner.diags[i].Line < scanner.diags[j].Line
		}
		return scanner.diags[i].Column < scanner.diags[j].Column
	})
	return scanner.diags
}

type profileScanner struct {
	path  string
	diags []Diagnostic
	seen  map[string]bool
}

func (s *profileScanner) add(node map[string]any, code, format string, args ...any) {
	line, column := nodePosition(node)
	key := code + ":" + nodeKey(line, column) + ":" + format
	if s.seen == nil {
		s.seen = map[string]bool{}
	}
	if s.seen[key] {
		return
	}
	s.seen[key] = true
	s.diags = append(s.diags, Diagnostic{
		Code:    code,
		Message: "[" + code + "] " + fmt.Sprintf(format, args...),
		Line:    line,
		Column:  column,
		Path:    s.path,
	})
}

func nodeKey(line, column int) string {
	return strconv.Itoa(line) + ":" + strconv.Itoa(column)
}

func (s *profileScanner) node(node map[string]any) {
	if node == nil {
		return
	}
	kind := stringField(node, "type")
	switch kind {
	case "WithStatement":
		s.add(node, CodeDynamicWith, "`with` statements are not supported")
	case "ImportExpression":
		s.add(node, CodeDynamicImport, "dynamic `import()` is not supported; use a static import")
	case "DebuggerStatement":
		s.add(node, CodeDynamicDebugger, "`debugger` statements are not supported")
	case "YieldExpression":
		s.add(node, CodeProfileSyntax, "generators and `yield` are outside the Rune TS profile")
	case "TSDeclareFunction", "TSModuleDeclaration":
		s.add(node, CodeProfileSyntax, "`declare` and `namespace` are outside the Rune TS profile")
	case "CallExpression", "NewExpression":
		s.checkCall(node, kind)
	case "AssignmentExpression":
		s.checkAssignment(node)
	case "MemberExpression":
		s.checkMember(node)
	}
	if decorators := arrayField(node, "decorators"); len(decorators) > 0 {
		s.add(node, CodeProfileSyntax, "decorators are outside the Rune TS profile")
	}
	for key, value := range node {
		switch key {
		case "type", "loc", "range", "start", "end", "raw":
			continue
		}
		s.value(value)
	}
}

func (s *profileScanner) value(value any) {
	switch v := value.(type) {
	case map[string]any:
		s.node(v)
	case []any:
		for _, item := range v {
			s.value(item)
		}
	}
}

func (s *profileScanner) checkCall(node map[string]any, kind string) {
	callee, ok := mapField(node, "callee")
	if !ok {
		return
	}
	name := identifierName(callee)
	if name == "arguments" {
		s.add(node, CodeDynamicArguments, "`arguments` is not supported; use rest parameters")
	}
	switch name {
	case "eval":
		s.add(node, CodeDynamicEval, "`eval` is not supported")
	case "import":
		s.add(node, CodeDynamicImport, "dynamic `import()` is not supported; use a static import")
	case "Function":
		s.add(node, CodeDynamicFunction, "`Function` construction is not supported")
	case "Proxy":
		s.add(node, CodeDynamicProxy, "`Proxy` is not supported")
	case "require":
		s.add(node, CodeDynamicRequire, "`require()` is not supported; use a static import")
	}
	if kind == "NewExpression" && name == "Function" {
		s.add(node, CodeDynamicFunction, "`new Function` is not supported")
	}
	if stringField(callee, "type") == "MemberExpression" {
		object := identifierNameField(callee, "object")
		property := memberPropertyName(callee)
		if object == "Reflect" {
			s.add(node, CodeDynamicProxy, "`Reflect` traps are not supported")
		}
		if object == "Object" && property == "setPrototypeOf" {
			s.add(node, CodeDynamicPrototype, "runtime prototype mutation is not supported")
		}
		if object == "Object" && property == "defineProperty" {
			s.add(node, CodeDynamicPrototype, "runtime property definition is not supported")
		}
		if property == "call" || property == "apply" {
			s.add(node, CodeDynamicThisBinding, "`call`/`apply` this-rebinding is not supported")
		}
	}
}

func (s *profileScanner) checkAssignment(node map[string]any) {
	left, ok := mapField(node, "left")
	if !ok {
		return
	}
	if stringField(left, "type") == "MemberExpression" && memberPropertyName(left) == "__proto__" {
		s.add(node, CodeDynamicPrototype, "runtime prototype mutation is not supported")
	}
	if right, ok := mapField(node, "right"); ok && identifierName(right) == "arguments" {
		s.add(node, CodeDynamicArguments, "`arguments` is not supported; use rest parameters")
	}
}

func (s *profileScanner) checkMember(node map[string]any) {
	if identifierNameField(node, "object") == "arguments" {
		s.add(node, CodeDynamicArguments, "`arguments` is not supported; use rest parameters")
	}
	computed, _ := node["computed"].(bool)
	if !computed {
		return
	}
	object := identifierNameField(node, "object")
	if object == "globalThis" || object == "window" || object == "self" {
		s.add(node, CodeDynamicGlobalReflection, "runtime string reflection on globals is not supported")
	}
}

func identifierName(node map[string]any) string {
	if stringField(node, "type") != "Identifier" {
		return ""
	}
	return stringField(node, "name")
}

func identifierNameField(node map[string]any, key string) string {
	inner, ok := mapField(node, key)
	if !ok {
		return ""
	}
	return identifierName(inner)
}

func memberPropertyName(node map[string]any) string {
	property, ok := mapField(node, "property")
	if !ok || stringField(property, "type") != "Identifier" {
		return ""
	}
	return stringField(property, "name")
}

// parseErrorCode maps parser-level strict-mode rejections back to the D7 code
// that describes why the construct is absent from the profile.
func parseErrorCode(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "with' statements are not allowed"):
		return CodeDynamicWith
	case strings.Contains(lower, "dynamic import"):
		return CodeDynamicImport
	case strings.Contains(lower, "eval"):
		return CodeDynamicEval
	}
	return ""
}

// scanDeclareStatements rejects `declare` before parsing. The vendored
// TypeScript parser does not preserve a reliable `declare` wrapper for every
// declaration form, so the profile gate is source-level and intentionally
// narrow: a line whose first token is `declare` outside comments/strings.
func scanDeclareStatements(source, path string) []Diagnostic {
	var diags []Diagnostic
	inBlockComment := false
	for index, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if inBlockComment {
			if end := strings.Index(trimmed, "*/"); end >= 0 {
				inBlockComment = false
				trimmed = strings.TrimSpace(trimmed[end+2:])
			} else {
				continue
			}
		}
		if strings.HasPrefix(trimmed, "/*") {
			if !strings.Contains(trimmed, "*/") {
				inBlockComment = true
			}
			continue
		}
		if strings.HasPrefix(trimmed, "//") || trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) == 0 || fields[0] != "declare" {
			continue
		}
		diags = append(diags, Diagnostic{
			Code:    CodeProfileSyntax,
			Message: "[" + CodeProfileSyntax + "] `declare` is outside the Rune TS profile",
			Line:    index + 1,
			Column:  1,
			Path:    path,
		})
	}
	return diags
}
