# 快速开始

Rune 以 Go module 方式开发。从仓库根目录运行下面的命令即可直接使用工具链：

```sh
go run ./cmd/rune check examples/fib.rn
go run ./cmd/rune fmt examples/fib.rn
go run ./cmd/rune run examples/fib.rn
go run ./cmd/rune build -o /tmp/rune-fib examples/fib.rn
go run ./cmd/rune ts examples/counter.rn
go run ./cmd/rune repl
go run ./cmd/rune lsp
```

开发脚本封装了常用入口：

```sh
scripts/dev.sh
scripts/dev.sh --shell
scripts/dev.sh --vscode
```

为编辑器集成构建本地 CLI：

```sh
go build -o .bin/rune ./cmd/rune
```

## 最小程序

```rune
main() => {
  @io.println("Hello, Rune")
}
```

`main` 是编译或运行程序时的入口。它必须返回 `Void`；如果没有显式声明返回
类型，Rune 会把 `main` 当作 `Void` 函数检查。

## 运行测试

Rune 源文件可以包含测试声明：

```rune
? "arithmetic" {
  @assert.eq(1 + 2, 3)
}
```

当前 Go 测试套件会驱动 Rune 测试器：

```sh
go test ./...
```

## 文档站点

这份文档是一个 VitePress 站点，所有文件都放在 `docs/` 内。

```sh
cd docs
npm install
npm run dev
npm run build
```
