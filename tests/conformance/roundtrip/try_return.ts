// expect: 1
// expect: 2
function choose(flag: boolean): number {
  try {
    return flag ? 1 : 2
  } catch (err) {
    return 0
  }
}

function main(): void {
  console.log(choose(false))
  console.log(choose(true))
}
