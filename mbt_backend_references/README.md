需要实现 mbt 后端，那么就要先把代码转换为 mbt 代码。对于 init 和 main 函数，特别的转换为mbt顶层函数，mbt 的top level fn是强制生命签名类型的，这是 mbt的 by design，必要的时候需要根据类型推导写入顶层函数类型签名。

```rune
main() => {
    @io.println("Hello")
}
```

```mbt
fn main {
    println("Hello")
}
```



其他函数啧转换为正常的 mbt 函数



```rune
fib(n) => {
  0 => 0
  1 => 1
  _ => fib(n - 1) + fib(n - 2)
}
```

```mbt
fn fib(n: Int) -> Int {
    match n {
        0 => 0
        1 => 1
        _ => fib(n - 1) + fib(n - 2)
    }
}
```

剩余语法细节参考 mbt_backend_references/moonbitlang


还需要 mbt完成构建系统的调用


```bash
> moon build -h
Build the current package

Usage: moon build [OPTIONS] [PATH]...

Arguments:
  [PATH]...  Paths to the packages that should be built

Options:
  -g, --debug                      Emit debug information
      --release                    Compile in release mode
      --strip                      Enable stripping debug information
      --no-strip                   Disable stripping debug information
      --target <TARGET>            Select output target [possible values: wasm, wasm-gc, js, native, llvm, all]
      --enable-coverage            Enable coverage instrumentation
      --sort-input                 Sort input files
      --output-wat                 Output WAT instead of WASM
  -d, --deny-warn                  Treat all warnings as errors
      --no-render                  Don't render diagnostics (in raw human-readable format)
      --output-json                Output diagnostics in JSON format
      --warn-list <WARN_LIST>      Warn list config
  -j, --jobs <JOBS>                Set the max number of jobs to run in parallel
      --render-no-loc <MIN_LEVEL>  Render no-location diagnostics starting from a certain level [default: error] [possible values: info, warn, error]
      --diagnostic-limit <N>       Limit the number of rendered diagnostics
  -h, --help                       Print help

Manifest Options:
      --frozen  Do not sync dependencies, assuming local dependencies are up-to-date
  -w, --watch   Monitor the file system and automatically build artifacts

Common Options:
      --target-dir <TARGET_DIR>  The target directory. Defaults to `<project-root>/_build`
  -q, --quiet                    Suppress output
  -v, --verbose                  Increase verbosity
      --trace                    Trace the execution of the program
      --dry-run                  Do not actually run the command

```


moonbitlang 只能编译完整的 mod 工程，所以需要给 output.mbt 附加 mod, pkg 模板
需要进入这三个文件所在的文件夹运行，moon build --target {wasm, wasm-gc, js, native, llvm, all} --release
所以 mbt 程序至少要三个文件才能编译，编译之后产物会放在那个目录下的 

`--target wasm --release` => _build/wasm-gc/release/build/hello.wasm
`--target js --debug` => _build/js/debug/build/hello.js
以此类推

最终需要达到以下效果
```bash
rune run examples/fib.rn --backend mbt

rune mbt examples/fib.rn

rune build examples/fib.rn --backend mbt --target js
rune build examples/fib.rn --backend mbt --target wasm
```
build 当中 --target 是专门给 mbt backend 准备的


所有的 rune stub 都要在 mbt 实现，尽量翻译成 ~/.moon/lib/core/ 当中的用法