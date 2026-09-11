package compiler

import "testing"

func TestProbeTodoDiags(t *testing.T) {
	src := `main() => {
  $list := ["Item 1"]
  $inputValue := ""
  <input
    @change={(e) => $inputValue = e.target.value}
    @keydown={(e) => ((e.key == "Enter") ? {
      $list.push(e.target.value)
      e.target.value = ""
    })
  } />
}`
	_, diags := AnalyzeSource("/tmp/probe.rn", src)
	for _, d := range diags {
		t.Logf("L%d C%d: %s", d.Pos.Line, d.Pos.Column, d.Message)
	}
	if len(diags) == 0 {
		t.Log("OK (no diags)")
	}
}
