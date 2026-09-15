package parser

import "testing"

func TestParseMultilineXMLEventAttributes(t *testing.T) {
	_, errs := Parse(`
main() => {
  $list := [{ name: "one" }]
  <div>
    {$list.map((item, index: Int) => (
      <input value={item.name} @change={(e) => { $list.set(index, {..item, name: e.target?.value ?? ""}) }} />
    ))}
    <input
      type={"text"}
      @change={(e) => e.target?.value ?? ""}
      @keydown={(e) => e.key} />
    <button @click={addTodo()}>Add Item</button>
  </div>
}
`)
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
}
