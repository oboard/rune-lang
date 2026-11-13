// expect: 3
// expect: 1
function main(): void {
  let i = 0
  do {
    i = i + 1
  } while (i < 3)
  console.log(i)

  let n = 0
  do {
    n = n + 1
  } while (false)
  console.log(n)
}
