import './wasm_exec.js';

const DEFAULT_READY_DELAY = 50;
const RUNTIME_HANDLE_KEY = '__dwscript_runtime__';
const PACKAGE_VERSION = '0.1.0';

let runtimePromise = null;

export const version = PACKAGE_VERSION;

export async function ensureRuntimeReady(options = {}) {
    if (globalThis[RUNTIME_HANDLE_KEY]) {
        warnIgnoredOptions(options);
        return globalThis[RUNTIME_HANDLE_KEY];
    }

    if (!runtimePromise) {
        runtimePromise = bootstrapRuntime(options);
    }

    return runtimePromise;
}

export async function createDWScript(options = {}) {
    const runtimeOptions = options.runtime;
    await ensureRuntimeReady(runtimeOptions);

    const DWScriptClass = getDWScriptClass();
    const instance = new DWScriptClass();

    if (options.autoInit === undefined || options.autoInit === true) {
        await instance.init(options.initOptions);
    }

    return instance;
}

export function getDWScriptClass() {
    if (typeof globalThis.DWScript !== 'function') {
        throw new Error('DWScript runtime has not been initialized. Call ensureRuntimeReady() first.');
    }
    return globalThis.DWScript;
}

export function isRuntimeInitialized() {
    return Boolean(globalThis[RUNTIME_HANDLE_KEY]);
}

export function resetRuntimeForTesting() {
    delete globalThis[RUNTIME_HANDLE_KEY];
    runtimePromise = null;
}

async function bootstrapRuntime(options) {
    if (globalThis[RUNTIME_HANDLE_KEY]) {
        return globalThis[RUNTIME_HANDLE_KEY];
    }

    const GoConstructor = globalThis.Go;
    if (typeof GoConstructor !== 'function') {
        throw new Error('Go WebAssembly runtime (wasm_exec.js) is not available.');
    }

    const go = new GoConstructor();
    const wasmBinary = options?.wasmBinary;
    const wasmURL = resolveWasmURL(options?.wasmURL);
    const instantiate = options?.instantiate ?? instantiateDefault;

    const result = await instantiate({
        go,
        wasmBinary,
        wasmURL,
        fetchOptions: options?.fetchOptions ?? {},
    });

    const instance = result.instance ?? result;

    const runPromise = go.run(instance);
    runPromise.catch((error) => {
        console.error('DWScript runtime exited with an error', error);
    });

    const readyDelay = options?.readyDelay ?? DEFAULT_READY_DELAY;
    if (readyDelay > 0) {
        await delay(readyDelay);
    }

    const handle = { go, instance, runPromise };
    globalThis[RUNTIME_HANDLE_KEY] = handle;
    return handle;
}

async function instantiateDefault({ go, wasmBinary, wasmURL, fetchOptions }) {
    if (wasmBinary) {
        return instantiateFromBinary(go, wasmBinary);
    }

    if (isNodeEnvironment()) {
        return instantiateNode(go, wasmURL);
    }

    return instantiateBrowser(go, wasmURL, fetchOptions);
}

async function instantiateFromBinary(go, source) {
    if (source instanceof WebAssembly.Module) {
        return { instance: new WebAssembly.Instance(source, go.importObject) };
    }

    let buffer;
    if (source instanceof ArrayBuffer) {
        buffer = new Uint8Array(source);
    } else if (ArrayBuffer.isView(source)) {
        buffer = new Uint8Array(source.buffer);
    } else {
        throw new Error('Unsupported wasmBinary type. Expected ArrayBuffer, TypedArray, or WebAssembly.Module.');
    }

    return WebAssembly.instantiate(buffer, go.importObject);
}

async function instantiateBrowser(go, wasmURL, fetchOptions = {}) {
    const resolvedURL = wasmURL ?? new URL('./dwscript.wasm', import.meta.url);

    if (WebAssembly.instantiateStreaming && isHttpProtocol(resolvedURL)) {
        try {
            return await WebAssembly.instantiateStreaming(fetch(resolvedURL, fetchOptions), go.importObject);
        } catch (error) {
            console.warn('Falling back to arrayBuffer instantiation due to streaming error', error);
        }
    }

    const response = await fetch(resolvedURL, fetchOptions);
    if (!response.ok) {
        throw new Error(`Failed to fetch DWScript WASM: ${response.status} ${response.statusText}`);
    }
    const bytes = await response.arrayBuffer();
    return WebAssembly.instantiate(bytes, go.importObject);
}

async function instantiateNode(go, wasmURL) {
    const resolvedURL = wasmURL ?? new URL('./dwscript.wasm', import.meta.url);

    if (isHttpProtocol(resolvedURL)) {
        const response = await fetch(resolvedURL);
        if (!response.ok) {
            throw new Error(`Failed to download DWScript WASM: ${response.status} ${response.statusText}`);
        }
        const bytes = await response.arrayBuffer();
        return WebAssembly.instantiate(bytes, go.importObject);
    }

    const [{ readFile }, { fileURLToPath }] = await Promise.all([
        import('node:fs/promises'),
        import('node:url'),
    ]);

    const filePath = fileURLToPath(resolvedURL);
    const buffer = await readFile(filePath);
    return WebAssembly.instantiate(buffer, go.importObject);
}

function resolveWasmURL(input) {
    if (!input) {
        return new URL('./dwscript.wasm', import.meta.url);
    }

    if (input instanceof URL) {
        return input;
    }

    if (typeof input === 'string') {
        try {
            return new URL(input, import.meta.url);
        } catch (error) {
            console.warn('Unable to resolve custom WASM URL, falling back to default', error);
            return new URL('./dwscript.wasm', import.meta.url);
        }
    }

    return new URL('./dwscript.wasm', import.meta.url);
}

function isHttpProtocol(url) {
    return url.protocol === 'https:' || url.protocol === 'http:';
}

function isNodeEnvironment() {
    return typeof process !== 'undefined' && !!process.versions?.node;
}

function delay(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

function warnIgnoredOptions(options) {
    if (!options) {
        return;
    }

    const keys = Object.keys(options);
    if (keys.length > 0) {
        console.warn('DWScript runtime already initialized. Additional runtime options are ignored:', keys);
    }
}
