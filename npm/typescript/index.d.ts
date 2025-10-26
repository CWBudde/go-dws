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
    setFileSystem(fs: VirtualFileSystem): void;
    dispose(): void;
}

export interface VirtualFileSystem {
    readFile(path: string): Promise<Uint8Array>;
    writeFile(path: string, data: Uint8Array): Promise<void>;
    listDir(path: string): Promise<string[]>;
    delete(path: string): Promise<void>;
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
