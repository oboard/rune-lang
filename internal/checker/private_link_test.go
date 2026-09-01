package checker

import "testing"

func TestPrivateLinkNameUsesStablePathAndName(t *testing.T) {
	got := privateLinkName("selfhost/compiler/compiler.rn", "markStart")
	want := "selfhost_compiler_compiler_markStart"
	if got != want {
		t.Fatalf("privateLinkName() = %q, want %q", got, want)
	}
}

func TestPrivateLinkNameUsesProjectRelativePath(t *testing.T) {
	got := privateLinkName("/workspace/rune-lang/selfhost/compiler/compiler.rn", "markStart")
	want := "selfhost_compiler_compiler_markStart"
	if got != want {
		t.Fatalf("privateLinkName() = %q, want %q", got, want)
	}
}

func TestPrivateLinkNamePreservesRelativeExternalPath(t *testing.T) {
	got := privateLinkName("fixtures/helper.rn", "run")
	want := "fixtures_helper_run"
	if got != want {
		t.Fatalf("privateLinkName() = %q, want %q", got, want)
	}
}

func TestPrivateLinkNamePreservesDistinctPaths(t *testing.T) {
	left := privateLinkName("one/shared.rn", "run")
	right := privateLinkName("two/shared.rn", "run")
	if left == right {
		t.Fatalf("private links collided: %q", left)
	}
}
