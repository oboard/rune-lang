package lsp

import (
	"strings"
	"testing"
)

func TestMethodDefinitionAndRename(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `User: {
    age: Int

    isAdult() => .age >= 18
}

main() => {
    user := User {
        age: 22
    }
    user.isAdult()
}
`
	s := &server{docs: map[string]string{uri: src}}

	def := s.definition(uri, positionOf(src, "user.isAdult", "isAdult")).(map[string]any)
	if got := def["uri"]; got != uri {
		t.Fatalf("definition uri = %v, want %s", got, uri)
	}
	start := def["range"].(map[string]any)["start"].(position)
	if start.Line != 3 || start.Character != 4 {
		t.Fatalf("definition start = %+v, want line 3 char 4", start)
	}

	edit := s.rename(uri, positionOf(src, "user.isAdult", "isAdult"), "adult").(map[string]any)
	changes := edit["changes"].(map[string]any)[uri].([]map[string]any)
	if len(changes) != 2 {
		t.Fatalf("rename edits = %d, want 2: %#v", len(changes), changes)
	}

	hover := s.hover(uri, positionOf(src, "user.isAdult", "isAdult")).(map[string]any)
	if got := hoverValue(hover); !strings.Contains(got, "User.isAdult() -> Bool") {
		t.Fatalf("hover = %q, want User.isAdult signature", got)
	}
}

func TestArrayMethodDefinitionUsesCoreStub(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `main() => {
    arr := [1, 2, 3]
    arr.map(value => value + 1)
}
`
	s := &server{docs: map[string]string{uri: src}}

	def := s.definition(uri, positionOf(src, "arr.map", "map")).(map[string]any)
	defURI := def["uri"].(string)
	if !strings.HasSuffix(defURI, "/core/array/array.rn") {
		t.Fatalf("definition uri = %s, want core/array/array.rn", defURI)
	}
	start := def["range"].(map[string]any)["start"].(position)
	if start.Line != 9 || start.Character != 4 {
		t.Fatalf("definition start = %+v, want line 9 char 4", start)
	}
	if got := s.rename(uri, positionOf(src, "arr.map", "map"), "collect"); got != nil {
		t.Fatalf("array method rename = %#v, want nil", got)
	}
	hover := s.hover(uri, positionOf(src, "arr.map", "map")).(map[string]any)
	value := hoverValue(hover)
	if !strings.Contains(value, "Array.map[T, U](f: (T) -> U) -> Array[U]") {
		t.Fatalf("hover = %q, want array map signature", value)
	}
	if !strings.Contains(value, "core/array/array.rn") {
		t.Fatalf("hover = %q, want core source path", value)
	}
}

func TestAnonymousObjectHover(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `main() => {
    obj := {
        name: "Alice"
        age: 30

        greet() => @io.println("Hello, my name is " + obj.name)
        nextAge() => .age + 1
    }

    @io.println(obj.name)
    @io.println(obj.age)
    @io.println(obj.nextAge())
    obj.greet()
}
`
	s := &server{docs: map[string]string{uri: src}}

	objHover := s.hover(uri, positionOf(src, "obj :=", "obj")).(map[string]any)
	objValue := hoverValue(objHover)
	wantObject := `obj: {
  name: String
  age: Int
  greet: () -> Void
  nextAge: () -> Int
}`
	if !strings.Contains(objValue, wantObject) {
		t.Fatalf("hover = %q, want %s", objValue, wantObject)
	}

	nameHover := s.hover(uri, positionOf(src, "@io.println(obj.name)", "name")).(map[string]any)
	if got := hoverValue(nameHover); !strings.Contains(got, "name: String") {
		t.Fatalf("hover = %q, want name field type", got)
	}

	nextAgeHover := s.hover(uri, positionOf(src, "obj.nextAge", "nextAge")).(map[string]any)
	if got := hoverValue(nextAgeHover); !strings.Contains(got, "nextAge: () -> Int") {
		t.Fatalf("hover = %q, want nextAge function field type", got)
	}
}

func hoverValue(hover map[string]any) string {
	return hover["contents"].(map[string]any)["value"].(string)
}

func positionOf(src string, context string, needle string) position {
	contextStart := strings.Index(src, context)
	if contextStart < 0 {
		panic("context not found")
	}
	start := contextStart + strings.Index(src[contextStart:], needle)
	if start < contextStart {
		panic("needle not found")
	}
	line := 0
	lineStart := 0
	for i, ch := range src[:start] {
		if ch == '\n' {
			line++
			lineStart = i + 1
		}
	}
	return position{Line: line, Character: start - lineStart}
}
