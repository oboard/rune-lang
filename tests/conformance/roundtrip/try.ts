// expect: boom
// expect: finally
function main(): void {
  try {
    throw "boom"
  } catch (err) {
    console.log(err)
  } finally {
    console.log("finally")
  }
}
