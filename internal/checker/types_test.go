package checker

import "testing"

func TestDisplayTypeFormatsFunctionTypes(t *testing.T) {
	got := DisplayType(Type("Func[{a: Int32},{k: Int32}]"))
	want := "{a: Int32} -> {k: Int32}"
	if got != want {
		t.Fatalf("DisplayType() = %q, want %q", got, want)
	}
}
