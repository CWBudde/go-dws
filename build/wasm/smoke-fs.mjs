// Node smoke test for the custom filesystem bridge (PLAN.md §3.4).
//
// Loads a freshly built dwscript.wasm and exercises the JavaScript-facing
// contract of setFileSystem() / init({fs}): a well-formed synchronous
// filesystem object is accepted, an incomplete one is rejected with a
// structured ArgumentError, and null resets to the built-in virtual
// filesystem.
//
// Usage: node build/wasm/smoke-fs.mjs <path-to-dwscript.wasm> <path-to-wasm_exec.js>

import { readFile } from 'node:fs/promises';
import { pathToFileURL } from 'node:url';

const [, , wasmPath, wasmExecPath] = process.argv;
if (!wasmPath || !wasmExecPath) {
    console.error('usage: node smoke-fs.mjs <dwscript.wasm> <wasm_exec.js>');
    process.exit(2);
}

let failures = 0;
function check(name, condition, detail) {
    if (condition) {
        console.log(`  ok   ${name}`);
    } else {
        failures += 1;
        console.error(`  FAIL ${name}${detail ? `: ${detail}` : ''}`);
    }
}

// wasm_exec.js is a classic script that assigns globalThis.Go.
await import(pathToFileURL(wasmExecPath).href);

const go = new globalThis.Go();
const { instance } = await WebAssembly.instantiate(await readFile(wasmPath), go.importObject);
go.run(instance); // resolves only when the module exits; it parks on a channel
await new Promise((resolve) => setTimeout(resolve, 50));

check('DWScript constructor is exported', typeof globalThis.DWScript === 'function');

// A complete, synchronous filesystem backed by a plain Map.
function makeFS(store = new Map()) {
    const enc = new TextEncoder();
    return {
        store,
        readFile(path) {
            if (!store.has(path)) throw new Error(`file not found: ${path}`);
            return enc.encode(store.get(path));
        },
        writeFile(path, data) {
            store.set(path, new TextDecoder().decode(data));
        },
        listDir(path) {
            return [...store.keys()]
                .filter((p) => p.startsWith(path))
                .map((p) => ({ name: p.slice(path.length).replace(/^\//, ''), size: store.get(p).length }));
        },
        delete(path) {
            store.delete(path);
        },
        exists(path) {
            return store.has(path);
        },
    };
}

const dws = new globalThis.DWScript();

// 1. A valid filesystem installs and returns null.
const good = makeFS(new Map([['/hello.txt', 'world']]));
check('setFileSystem(valid) returns null', dws.setFileSystem(good) === null);

// 2. An incomplete object is rejected up front with a structured error.
const bad = dws.setFileSystem({ readFile: () => new Uint8Array() });
check('setFileSystem(incomplete) returns an Error', bad instanceof Error, String(bad));
check(
    'setFileSystem(incomplete) reports ArgumentError',
    bad instanceof Error && bad.type === 'ArgumentError',
    bad instanceof Error ? bad.type : String(bad),
);
check(
    'setFileSystem(incomplete) names the missing methods',
    bad instanceof Error && ['writeFile', 'listDir', 'delete', 'exists'].every((m) => bad.message.includes(m)),
    bad instanceof Error ? bad.message : String(bad),
);

// 3. A non-object is rejected.
const notAnObject = dws.setFileSystem('nope');
check('setFileSystem("nope") returns an Error', notAnObject instanceof Error, String(notAnObject));

// 3b. An object whose property reads throw is rejected the same way, instead
// of tearing down the WASM instance or yielding a silent null.
const throwing = dws.setFileSystem(
    new Proxy(
        {},
        {
            get() {
                throw new Error('accessor exploded');
            },
        },
    ),
);
check('setFileSystem(throwing proxy) returns an Error', throwing instanceof Error, String(throwing));
check(
    'setFileSystem(throwing proxy) reports ArgumentError',
    throwing instanceof Error && throwing.type === 'ArgumentError',
    throwing instanceof Error ? throwing.type : String(throwing),
);
check(
    'setFileSystem(throwing proxy) surfaces the host message',
    throwing instanceof Error && throwing.message.includes('accessor exploded'),
    throwing instanceof Error ? throwing.message : String(throwing),
);

// 4. null resets without error.
check('setFileSystem(null) returns null', dws.setFileSystem(null) === null);

// 5. init({fs}) accepts a valid filesystem and still resolves.
await dws.init({ fs: makeFS(), onOutput: () => {} });
check('init({fs: valid}) resolves', true);

// 6. init({fs}) rejects an invalid filesystem instead of warning.
let rejected = false;
try {
    await dws.init({ fs: { readFile: () => null } });
} catch (err) {
    rejected = err instanceof Error && err.type === 'ArgumentError';
}
check('init({fs: invalid}) rejects with ArgumentError', rejected);

// 7. The interpreter still works after installing a filesystem.
dws.setFileSystem(makeFS());
const result = dws.eval("PrintLn('smoke');");
check('eval still runs', result.success === true, JSON.stringify(result.error?.message));
check('eval produced output', String(result.output).includes('smoke'), String(result.output));

// 8. A script reads and writes through the installed filesystem. This is the
//    check that distinguishes a filesystem the host can install from one the
//    interpreter actually consults.
const scripted = makeFS(new Map([['/in.txt', 'from the host']]));
dws.setFileSystem(scripted);
const roundTrip = dws.eval(
    "PrintLn(LoadTextFromFile('/in.txt')); SaveTextToFile('/out.txt', 'from the script');",
);
check('script eval succeeds', roundTrip.success === true, JSON.stringify(roundTrip.error?.message));
check(
    'LoadTextFromFile reads through the installed filesystem',
    String(roundTrip.output).includes('from the host'),
    String(roundTrip.output),
);
check(
    'SaveTextToFile writes through the installed filesystem',
    scripted.store.get('/out.txt') === 'from the script',
    String(scripted.store.get('/out.txt')),
);

dws.dispose();

if (failures > 0) {
    console.error(`\n${failures} check(s) failed`);
    process.exit(1);
}
console.log('\nall filesystem smoke checks passed');
process.exit(0);
