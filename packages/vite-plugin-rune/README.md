# @rune-lang/vite-plugin

Import Rune modules from Vite projects.

```ts
import { add, type User } from "./math.rn";
```

```ts
import { defineConfig } from "vite";
import { rune } from "@rune-lang/vite-plugin";

export default defineConfig({
  plugins: [
    rune({
      runeRoot: process.env.RUNE_ROOT
    })
  ]
});
```

The plugin compiles `.rn` files with `rune ts` and writes sibling
`*.rn.d.ts` files with `rune dts` so TypeScript can type-check imports.
