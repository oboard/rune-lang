---
layout: home

hero:
  name: Rune
  text: 语言工具链
  tagline: 从紧凑的语言表面完成解析、类型检查、解释执行、编译到 Go，并生成面向 DOM 的 TypeScript。
  image:
    src: /rune-icon.svg
    alt: Rune
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/guide/getting-started
    - theme: alt
      text: 基础语法
      link: /zh/language/fundamentals

features:
  - title: 紧凑语法
    details: 函数、代码块、对象、数组、模式函数体和 match 表达式共享表达式优先的模型。
  - title: 多运行目标
    details: 可以通过解释器运行、编译到 Go，或为 DOM 风格目标生成 TypeScript。
  - title: 内置工具链
    details: 仓库内置 parser、checker、formatter、LSP、REPL、核心库 stub、编辑器集成和核心库文档。
---

## 概览

Rune 是一个用 Go 编写的表达式导向语言工具链。当前实现会解析 Rune
源码、进行类型检查、降低到共享 IR，然后可以解释执行、编译到 Go，或为
DOM 风格程序生成 TypeScript。

```rune
add(a: Int, b: Int) -> Int => a + b

main() => {
  @io.println(add(1, 2))
}
```

Rune 的语法面比较小：函数把参数映射到表达式，代码块返回最后一个表达式，
数据主要由数组和记录式对象表示，控制流由 match 表达式、模式函数体、三元
表达式和普通函数调用组成。

## 文档入口

- [快速开始](/zh/guide/getting-started) 介绍本地命令。
- [基础语法](/zh/language/fundamentals) 覆盖当前支持的完整语法。
- [核心库](/zh/language/core-library) 列出内置模块和方法。
- [GitHub: oboard/rune-lang](https://github.com/oboard/rune-lang) 是源码仓库。

English documentation is available at [English](/).
