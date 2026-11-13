// code: RUNE-DYN-009
function call(value: unknown): void {
  value.run()
}
call({ run: 1 })
