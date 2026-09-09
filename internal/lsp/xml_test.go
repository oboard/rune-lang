package lsp

import "testing"

func TestXMLNativeTagDefinitionIsNil(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `render() -> HTMLElement => {
  <div></div>
}
`
	s := &server{docs: map[string]string{uri: src}}
	def := s.definition(uri, positionOf(src, "<div", "div"))
	if def != nil {
		t.Fatalf("native tag definition = %#v, want nil", def)
	}
}
