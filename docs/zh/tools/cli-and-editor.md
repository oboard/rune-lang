# CLI 与编辑器

Rune CLI 位于 `cmd/rune`。

```text
rune check <path>      解析并类型检查文件或目录
rune fmt <path>        格式化文件或目录（别名：format）
rune run <path>        运行文件，或只有一个 main 的目录
rune build <file.rn>   编译 Rune 程序为可执行文件
rune ts <file.rn>      编译 Rune 程序为 TypeScript
rune repl              启动 Rune REPL
rune lsp               启动 Rune language server
```

开发时可以直接通过 `go run` 调用：

```sh
go run ./cmd/rune check examples/fib.rn
go run ./cmd/rune check core
go run ./cmd/rune run examples/fib.rn
go run ./cmd/rune ts examples/counter.rn
```

也可以构建二进制：

```sh
go build -o .bin/rune ./cmd/rune
```

## 格式化

```sh
go run ./cmd/rune fmt examples/fib.rn
```

formatter 会规范化声明、对象字段、数组、代码块、match 分支、XML 表达式和注释。

## REPL

```sh
go run ./cmd/rune repl
```

REPL 通过解释器路径求值表达式和声明，不需要 `main` 函数。

## Language Server

```sh
go run ./cmd/rune lsp
```

`rune lsp` 默认通过 stdio 服务，也接受 `--stdio`。

`vscode-rune/` 中的 VSCode 扩展提供：

- TextMate 高亮。
- snippets。
- diagnostics。
- hover。
- completion。
- go to definition。
- rename。
- document symbols。
- formatting。
- 函数和 lambda 推断类型的 inlay hints。
- 针对 `main` 的 run/debug code lens。

本地 VSCode 开发：

```sh
scripts/dev.sh --vscode
```
