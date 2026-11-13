package tsfront

import "testing"

func TestProfileDoesNotRejectIdentifierNameProperties(t *testing.T) {
	source := `const o = { eval: 1, arguments: 2 };
console.log(o.eval);
console.log(o.arguments);
`
	_, diags := ConvertSource("identifiers.ts", source)
	for _, diag := range diags {
		if diag.Code != "" {
			t.Fatalf("unexpected profile diagnostic %s: %s", diag.Code, diag.Message)
		}
	}
}

func TestProfileReportsStableCodes(t *testing.T) {
	cases := []struct {
		name   string
		source string
		code   string
	}{
		{"eval", `eval("1");`, CodeDynamicEval},
		{"function constructor", `const f = Function("return 1");`, CodeDynamicFunction},
		{"proxy", `const p = new Proxy({}, {});`, CodeDynamicProxy},
		{"dynamic import", `import("./m.ts");`, CodeDynamicImport},
		{"arguments", `function f() { return arguments.length; }`, CodeDynamicArguments},
		{"global reflection", `globalThis["name"];`, CodeDynamicGlobalReflection},
		{"debugger", `debugger;`, CodeDynamicDebugger},
		{"require", `require("fs");`, CodeDynamicRequire},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, diags := ConvertSource("case.ts", tc.source)
			for _, diag := range diags {
				if diag.Code == tc.code {
					return
				}
			}
			t.Fatalf("missing %s in diagnostics: %+v", tc.code, diags)
		})
	}
}
