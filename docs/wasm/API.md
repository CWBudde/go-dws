# DWScript WebAssembly JavaScript API

This document describes the JavaScript API for the DWScript WebAssembly module.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
  - [DWScript Class](#dwscript-class)
  - [Methods](#methods)
  - [Events](#events)
  - [Types](#types)
- [Examples](#examples)
- [Error Handling](#error-handling)
- [Performance](#performance)
- [Browser Compatibility](#browser-compatibility)

## Installation

### Browser (CDN)

```html
<script src="wasm_exec.js"></script>
<script>
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("dwscript.wasm"), go.importObject)
        .then((result) => {
            go.run(result.instance);
            // DWScript is now available globally
        });
</script>
```

### NPM Package (Coming Soon)

```bash
npm install @meko-tech/dwscript
```

```javascript
import DWScript from '@meko-tech/dwscript';

const dws = new DWScript();
await dws.init();
```

## Quick Start

```javascript
// Load WASM module
const go = new Go();
const result = await WebAssembly.instantiateStreaming(
    fetch('dwscript.wasm'),
    go.importObject
);
go.run(result.instance);

// Create DWScript instance
const dws = new DWScript();
await dws.init();

// Run some DWScript code
const result = dws.eval(`
    var x: Integer := 42;
    PrintLn('The answer is ' + IntToStr(x));
`);

console.log(result.output); // "The answer is 42\n"
```

## API Reference

### DWScript Class

The `DWScript` class provides the main interface for running DWScript code in JavaScript.

#### Constructor

```javascript
const dws = new DWScript();
```

Creates a new DWScript instance. Each instance has its own:
- Compilation context
- Variable scope
- Compiled programs cache

### Methods

#### `init(options)`

Initialize the DWScript instance with optional configuration.

**Parameters:**
- `options` (Object, optional) - Configuration options
  - `onOutput` (Function) - Callback for program output
  - `onError` (Function) - Callback for errors
  - `onInput` (Function) - Callback for input requests
  - `fs` (Object) - Custom filesystem implementation (not yet implemented)

**Returns:** `Promise<void>`

**Example:**
```javascript
await dws.init({
    onOutput: (text) => {
        console.log('Output:', text);
    },
    onError: (error) => {
        console.error('Error:', error);
    }
});
```

#### `compile(source)`

Compile DWScript source code without executing it.

**Parameters:**
- `source` (String) - DWScript source code

**Returns:** `Program` object
- `id` (Number) - Program identifier
- `success` (Boolean) - Compilation status

**Throws:** Error object with `type` and `message` on compilation failure

**Example:**
```javascript
try {
    const program = dws.compile(`
        var x: Integer := 10;
        PrintLn(IntToStr(x));
    `);
    console.log('Compiled program ID:', program.id);
} catch (error) {
    console.error('Compilation failed:', error.message);
    if (error.type === 'CompileError') {
        console.error('Source:', error.source);
    }
}
```

#### `run(program)`

Execute a previously compiled program.

**Parameters:**
- `program` (Program) - Program object from `compile()`

**Returns:** `Result` object
- `success` (Boolean) - Execution status
- `output` (String) - Program output
- `executionTime` (Number) - Execution time in milliseconds
- `error` (Error, optional) - Error object if execution failed

**Example:**
```javascript
const program = dws.compile('PrintLn("Hello!");');
const result = dws.run(program);

if (result.success) {
    console.log('Output:', result.output);
    console.log('Took:', result.executionTime, 'ms');
} else {
    console.error('Runtime error:', result.error.message);
}
```

#### `eval(source)`

Compile and execute DWScript code in one step.

**Parameters:**
- `source` (String) - DWScript source code

**Returns:** `Result` object (same as `run()`)

**Example:**
```javascript
const result = dws.eval(`
    var x: Integer := 5;
    var y: Integer := 7;
    PrintLn('Sum: ' + IntToStr(x + y));
`);

console.log(result.output); // "Sum: 12\n"
```

#### `on(event, callback)`

Register an event listener.

**Parameters:**
- `event` (String) - Event name: `'output'`, `'error'`, or `'input'`
- `callback` (Function) - Event handler function

**Returns:** `null`

**Example:**
```javascript
dws.on('output', (text) => {
    document.getElementById('console').innerText += text;
});

dws.on('error', (error) => {
    console.error('Runtime error:', error);
});

dws.on('input', (prompt) => {
    return window.prompt(prompt);
});
```

#### `setFileSystem(fs)`

Set a custom filesystem implementation (not yet implemented).

**Parameters:**
- `fs` (Object) - Filesystem object with methods:
  - `readFile(path)` → Promise<Uint8Array>
  - `writeFile(path, data)` → Promise<void>
  - `listDir(path)` → Promise<Array<string>>
  - `delete(path)` → Promise<void>

**Returns:** `null`

**Note:** Currently logs a warning. Full implementation coming soon.

#### `version()`

Get version information.

**Returns:** Object
- `version` (String) - Version number
- `build` (String) - Build type ("wasm")
- `platform` (String) - Platform ("javascript")

**Example:**
```javascript
const version = dws.version();
console.log(`DWScript ${version.version} (${version.build})`);
```

#### `dispose()`

Clean up resources and release memory.

Call this when you're done with the DWScript instance to prevent memory leaks.

**Returns:** `null`

**Example:**
```javascript
// When done
dws.dispose();
dws = null;
```

### Events

The DWScript instance can emit the following events:

#### `'output'`

Emitted when the program produces output.

**Callback signature:** `(text: string) => void`

**Example:**
```javascript
dws.on('output', (text) => {
    console.log('Program output:', text);
});
```

#### `'error'`

Emitted when a runtime error occurs.

**Callback signature:** `(error: Error) => void`

**Example:**
```javascript
dws.on('error', (error) => {
    console.error('Error type:', error.type);
    console.error('Message:', error.message);
});
```

#### `'input'`

Emitted when the program requests input.

**Callback signature:** `(prompt: string) => string`

**Example:**
```javascript
dws.on('input', (prompt) => {
    return window.prompt(prompt) || '';
});
```

### Types

#### Program

Object representing a compiled program.

```typescript
interface Program {
    id: number;        // Unique program identifier
    success: boolean;  // Compilation status
}
```

#### Result

Object representing execution results.

```typescript
interface Result {
    success: boolean;      // Execution status
    output: string;        // Program output
    executionTime: number; // Execution time in milliseconds
    error?: Error;         // Error object if failed
}
```

#### Error

Enhanced error object with DWScript-specific information.

```typescript
interface DWScriptError extends Error {
    type: string;         // Error type (e.g., 'RuntimeError', 'CompileError')
    message: string;      // Error message
    source?: string;      // Source code (for compile errors)
    line?: number;        // Line number (when available)
    column?: number;      // Column number (when available)
    executionTime?: number; // Execution time before error
}
```

Error types:
- `InitializationError` - Failed to create DWScript instance
- `ArgumentError` - Invalid argument passed to method
- `CompileError` - Source code compilation failed
- `RuntimeError` - Error during program execution
- `ProgramError` - Invalid program reference

## Examples

### Basic Hello World

```javascript
const dws = new DWScript();
await dws.init();

const result = dws.eval('PrintLn("Hello, World!");');
console.log(result.output); // "Hello, World!\n"
```

### Compile Once, Run Multiple Times

```javascript
const program = dws.compile(`
    var i: Integer;
    for i := 1 to 5 do
        PrintLn(IntToStr(i));
`);

// Run the same program multiple times
for (let i = 0; i < 3; i++) {
    console.log('Run', i + 1);
    const result = dws.run(program);
    console.log(result.output);
}
```

### With Output Callback

```javascript
await dws.init({
    onOutput: (text) => {
        document.getElementById('output').innerText += text;
    }
});

dws.eval(`
    var x: Integer;
    for x := 1 to 10 do
        PrintLn('Number: ' + IntToStr(x));
`);
```

### Error Handling

```javascript
try {
    const result = dws.eval(`
        var x: Integer := 'invalid'; // Type error
    `);
} catch (error) {
    console.error('Error type:', error.type);
    console.error('Message:', error.message);

    if (error.type === 'CompileError') {
        console.error('Failed source:', error.source);
    }
}
```

### Event-Based Architecture

```javascript
class DWScriptRunner {
    constructor() {
        this.dws = new DWScript();
        this.setupEventHandlers();
    }

    async init() {
        await this.dws.init();
    }

    setupEventHandlers() {
        this.dws.on('output', (text) => {
            this.handleOutput(text);
        });

        this.dws.on('error', (error) => {
            this.handleError(error);
        });
    }

    handleOutput(text) {
        const outputEl = document.getElementById('output');
        outputEl.innerText += text;
    }

    handleError(error) {
        console.error('DWScript error:', error);
        this.showErrorDialog(error.message);
    }

    run(code) {
        return this.dws.eval(code);
    }

    dispose() {
        this.dws.dispose();
    }
}

// Usage
const runner = new DWScriptRunner();
await runner.init();
runner.run('PrintLn("Hello from DWScript!");');
```

## Error Handling

The DWScript API uses structured error objects for better error handling:

### Catching Compilation Errors

```javascript
try {
    const program = dws.compile(`
        var x: Integer := 'not a number'; // Type mismatch
    `);
} catch (error) {
    if (error.type === 'CompileError') {
        console.error('Compilation failed:');
        console.error('  Message:', error.message);
        console.error('  Source:', error.source);
    }
}
```

### Handling Runtime Errors

```javascript
const result = dws.eval(`
    var x: Integer := 10;
    var y: Integer := 0;
    var z: Integer := x div y; // Division by zero
`);

if (!result.success) {
    console.error('Runtime error:', result.error.message);
    console.error('Error type:', result.error.type);
    console.error('Execution time before error:', result.error.executionTime, 'ms');
}
```

### Using Error Events

```javascript
dws.on('error', (error) => {
    // Log to analytics
    analytics.logError({
        type: error.type,
        message: error.message,
        timestamp: new Date()
    });

    // Show user-friendly message
    if (error.type === 'RuntimeError') {
        showNotification('Program execution failed', 'error');
    }
});
```

## Performance

### Benchmark Results

Typical performance metrics (varies by program complexity):

- **Initialization**: < 100ms
- **Compilation**: 1-5ms for simple programs
- **Execution**: 50-80% of native Go performance

### Best Practices

1. **Compile Once, Run Many Times**
   ```javascript
   // Good: Compile once
   const program = dws.compile(code);
   for (let i = 0; i < 1000; i++) {
       dws.run(program);
   }

   // Bad: Compile every time
   for (let i = 0; i < 1000; i++) {
       dws.eval(code);
   }
   ```

2. **Dispose When Done**
   ```javascript
   const dws = new DWScript();
   await dws.init();
   // ... use dws ...
   dws.dispose(); // Clean up resources
   ```

3. **Batch Output**
   ```javascript
   let outputBuffer = '';
   dws.on('output', (text) => {
       outputBuffer += text;
   });

   // Update UI in batches
   setInterval(() => {
       if (outputBuffer) {
           document.getElementById('output').innerText += outputBuffer;
           outputBuffer = '';
       }
   }, 100);
   ```

## Browser Compatibility

### Supported Browsers

- ✅ Chrome/Chromium 57+
- ✅ Firefox 52+
- ✅ Safari 11+
- ✅ Edge 16+

### Required Features

- WebAssembly support
- ES6 Promises
- `async/await` (or use Babel for older browsers)

### Checking Support

```javascript
if (!WebAssembly) {
    alert('Your browser does not support WebAssembly');
} else {
    // Load DWScript
}
```

## See Also

- [Build Documentation](BUILD.md) - How to build the WASM module
- [Examples](../../examples/wasm/) - Full example applications
- [DWScript Language Reference](https://www.delphitools.info/dwscript/) - Language documentation

## Troubleshooting

### WASM Module Fails to Load

**Problem:** `fetch('dwscript.wasm')` fails with CORS error

**Solution:** Serve files from a web server, not `file://` protocol
```bash
python3 -m http.server 8080
# or
npx http-server
```

### Memory Leaks

**Problem:** Browser memory usage grows over time

**Solution:** Call `dispose()` when done with DWScript instances
```javascript
const dws = new DWScript();
// ... use it ...
dws.dispose(); // Important!
```

### Slow Execution

**Problem:** Programs run slower than expected

**Solution:**
1. Use `compile()` + `run()` instead of `eval()` for repeated execution
2. Check browser console for errors
3. Profile using browser DevTools

### TypeScript Support

For TypeScript projects, create a type definition file:

```typescript
// dwscript.d.ts
declare class DWScript {
    constructor();
    init(options?: {
        onOutput?: (text: string) => void;
        onError?: (error: Error) => void;
        onInput?: (prompt: string) => string;
    }): Promise<void>;
    compile(source: string): { id: number; success: boolean };
    run(program: { id: number }): {
        success: boolean;
        output: string;
        executionTime: number;
        error?: Error;
    };
    eval(source: string): {
        success: boolean;
        output: string;
        executionTime: number;
        error?: Error;
    };
    on(event: 'output' | 'error' | 'input', callback: Function): void;
    version(): { version: string; build: string; platform: string };
    dispose(): void;
}
```
