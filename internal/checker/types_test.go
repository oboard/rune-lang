package checker

import "testing"

func TestDisplayTypeFormatsFunctionTypes(t *testing.T) {
	got := DisplayType(Type("Func[{a: Int},{k: Int}]"))
	want := "{a: Int} -> {k: Int}"
	if got != want {
		t.Fatalf("DisplayType() = %q, want %q", got, want)
	}
}
