import * as __import0 from "lib";
async function instantiate(module, imports = {}) {
  const __module0 = imports.lib;
  const adaptedImports = {
    lib: Object.assign(Object.create(__module0), {
      Calculator_Push(calculator, value) {
        // assembly/calculator/Push(externref, u32) => void
        value = value >>> 0;
        __module0.Calculator_Push(calculator, value);
      },
    }),
  };
  const { exports } = await WebAssembly.instantiate(module, adaptedImports);
  const memory = exports.memory || imports.env.memory;
  const adaptedExports = Object.setPrototypeOf({
    fahrenheitToCelsius(calc, val) {
      // assembly/index/fahrenheitToCelsius(externref, u32) => u32
      return exports.fahrenheitToCelsius(calc, val) >>> 0;
    },
    celsiusToFahrenheit(calc, val) {
      // assembly/index/celsiusToFahrenheit(externref, u32) => u32
      return exports.celsiusToFahrenheit(calc, val) >>> 0;
    },
  }, exports);
  return adaptedExports;
}
export const {
  fahrenheitToCelsius,
  celsiusToFahrenheit
} = await (async url => instantiate(
  await (
    globalThis.fetch && globalThis.WebAssembly.compileStreaming
      ? globalThis.WebAssembly.compileStreaming(globalThis.fetch(url))
      : globalThis.WebAssembly.compile(await (await import("node:fs/promises")).readFile(url))
  ), {
    lib: __maybeDefault(__import0),
  }
))(new URL("temperature.wasm", import.meta.url));
function __maybeDefault(module) {
  return typeof module.default === "object" && Object.keys(module).length == 1
    ? module.default
    : module;
}
