// code: RUNE-DYN-005
const shape = { value: 1 };
Object.setPrototypeOf(shape, null);
console.log(shape.value);
