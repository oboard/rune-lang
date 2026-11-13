// expect: 10
// expect: 3
function total(limit: number): number {
  let sum = 0
  let value = 1
  while (value <= limit) {
    sum = sum + value
    value = value + 1
  }
  return sum
}

function main(): void {
  console.log(total(4))
  let count = 0
  while (count < 3) {
    count = count + 1
  }
  console.log(count)
}
