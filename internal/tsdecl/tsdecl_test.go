package tsdecl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseInterfacesAndMembers(t *testing.T) {
	src := `
interface Event {
    readonly type: string;
    preventDefault(): void;
}

interface MouseEvent extends Event {
    clientX: number;
    clientY: number;
}

interface KeyboardEvent extends Event {
    readonly key: string;
    readonly code: string;
    repeat: boolean;
}

interface Document {
    createElement(tagName: string): HTMLElement;
    querySelector(selectors: string): Element | null;
}

/** global doc */
declare var document: Document;
declare function fetch(input: string, init?: number): Promise<number>;
`
	dom := Parse(src)
	if dom.Interfaces["Event"] == nil {
		t.Fatal("Event not parsed")
	}
	if got := dom.Interfaces["Event"].Members[0].Name; got != "type" {
		t.Fatalf("Event first member = %q, want type", got)
	}
	got := TypeName(dom.Interfaces["Event"].Members[0].Type)
	if got != "String" {
		t.Fatalf("Event.type type = %q, want String", got)
	}
	if got := dom.Interfaces["MouseEvent"].Bases; len(got) != 1 || got[0] != "Event" {
		t.Fatalf("MouseEvent bases = %#v", got)
	}
	kb := dom.Interfaces["KeyboardEvent"]
	found := false
	for _, m := range kb.Members {
		if m.Name == "key" {
			found = true
			if TypeName(m.Type) != "String" {
				t.Fatalf("key type = %q", TypeName(m.Type))
			}
		}
	}
	if !found {
		t.Fatal("KeyboardEvent.key not found")
	}
	doc := dom.Interfaces["Document"]
	if doc == nil {
		t.Fatal("Document not parsed")
	}
	flat := FlattenedMembers(dom, "Document")
	var ce *Member
	for i := range flat {
		if flat[i].Name == "createElement" {
			ce = &flat[i]
		}
	}
	if ce == nil || !ce.IsMethod {
		t.Fatalf("createElement missing, flat: %#v", flat)
	}
	if TypeName(ce.Type) != "HTMLElement" {
		t.Fatalf("createElement return = %q", TypeName(ce.Type))
	}
	if len(ce.Params) != 1 || ce.Params[0].Name != "tagName" || TypeName(ce.Params[0].Type) != "String" {
		t.Fatalf("createElement params = %#v", ce.Params)
	}
	if dom.Globals["document"] == nil {
		t.Fatal("document global missing")
	}
	if got := TypeName(dom.Globals["document"].Type); got != "Document" {
		t.Fatalf("document type = %q", got)
	}
	if len(dom.Functions) != 1 || dom.Functions[0].Name != "fetch" {
		t.Fatalf("functions = %#v", dom.Functions)
	}
}

func TestParseUnionAndInheritance(t *testing.T) {
	src := `
interface Node { readonly nodeType: number; }
interface ParentNode { children: HTMLCollection; }
interface Element extends Node, ParentNode {
    id: string;
    classList: DOMTokenList;
    closest(selectors: string): Element | null;
}
interface HTMLElement extends Element {
    innerText: string;
    onclick: ((ev: MouseEvent) => void) | null;
}
`
	dom := Parse(src)
	flat := FlattenedMembers(dom, "HTMLElement")
	names := map[string]bool{}
	for _, m := range flat {
		names[m.Name] = true
	}
	for _, want := range []string{"nodeType", "children", "id", "classList", "closest", "innerText"} {
		if !names[want] {
			t.Fatalf("HTMLElement missing %q (have %v)", want, names)
		}
	}
	var oc *Member
	for i := range flat {
		if flat[i].Name == "onclick" {
			oc = &flat[i]
		}
	}
	if oc == nil {
		t.Fatal("onclick not found in flattened members")
	}
	union, ok := oc.Type.(Union)
	if !ok {
		t.Fatalf("onclick type is %T, want Union", oc.Type)
	}
	if len(union.Options) != 2 {
		t.Fatalf("onclick union size = %d", len(union.Options))
	}
	cb, ok := union.Options[0].(Callback)
	if !ok {
		t.Fatalf("onclick branch is %T, want Callback", union.Options[0])
	}
	if len(cb.Params) != 1 || TypeName(cb.Params[0].Type) != "MouseEvent" {
		t.Fatalf("callback params = %#v", cb.Params)
	}
}

func TestLoadFileMissing(t *testing.T) {
	_, err := LoadFile(filepath.Join(os.TempDir(), "no-such-file.d.ts"))
	if err == nil {
		t.Fatal("LoadFile error expected")
	}
}

func TestRealLibDOM(t *testing.T) {
	const path = "/Applications/Visual Studio Code.app/Contents/Resources/app/extensions/node_modules/typescript/lib/lib.dom.d.ts"
	if _, err := os.Stat(path); err != nil {
		t.Skip("lib.dom.d.ts not installed on this machine")
	}
	dom, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if len(dom.Interfaces) < 500 {
		t.Fatalf("parsed only %d interfaces; want far more", len(dom.Interfaces))
	}
	if dom.Interfaces["Document"] == nil {
		t.Fatal("Document missing")
	}
	if dom.Interfaces["HTMLElement"] == nil {
		t.Fatal("HTMLElement missing")
	}
	if dom.Interfaces["KeyboardEvent"] == nil {
		t.Fatal("KeyboardEvent missing")
	}
	if dom.Globals["document"] == nil {
		t.Fatal("document global missing")
	}
	fetch := false
	for _, fn := range dom.Functions {
		if fn.Name == "fetch" {
			fetch = true
		}
	}
	if !fetch {
		t.Fatal("fetch missing")
	}
	flat := FlattenedMembers(dom, "HTMLElement")
	var add *Member
	for i := range flat {
		if flat[i].Name == "addEventListener" {
			add = &flat[i]
		}
	}
	if add == nil {
		t.Fatal("addEventListener should resolve through EventTarget; got none")
	}
	for _, m := range flat {
		if m.Name == "style" {
			if !strings.HasPrefix(TypeName(m.Type), "CSS") {
				t.Logf("style type = %q", TypeName(m.Type))
			}
		}
	}
}
