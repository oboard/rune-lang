package lsp

import "testing"

func TestTernaryExpressionDiagnostics(t *testing.T) {
	uri := "file:///tmp/if.rn"
	src := `sign(value: Int32) -> String => (value > 0) ? {
  "positive"
} : (value == 0) ? {
  "zero"
} : {
  "negative"
}

main() => @io.println(sign(1))`
	s := NewSession()
	s.SetDocument(uri, src)
	if diags := s.Diagnostics(uri); len(diags) != 0 {
		t.Fatalf("Diagnostics() = %#v, want none", diags)
	}
}
