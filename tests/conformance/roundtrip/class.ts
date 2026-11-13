// expect: 3
// expect: 0
class Point {
  x: number
  y: number
  constructor(x: number, y: number) {
    this.x = x
    this.y = y
  }
  sum(): number {
    return this.x + this.y
  }
  static origin(): Point {
    return new Point(0, 0)
  }
}

function main(): void {
  const point = new Point(1, 2)
  console.log(point.sum())
  const origin = Point.origin()
  console.log(origin.sum())
}
