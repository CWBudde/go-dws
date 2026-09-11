# @cwbudde/dwscript

Official npm package for the DWScript interpreter compiled to WebAssembly. It bundles the Go → WASM build (`dwscript.wasm` + `wasm_exec.js`) together with a compact loader that works in both browsers and Node.js runtimes.

## Installation

```bash
npm install @cwbudde/dwscript
# or
pnpm add @cwbudde/dwscript
yarn add @cwbudde/dwscript
```

The package ships the following public files:

```
index.js          # ESM entry (exports helpers + default async factory)
index.cjs         # CommonJS bridge that proxies to the ESM loader
loader.js         # Runtime/bootstrap helper (ESM)
dwscript.wasm     # Prebuilt WebAssembly binary (rebuilt at pack time, see below)
wasm_exec.js      # Official Go WASM runtime support file
typescript/       # TypeScript declarations
examples/         # Usage samples (Node.js, React, Vue, vanilla)
```

## Quick Start (Browser)

```javascript
import { createDWScript } from '@cwbudde/dwscript';

async function run() {
  // Loads wasm_exec.js, dwscript.wasm, runs Go runtime, and calls DWScript.init()
  const dws = await createDWScript({
    initOptions: {
      onOutput: (text) => console.log(text),
      onError: (error) => console.error(error),
    },
  });

  const result = dws.eval('PrintLn("Hello from DWScript!");');
  console.log(result.output); // => Hello from DWScript!
}

run();
```

Add the wasm asset to your bundler's asset pipeline. Most modern bundlers understand `new URL('./dwscript.wasm', import.meta.url)` automatically because that is what the loader uses internally. If you host the WASM file elsewhere, pass a `runtime.wasmURL` option (see below).

## Quick Start (Node.js ≥ 18)

```javascript
import createDWScript from '@cwbudde/dwscript';

const dws = await createDWScript();
const program = dws.compile('PrintLn("Node + DWScript");');
const result = dws.run(program);
console.log(result.output);
```

Node.js has a built-in `fetch`, so no extra polyfills are required. The loader automatically reads the local `dwscript.wasm` via `fs/promises` when running on Node.

## CommonJS Usage

```javascript
const { createDWScript } = require('@cwbudde/dwscript');

(async () => {
  const dws = await createDWScript();
  const { output } = dws.eval('PrintLn("CommonJS ready");');
  console.log(output);
})();
```

> **Note**: The CommonJS bridge wraps the ESM loader internally, so helper functions return Promises.

## API Surface

The Go runtime exposes a `DWScript` class that mirrors the JavaScript API described in `docs/wasm/API.md`. The npm loader keeps that surface intact and adds a few helper utilities.

### Helper exports

| Export | Type | Description |
| ------ | ---- | ----------- |
| `createDWScript(options?)` | `Promise<DWScript>` | Loads the runtime (if needed), creates a new `DWScript` instance, and calls `init()` unless `autoInit=false` |
| `ensureRuntimeReady(options?)` | `Promise<{ go, instance }>` | Boots the Go WASM runtime. Useful if you want to defer `new DWScript()` until later |
| `getDWScriptClass()` | `DWScript` constructor | Returns the global constructor registered by the Go binary |
| `isRuntimeInitialized()` | `boolean` | Whether the runtime already booted |
| `resetRuntimeForTesting()` | `void` | Clears the cached runtime handle (primarily for automated tests) |
| `version` | `string` | Package version (`0.1.0`) |

### DWScript instance

Once created, the `DWScript` instance supports the same methods documented in the Go project:

- `init(options?: { onOutput, onError, onInput, fs })`
- `compile(source: string)` → `{ id, success }`
- `run(program)` → `{ success, output, executionTime, error? }`
- `eval(source: string)` → same as `run`
- `on(event, callback)` for `output`, `error`, `input`
- `setFileSystem(fs)` installs a host filesystem (synchronous methods only; see
  `docs/wasm/API.md`), returns `null` or an `ArgumentError`
- `version()` returns `{ version, build: 'wasm', platform: 'javascript' }`
- `dispose()` cleans up callbacks and Go handles

TypeScript definitions live in `typescript/index.d.ts` and describe the helper options, runtime handles, and VFS interface.

## Runtime Options

Both `createDWScript` and `ensureRuntimeReady` accept a `runtime` object with advanced controls:

```javascript
await ensureRuntimeReady({
  wasmURL: 'https://cdn.example.com/dwscript.wasm',
  readyDelay: 100, // wait longer before resolving to let Go finish booting
  fetchOptions: { credentials: 'include' },
  wasmBinary: await fetch('/custom/dws.wasm').then((res) => res.arrayBuffer()),
});
```

- `wasmURL`: Absolute or relative URL for the WASM binary. Defaults to the bundled `dwscript.wasm`.
- `wasmBinary`: Preloaded bytes or `WebAssembly.Module`. Skips network/file I/O entirely.
- `fetchOptions`: Extra `fetch` init data when downloading the WASM file.
- `readyDelay`: Milliseconds to wait after calling `go.run()`. Defaults to `50` to give Go time to register `DWScript` globally.
- `instantiate`: Optional hook to fully customize how the WebAssembly module is instantiated.

Once the runtime is initialized, subsequent calls ignore additional runtime options and reuse the cached Go instance.

## Examples

Ready-to-run snippets live in `examples/`:

- `examples/node.js` – Basic Node.js script reading DWScript output
- `examples/react.jsx` – React hook that compiles + runs code from a textarea
- `examples/vue.js` – Vue 3 component with two-way binding
- `examples/vanilla.html` – Minimal HTML page that loads the package via a bundler entry

Feel free to copy these into your project or adapt them as integration tests.

## Bundler Tips

1. **WASM asset** – Ensure your bundler copies `dwscript.wasm`. For Vite/Rollup you can add:
   ```javascript
   // vite.config.js
   export default {
     assetsInclude: ['**/*.wasm'],
   };
   ```
2. **Top-level await** – The loader avoids TLA so it works in older bundlers. Use the async helpers instead of relying on side effects.
3. **Tree shaking** – Only `wasm_exec.js` is marked as having side effects. The rest of the helpers can be tree-shaken if unused.
4. **Workers** – Because the loader relies on Go's `wasm_exec.js`, you can use it inside Web Workers by bundling the same files and calling `createDWScript()` in the worker context.

## Publishing & CI

This package is meant to be published via the `npm-publish.yml` GitHub Actions workflow (see repo root). The workflow builds the WASM binary, copies it into `npm/`, and runs `npm publish --provenance` when a release tag is pushed.

### The bundled `dwscript.wasm` is a build output

`npm/dwscript.wasm` is tracked for convenience, but it is generated from the Go sources in this repository and is authoritative only for the commit that produced it. Any commit touching the interpreter or the WASM bridge leaves it behind.

So that nobody ships a stale binary, the package has a `prepare` script (`scripts/prepare-wasm.mjs`) that rebuilds `dwscript.wasm` and `wasm_exec.js` from source whenever the repository is reachable next to the package. npm runs it for exactly the cases that bypass the release workflow:

- `npm pack` and `npm publish` from a checkout
- installing the package as a Git dependency

Those runs need Go 1.24+ on `PATH`; the script fails with an explanatory message rather than packaging the tracked binary. Installs from the npm registry never run `prepare`, so ordinary consumers are unaffected, and the script no-ops when the Go sources are absent. Set `DWSCRIPT_SKIP_WASM_BUILD=1` to force it to keep the tracked binary.

## License

MIT © Christian Budde
