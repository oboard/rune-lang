// expect: 3
// expect: base:1
// expect: 1
// expect: 2
// expect: child+base
// expect: base
// expect: child
// expect: base-kind
// expect: child-kind
class Base {
  x: number
  constructor(x: number) {
    this.x = x
  }
  value(): number {
    return this.x
  }
  label(): string {
    return "base:" + this.x
  }
  describe(): string {
    return "base"
  }
  static kind(): string {
    return "base"
  }
  static readonly KIND: string = "base-kind"
}

class Child extends Base {
  y: number
  constructor(x: number, y: number) {
    super(x)
    this.y = y
  }
  value(): number {
    return this.x + this.y
  }
  describe(): string {
    return "child+" + super.describe()
  }
  static tag(): string {
    return "child"
  }
  static readonly KIND: string = "child-kind"
}

function main(): void {
  const child = new Child(1, 2)
  console.log(child.value())
  console.log(child.label())
  console.log(child.x)
  console.log(child.y)
  console.log(child.describe())
  console.log(Child.kind())
  console.log(Child.tag())
  console.log(Base.KIND)
  console.log(Child.KIND)
}
