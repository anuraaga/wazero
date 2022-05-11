// The entry file of your WebAssembly module.

import "wasi";

import * as calculator from "./calculator";

export function fahrenheitToCelsius(calc: externref, val: u32): u32 {
  calculator.Push(calc, val);
  calculator.Push(calc, 32);
  calculator.Subtract(calc);
  calculator.Push(calc, 5);
  calculator.Multiply(calc);
  calculator.Push(calc, 9);
  calculator.Divide(calc);
  return calculator.Pop(calc);
}

export function celsiusToFahrenheit(calc: externref, val: u32): u32 {
  calculator.Push(calc, val);
  calculator.Push(calc, 9);
  calculator.Multiply(calc);
  calculator.Push(calc, 5);
  calculator.Divide(calc);
  calculator.Push(calc, 32);
  calculator.Add(calc);
  return calculator.Pop(calc);
}
