// While we would prefer to define a class Calculator that wraps an externref,
// it is not currently possible to use externref anywhere except the stack. So
// the best we can do is accept the externref as a parameter to calls to
// floating functions.
// https://github.com/AssemblyScript/assemblyscript/issues/1859
export type Calculator = externref;

// @ts-ignore
@external("lib", "Calculator_Push")
export declare function Push(calculator: Calculator, value: u32): void;

// @ts-ignore
@external("lib", "Calculator_Pop")
export declare function Pop(calculator: Calculator): u32;

// @ts-ignore
@external("lib", "Calculator_Add")
export declare function Add(calculator: Calculator): void;

// @ts-ignore
@external("lib", "Calculator_Subtract")
export declare function Subtract(calculator: Calculator): void;

// @ts-ignore
@external("lib", "Calculator_Multiply")
export declare function Multiply(calculator: Calculator): void;

// @ts-ignore
@external("lib", "Calculator_Divide")
export declare function Divide(calculator: Calculator): void;
