// expect: 6
function sum(limit: number): number {
  let total = 0
  for (let i = 0; i < limit; i++) {
    total = total + i
  }
  return total
}

function main(): void {
  console.log(sum(4))
}
