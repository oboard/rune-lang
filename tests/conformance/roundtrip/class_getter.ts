// expect: 32
class Temp {
  celsius: number
  constructor(c: number) {
    this.celsius = c
  }
  get fahrenheit(): number {
    return this.celsius * 9 + 32
  }
}

function main(): void {
  const temp = new Temp(0)
  console.log(temp.fahrenheit)
}
