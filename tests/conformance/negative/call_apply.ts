// code: RUNE-DYN-013
function greet(name: string): string {
  return "hello " + name;
}
console.log(greet.call(null, "Ada"));
