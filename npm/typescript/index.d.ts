export type DWScriptEvent = 'output' | 'error' | 'input';

export interface RuntimeError extends Error {
    type: string;
    source?: string;
    line?: number;
    column?: number;
    executionTime?: number;
}

export interface Program {
    id: number;
    success: boolean;
}

export interface Result {
    success: boolean;
    output: string;
    executionTime: number;
    error?: RuntimeError;
}

export interface DWScriptInitOptions {
    onOutput?: (text: string) => void;
    onError?: (error: RuntimeError) => void;
    onInput?: (prompt: string) => string | Promise<string>;
    fs?: VirtualFileSystem;
}

export interface DWScriptInstance {
    init(options?: DWScriptInitOptions): Promise<void>;
    compile(source: string): Program;
    run(program: Program): Result;
    eval(source: string): Result;
    on(event: 'output', callback: (text: string) => void): void;
    on(event: 'error', callback: (error: RuntimeError) => void): void;
    on(event: 'input', callback: (prompt: string) => string | Promise<string>): void;
    version(): { version: string; build: string; platform: string };
    /** Returns null on success, or an Error (type 'ArgumentError') if the object is invalid. */
    setFileSystem(fs: VirtualFileSystem | null): null | Error;
    dispose(): void;
}

export interface VirtualFileSystemEntry {
    name: string;
    size?: number;
    isDir?: boolean;
    /** Modification time in milliseconds since the epoch. */
    modTime?: number;
}

/**
 * Host-supplied filesystem. Every method must return synchronously: the Go
 * side is a synchronous interface and awaiting a Promise would deadlock the
 * WASM event loop. A method returning a thenable fails with an explicit error.
 * Report failures by throwing.
 */
export interface VirtualFileSystem {
    readFile(path: string): Uint8Array | string;
    writeFile(path: string, data: Uint8Array): void;
    listDir(path: string): Array<string | VirtualFileSystemEntry>;
    delete(path: string): void;
    exists(path: string): boolean;
}

export interface RuntimeOptions {
    wasmURL?: string | URL;
    wasmBinary?: ArrayBuffer | ArrayBufferView | WebAssembly.Module;
    fetchOptions?: RequestInit;
    readyDelay?: number;
    instantiate?: (context: InstantiateContext) => Promise<WebAssembly.WebAssemblyInstantiatedSource> | WebAssembly.WebAssemblyInstantiatedSource;
}

export interface InstantiateContext {
    go: any;
    wasmBinary?: ArrayBuffer | ArrayBufferView | WebAssembly.Module;
    wasmURL?: URL;
    fetchOptions?: RequestInit;
}

export interface RuntimeHandle {
    go: any;
    instance: WebAssembly.Instance;
    runPromise?: Promise<void>;
}

export interface CreateOptions {
    runtime?: RuntimeOptions;
    autoInit?: boolean;
    initOptions?: DWScriptInitOptions;
}

export interface DWScriptConstructor {
    new (): DWScriptInstance;
}

export function createDWScript(options?: CreateOptions): Promise<DWScriptInstance>;
export function ensureRuntimeReady(options?: RuntimeOptions): Promise<RuntimeHandle>;
export function getDWScriptClass(): DWScriptConstructor;
export function isRuntimeInitialized(): boolean;
export function resetRuntimeForTesting(): void;
export const version: string;

export default createDWScript;
