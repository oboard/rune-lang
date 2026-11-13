// expect: string:Ada
// expect: number
function describe(value: string | number): string {
  return typeof value === "string" ? "string:" + value : "number"
}

function main(): void {
  console.log(describe("Ada"))
  console.log(describe(7))
}
