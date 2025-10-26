# WebAssembly Compilation Design

**Date**: 2025-10-26
**Status**: Design Phase
**Related PLAN.md Item**: 10.15

## Overview

This document describes the design for adding comprehensive WebAssembly (WASM) compilation support to go-dws. The goal is to enable DWScript execution in web browsers through three main deliverables:

1. **Interactive Web Playground** - Browser-based REPL for trying DWScript online
2. **Embeddable NPM Library** - Package for integrating DWScript into web applications
3. **Cross-platform Distribution** - Alternative execution environment with WASM support

## Requirements

### Functional Requirements

- **FR1**: Full DWScript language compatibility in WASM build (100% feature parity with native build)
- **FR2**: Web playground with code editor, execution, and shareable URLs
- **FR3**: NPM package supporting both Node.js and browser environments
- **FR4**: Static site deployment via GitHub Pages (no backend required)
- **FR5**: Platform abstraction for I/O (filesystem, console, time, etc.)

### Non-Functional Requirements

- **NFR1**: WASM binary size ≤ 3MB uncompressed, ≤ 1MB gzipped
- **NFR2**: Playground load time ≤ 2 seconds on broadband
- **NFR3**: WASM initialization time ≤ 100ms
- **NFR4**: Execution speed ≥ 50% of native Go performance
- **NFR5**: Support for Chrome, Firefox, and Safari browsers

## Architecture

### Build Modes

The WASM build supports three compile-time modes (decision deferred to implementation):

| Mode | Description | Size | Complexity |
|------|-------------|------|------------|
| **Monolithic** | Single WASM binary with full interpreter | 2-5MB | Low - simple, self-contained |
| **Modular** | Core WASM + optional modules loaded on demand | 500KB-1MB initial | Medium - requires module loader |
| **Hybrid** | WASM core + JavaScript glue for I/O and browser APIs | 1-2MB | Medium - more integration code |

### Overall System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     DWScript Source Code                     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
          ┌────────────────────────┐
          │  Platform Abstraction   │
          │    (pkg/platform/)      │
          ├────────────┬───────────┤
          │   Native   │   WASM    │
          └────────────┴───────────┘
                 │            │
                 ▼            ▼
          ┌──────────┐  ┌──────────────┐
          │ Native   │  │ WASM Binary  │
          │ Binary   │  │ (.wasm)      │
          └──────────┘  └──────┬───────┘
                               │
                ┌──────────────┼──────────────┐
                ▼              ▼              ▼
          ┌──────────┐  ┌───────────┐  ┌──────────┐
          │   Web    │  │    NPM    │  │ Embedded │
          │Playground│  │  Package  │  │   Apps   │
          └──────────┘  └───────────┘  └──────────┘
```

### Build Infrastructure

**Directory Structure**:
```
build/wasm/              # WASM-specific build scripts
├── Makefile             # Build targets (wasm, wasm-test, wasm-optimize)
├── build.sh             # Main build script
└── optimize.sh          # wasm-opt optimization script

cmd/dwscript-wasm/       # WASM-specific entry point
└── main.go              # Exports to JavaScript via syscall/js

pkg/platform/            # Platform abstraction layer
├── platform.go          # Interfaces: FileSystem, Console, Platform
├── native/              # Native implementations
│   └── platform.go
└── wasm/                # WASM implementations
    ├── platform.go
    ├── filesystem.go    # Virtual FS (memory or IndexedDB)
    └── console.go       # JavaScript console bridge

pkg/wasm/                # WASM-specific bridge code
├── api.go               # JavaScript API exports
├── callbacks.go         # JS→Go callback system
└── utils.go             # Type conversions, memory management
```

**Build Commands**:
```bash
# Build WASM binary
make wasm                          # Default build
make wasm MODE=monolithic          # Monolithic build
make wasm MODE=modular             # Modular build
make wasm MODE=hybrid              # Hybrid build

# Optimize WASM binary
make wasm-optimize                 # Run wasm-opt

# Test WASM build
make wasm-test                     # Run tests in Node.js and browser
```

**Build Process**:
1. Set `GOOS=js GOARCH=wasm` environment variables
2. Build with `go build -o dwscript.wasm ./cmd/dwscript-wasm`
3. Copy `$(go env GOROOT)/misc/wasm/wasm_exec.js` to output
4. Optionally run `wasm-opt -O3 -o dwscript.opt.wasm dwscript.wasm`
5. Package for distribution (playground, NPM, etc.)

### Platform Abstraction Layer

**Purpose**: Separate platform-dependent code (file I/O, console, time) from core interpreter logic, enabling identical behavior across native and WASM builds.

**Core Interfaces**:
```go
// pkg/platform/platform.go

type FileSystem interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte) error
    ListDir(path string) ([]FileInfo, error)
    Delete(path string) error
    Exists(path string) bool
}

type Console interface {
    Print(s string)
    PrintLn(s string)
    ReadLine() (string, error)
}

type Platform interface {
    FS() FileSystem
    Console() Console
    Now() time.Time
    Sleep(duration time.Duration)
}
```

**Native Implementation** (`pkg/platform/native/`):
- Uses standard Go packages: `os`, `io`, `time`
- Direct filesystem access via `os.ReadFile`, `os.WriteFile`
- Terminal I/O via `os.Stdin`, `os.Stdout`

**WASM Implementation** (`pkg/platform/wasm/`):
- **Virtual Filesystem**: In-memory `map[string][]byte` or IndexedDB via `syscall/js`
- **Console Bridge**: Calls `console.log()` or user-provided JavaScript callbacks
- **Time Functions**: JavaScript `Date` API via `syscall/js`
- **Sleep**: JavaScript `setTimeout` with Promise/channel bridge (non-blocking)

**Feature Compatibility Matrix**:

| DWScript Feature | Native | WASM Alternative |
|------------------|--------|------------------|
| File I/O (`ReadFile`, `WriteFile`) | `os` package | Virtual FS (memory or IndexedDB) |
| `PrintLn()` | `stdout` | `console.log()` or callback |
| `ReadLn()` | `stdin` | `prompt()` or callback |
| Time functions | `time` package | JavaScript `Date` API |
| Random numbers | `math/rand` | `crypto.getRandomValues()` |
| HTTP requests | `net/http` | `fetch()` API |
| External libraries | CGo/syscall | JavaScript interop |

**Testing Strategy**:
- Shared test suite runs on both native and WASM builds
- CI tests WASM build using Node.js or headless browser (Playwright)
- Feature flags for platform-specific capabilities
- Clear documentation of limitations (e.g., filesystem persistence)

## JavaScript/Go Bridge

### JavaScript API Design

**Exported API** (JavaScript side):
```javascript
class DWScript {
  /**
   * Initialize the WASM module
   * @param {Object} options - Configuration options
   * @param {VirtualFS} options.fs - Custom filesystem implementation
   * @param {Function} options.onOutput - Output callback
   * @param {Function} options.onError - Error callback
   */
  async init(options = {});

  /**
   * Compile DWScript source code
   * @param {string} source - DWScript source code
   * @returns {Program} Compiled program object
   */
  compile(source);

  /**
   * Execute a compiled program
   * @param {Program} program - Previously compiled program
   * @returns {Result} Execution result
   */
  run(program);

  /**
   * Compile and execute in one step
   * @param {string} source - DWScript source code
   * @returns {Result} Execution result
   */
  eval(source);

  /**
   * Set custom filesystem implementation
   * @param {VirtualFS} fs - Filesystem object
   */
  setFileSystem(fs);

  /**
   * Register event listeners
   * @param {string} event - Event name (output, error, etc.)
   * @param {Function} callback - Event handler
   */
  on(event, callback);
}

// Result object
interface Result {
  success: boolean;
  output: string;
  error?: string;
  executionTime: number;
}
```

**Usage Example**:
```javascript
import DWScript from './dwscript.js';

const dws = new DWScript();
await dws.init();

// Simple evaluation
const result = dws.eval(`
  var x: Integer := 42;
  PrintLn('The answer is ' + IntToStr(x));
`);
console.log(result.output); // "The answer is 42"

// With custom output handler
dws.on('output', (text) => {
  document.getElementById('console').innerText += text + '\n';
});

// Compile once, run multiple times
const program = dws.compile('PrintLn("Hello");');
dws.run(program);
dws.run(program); // Reuse compiled program
```

### Go → JavaScript Bridge

**Implementation** (`pkg/wasm/api.go`):
```go
import "syscall/js"

func registerAPI() {
    js.Global().Set("DWScript", js.FuncOf(newDWScript))
}

func newDWScript(this js.Value, args []js.Value) interface{} {
    obj := make(map[string]interface{})
    obj["init"] = js.FuncOf(initWASM)
    obj["compile"] = js.FuncOf(compile)
    obj["run"] = js.FuncOf(run)
    obj["eval"] = js.FuncOf(eval)
    return obj
}
```

**Type Conversions**:
- Go `string` ↔ JavaScript `string`: `js.ValueOf(s)`, `v.String()`
- Go `int` ↔ JavaScript `number`: `js.ValueOf(i)`, `v.Int()`
- Go `bool` ↔ JavaScript `boolean`: `js.ValueOf(b)`, `v.Bool()`
- Go `error` → JavaScript `Error`: Custom error object with message and stack
- Go `struct` → JavaScript `Object`: Map fields to properties

**Memory Management**:
- Call `.Release()` on `js.Value` objects when done to prevent leaks
- Use `js.Global().Get("ArrayBuffer")` for binary data transfer
- Minimize crossing WASM/JS boundary (expensive)

### JavaScript → Go Bridge

**Callback System** (`pkg/wasm/callbacks.go`):
```go
type Callbacks struct {
    onOutput js.Value
    onError  js.Value
}

func (c *Callbacks) Output(s string) {
    if !c.onOutput.IsNull() {
        c.onOutput.Invoke(s)
    }
}

func (c *Callbacks) Error(err error) {
    if !c.onError.IsNull() {
        c.onError.Invoke(err.Error())
    }
}
```

**Virtual Filesystem Interface**:
```javascript
// JavaScript provides filesystem implementation
const customFS = {
  readFile: async (path) => { /* ... */ },
  writeFile: async (path, data) => { /* ... */ },
  listDir: async (path) => { /* ... */ },
  delete: async (path) => { /* ... */ }
};

dws.setFileSystem(customFS);
```

### Error Handling

**Go Panics → JavaScript Exceptions**:
```go
defer func() {
    if r := recover(); r != nil {
        errObj := map[string]interface{}{
            "message": fmt.Sprint(r),
            "stack": string(debug.Stack()),
        }
        return js.ValueOf(errObj)
    }
}()
```

**DWScript Runtime Errors**:
```javascript
// JavaScript receives structured error objects
{
  type: "RuntimeError",
  message: "Division by zero",
  line: 42,
  column: 15,
  file: "example.dws"
}
```

## Web Playground

### Directory Structure

```
playground/
├── index.html              # Main HTML page
├── app.js                  # Playground application logic
├── editor.js               # Monaco Editor integration
├── styles.css              # Styling
├── dwscript.wasm           # Compiled WASM binary
├── wasm_exec.js            # Go WASM support file
└── examples/               # Sample DWScript programs
    ├── hello.dws
    ├── fibonacci.dws
    ├── classes.dws
    └── ...
```

### Features

**Code Editor**:
- **Editor Library**: Monaco Editor (VS Code's editor engine)
- **Syntax Highlighting**: Custom DWScript language definition
- **Auto-completion**: Basic keyword completion
- **Error Markers**: Real-time syntax error highlighting
- **Themes**: Light and dark themes

**User Interface**:
- **Layout**: Split-pane with resizable divider
  - Left: Code editor (60% width)
  - Right: Output console (40% width)
- **Toolbar**:
  - Run button (Ctrl+Enter keyboard shortcut)
  - Examples dropdown menu
  - Clear output button
  - Share button (copy URL)
  - Theme toggle
- **Output Console**:
  - Displays program output
  - Shows compilation and runtime errors
  - Execution time display
  - Clear button

**Code Sharing**:
- Encode source code in URL fragment: `#code=base64EncodedSource`
- No backend required (fully client-side)
- Example: `https://cwbudde.github.io/go-dws/#code=dmFyIHg6IEludGVnZXI7...`
- Shareable links work offline after initial load

**Local Storage**:
- Auto-save current code to `localStorage` every 5 seconds
- Restore code on page reload
- Clear button to reset to default example

**Examples**:
- Pre-built example programs showcasing DWScript features
- Categories: Basic, Control Flow, Functions, Classes, Advanced
- Load example button replaces current code

### Monaco Editor Integration

**Language Definition** (`playground/monaco-dwscript.js`):
```javascript
monaco.languages.register({ id: 'dwscript' });

monaco.languages.setMonarchTokensProvider('dwscript', {
  keywords: [
    'var', 'const', 'type', 'function', 'procedure', 'begin', 'end',
    'if', 'then', 'else', 'while', 'do', 'for', 'to', 'downto',
    'case', 'of', 'class', 'property', 'constructor', 'destructor',
    // ... more keywords
  ],
  operators: [':=', '=', '<>', '<', '>', '<=', '>=', '+', '-', '*', '/', 'div', 'mod'],
  // ... tokenizer rules
});
```

**Error Markers**:
```javascript
// Update editor with syntax errors from WASM
const markers = errors.map(err => ({
  severity: monaco.MarkerSeverity.Error,
  startLineNumber: err.line,
  startColumn: err.column,
  endLineNumber: err.line,
  endColumn: err.column + err.length,
  message: err.message
}));
monaco.editor.setModelMarkers(model, 'dwscript', markers);
```

### GitHub Pages Deployment

**Build Process**:
```bash
# Makefile target
playground: wasm
    mkdir -p playground/dist
    cp build/wasm/dwscript.wasm playground/dist/
    cp build/wasm/wasm_exec.js playground/dist/
    cp playground/*.{html,js,css} playground/dist/
    cp -r playground/examples playground/dist/
```

**GitHub Actions Workflow** (`.github/workflows/playground-deploy.yml`):
```yaml
name: Deploy Playground

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make playground
      - uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./playground/dist
```

**URL**: `https://cwbudde.github.io/go-dws/`

**Performance**:
- Static files served from GitHub's CDN
- WASM binary cached by browser
- First load: ~2 seconds
- Subsequent loads: <500ms (cached)

## NPM Package

### Package Structure

```
npm/
├── package.json            # NPM package metadata
├── README.md              # Usage documentation
├── index.js               # Main entry point (ESM)
├── index.cjs              # CommonJS entry point
├── dwscript.wasm          # WASM binary
├── wasm_exec.js           # Go WASM loader
├── loader.js              # WASM initialization helper
├── typescript/
│   └── index.d.ts         # TypeScript type definitions
├── examples/
│   ├── node.js            # Node.js usage example
│   ├── react.jsx          # React integration
│   ├── vue.js             # Vue.js integration
│   └── vanilla.html       # Plain JavaScript
└── LICENSE                # MIT license
```

### package.json

```json
{
  "name": "@cwbudde/dwscript",
  "version": "0.1.0",
  "description": "DWScript interpreter compiled to WebAssembly",
  "main": "index.cjs",
  "module": "index.js",
  "types": "typescript/index.d.ts",
  "exports": {
    ".": {
      "import": "./index.js",
      "require": "./index.cjs",
      "types": "./typescript/index.d.ts"
    },
    "./dwscript.wasm": "./dwscript.wasm"
  },
  "files": [
    "index.js",
    "index.cjs",
    "loader.js",
    "dwscript.wasm",
    "wasm_exec.js",
    "typescript/"
  ],
  "keywords": ["dwscript", "delphi", "pascal", "interpreter", "wasm"],
  "author": "cwbudde",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/cwbudde/go-dws"
  }
}
```

### TypeScript Definitions

```typescript
// typescript/index.d.ts

export interface DWScriptOptions {
  fs?: VirtualFileSystem;
  onOutput?: (text: string) => void;
  onError?: (error: RuntimeError) => void;
}

export interface Program {
  readonly id: number;
}

export interface Result {
  success: boolean;
  output: string;
  error?: RuntimeError;
  executionTime: number;
}

export interface RuntimeError {
  type: string;
  message: string;
  line?: number;
  column?: number;
  file?: string;
}

export interface VirtualFileSystem {
  readFile(path: string): Promise<Uint8Array>;
  writeFile(path: string, data: Uint8Array): Promise<void>;
  listDir(path: string): Promise<string[]>;
  delete(path: string): Promise<void>;
}

export default class DWScript {
  constructor();
  init(options?: DWScriptOptions): Promise<void>;
  compile(source: string): Program;
  run(program: Program): Result;
  eval(source: string): Result;
  setFileSystem(fs: VirtualFileSystem): void;
  on(event: 'output' | 'error', callback: Function): void;
}
```

### Usage Examples

**Node.js**:
```javascript
// examples/node.js
import DWScript from '@cwbudde/dwscript';
import fs from 'fs';

const dws = new DWScript();
await dws.init();

const source = fs.readFileSync('program.dws', 'utf-8');
const result = dws.eval(source);
console.log(result.output);
```

**React**:
```jsx
// examples/react.jsx
import React, { useState, useEffect } from 'react';
import DWScript from '@cwbudde/dwscript';

function DWScriptRunner() {
  const [dws, setDws] = useState(null);
  const [output, setOutput] = useState('');

  useEffect(() => {
    const init = async () => {
      const instance = new DWScript();
      await instance.init({ onOutput: setOutput });
      setDws(instance);
    };
    init();
  }, []);

  const runCode = (code) => {
    if (dws) {
      const result = dws.eval(code);
      setOutput(result.output);
    }
  };

  return <div>{/* UI */}</div>;
}
```

**Web Workers**:
```javascript
// Run DWScript in background thread
const worker = new Worker('dwscript-worker.js');
worker.postMessage({ code: 'PrintLn("Hello");' });
worker.onmessage = (e) => console.log(e.data.output);
```

### Publishing

**Automated Publishing** (`.github/workflows/npm-publish.yml`):
```yaml
name: Publish NPM Package

on:
  release:
    types: [published]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
          registry-url: 'https://registry.npmjs.org'
      - run: make npm-package
      - run: cd npm && npm publish --provenance
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

**Registry**: npmjs.com (public)
**Package URL**: `https://www.npmjs.com/package/@cwbudde/dwscript`

## Testing Strategy

### WASM-Specific Tests

**Unit Tests** (`pkg/wasm/*_test.go`, `pkg/platform/wasm/*_test.go`):
```bash
GOOS=js GOARCH=wasm go test ./pkg/wasm ./pkg/platform/wasm
```

**Integration Tests** (Node.js test runner):
```javascript
// tests/wasm/integration.test.js
import { describe, it, expect } from 'vitest';
import DWScript from '../../npm/index.js';

describe('DWScript WASM', () => {
  it('should execute basic arithmetic', async () => {
    const dws = new DWScript();
    await dws.init();
    const result = dws.eval('PrintLn(IntToStr(2 + 2));');
    expect(result.output).toBe('4\n');
  });
});
```

**Browser Tests** (Playwright):
```javascript
// tests/e2e/playground.spec.js
import { test, expect } from '@playwright/test';

test('playground runs code', async ({ page }) => {
  await page.goto('http://localhost:8080');
  await page.fill('.editor', 'PrintLn("Hello");');
  await page.click('#run-button');
  await expect(page.locator('.output')).toContainText('Hello');
});
```

### Cross-Browser Testing

**CI Matrix** (`.github/workflows/wasm-test.yml`):
```yaml
strategy:
  matrix:
    browser: [chromium, firefox, webkit]
```

**Tested Browsers**:
- Chrome/Chromium (latest)
- Firefox (latest)
- Safari/WebKit (latest)

### Performance Testing

**Bundle Size Monitoring**:
```yaml
# .github/workflows/bundle-size.yml
- name: Check WASM size
  run: |
    SIZE=$(stat -f%z dwscript.wasm)
    if [ $SIZE -gt 3145728 ]; then  # 3MB
      echo "WASM binary too large: $SIZE bytes"
      exit 1
    fi
```

**Execution Benchmarks**:
```javascript
// tests/benchmarks/performance.js
const iterations = 1000;
const start = performance.now();
for (let i = 0; i < iterations; i++) {
  dws.eval('var x := 42 * 2;');
}
const elapsed = performance.now() - start;
console.log(`${iterations} iterations: ${elapsed}ms (${elapsed/iterations}ms each)`);
```

### Feature Parity Tests

**Shared Test Suite**:
- Same DWScript test files run on native and WASM builds
- Compare output for identical behavior
- Example: `testdata/features/*.dws`

**CI Validation**:
```bash
# Native
./dwscript run testdata/features/arrays.dws > native-output.txt

# WASM (Node.js)
node wasm-runner.js testdata/features/arrays.dws > wasm-output.txt

# Compare
diff native-output.txt wasm-output.txt
```

## Documentation

### Required Documentation

1. **`docs/wasm/BUILD.md`** - Building WASM binaries
   - Prerequisites (Go 1.21+, wasm-opt)
   - Build commands and modes
   - Optimization techniques
   - Troubleshooting

2. **`docs/wasm/API.md`** - JavaScript API reference
   - Complete API documentation
   - Usage examples for each method
   - Error handling guide
   - Type reference

3. **`docs/wasm/EMBEDDING.md`** - Embedding in web applications
   - Integration patterns (React, Vue, Svelte, vanilla JS)
   - Web Workers guide
   - Service Workers for offline support
   - Performance best practices
   - Security considerations

4. **`docs/wasm/PLAYGROUND.md`** - Playground architecture
   - Directory structure
   - Monaco Editor integration
   - Customization guide
   - Deployment process

5. **`npm/README.md`** - NPM package usage
   - Installation instructions
   - Quick start examples
   - API reference
   - Browser/Node.js differences
   - TypeScript usage

6. **Main README.md update**:
   - Add WASM section
   - Link to playground
   - Link to NPM package
   - Browser compatibility table

## Performance Expectations

### Binary Size

| Build Mode | Uncompressed | Gzipped | Brotli |
|------------|--------------|---------|--------|
| Monolithic | 2-3 MB | 600-900 KB | 500-800 KB |
| Modular (core) | 800 KB - 1.2 MB | 300-500 KB | 250-450 KB |
| Hybrid | 1.5-2.5 MB | 500-800 KB | 400-700 KB |

### Execution Speed

- **WASM vs Native**: 50-80% of native Go performance (typical for Go WASM)
- **Startup Time**: <100ms to initialize WASM module
- **Compilation**: ~1ms for simple programs (<100 LOC)
- **Execution**: Depends on program complexity

### Network Performance

- **First Load** (uncached): 1-2 seconds on broadband (5 Mbps)
- **Subsequent Loads** (cached): <500ms
- **Playground Load**: ~1.5 seconds total (WASM + Monaco Editor)

## CI/CD Pipeline

### GitHub Actions Workflows

1. **`wasm-build.yml`** - Build WASM on every push
   ```yaml
   on: [push, pull_request]
   jobs:
     build:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v3
         - uses: actions/setup-go@v4
         - run: make wasm
         - uses: actions/upload-artifact@v3
           with:
             name: dwscript-wasm
             path: build/wasm/dwscript.wasm
   ```

2. **`wasm-test.yml`** - Run WASM tests
   ```yaml
   jobs:
     test:
       strategy:
         matrix:
           browser: [chromium, firefox, webkit]
       steps:
         - run: npm install
         - run: npx playwright test --project=${{ matrix.browser }}
   ```

3. **`playground-deploy.yml`** - Deploy playground to GitHub Pages
   - Trigger: Push to `main` branch
   - Build playground
   - Deploy to `gh-pages` branch

4. **`npm-publish.yml`** - Publish NPM package
   - Trigger: Release tag (e.g., `v1.0.0`)
   - Build WASM and NPM package
   - Publish to npmjs.com with provenance
   - Attach WASM binary to GitHub Release

### Release Process

1. **Create Release Tag**: `git tag -a v1.0.0 -m "Release 1.0.0"`
2. **Push Tag**: `git push origin v1.0.0`
3. **CI Automation**:
   - Builds native binaries (Linux, macOS, Windows)
   - Builds WASM binary
   - Publishes NPM package
   - Deploys playground
   - Creates GitHub Release with binaries
4. **Manual Step**: Update release notes on GitHub

## Security Considerations

### Sandboxing

- WASM runs in browser sandbox (no file system access, network restrictions)
- Virtual filesystem is isolated per-instance
- No access to user's real filesystem

### Resource Limits

- **Memory**: Browser enforces WASM memory limits (typically 2-4GB)
- **Execution Time**: Implement timeout for long-running scripts (playground)
- **CPU**: JavaScript's `requestIdleCallback` for non-blocking execution

### Untrusted Code Execution

- Playground runs user code in WASM sandbox
- No eval() or arbitrary JavaScript execution from DWScript
- XSS protection: Output sanitized before DOM insertion

## Future Enhancements

### Post-MVP Features

1. **Persistent Storage**: IndexedDB-backed virtual filesystem
2. **Multi-file Projects**: Support multiple .dws files with imports
3. **Debugger Integration**: Breakpoints and step-through in playground
4. **Code Completion**: Context-aware autocomplete in Monaco
5. **Share Backend**: Optional backend for saving/sharing snippets
6. **Syntax Themes**: Additional Monaco themes for DWScript
7. **Mobile Support**: Responsive playground for tablets/phones
8. **Offline PWA**: Progressive Web App with Service Worker

### Performance Optimizations

1. **Streaming Compilation**: Use `WebAssembly.compileStreaming()`
2. **SIMD Support**: Leverage WASM SIMD instructions
3. **Multi-threading**: WASM threads for parallel execution
4. **Code Splitting**: Lazy-load WASM modules
5. **Ahead-of-Time Compilation**: Pre-compile common DWScript patterns

## Open Questions

1. **Build Mode Selection**: Should we support all three modes or pick one? (Deferred)
2. **IndexedDB vs In-Memory**: Default filesystem implementation? (Start with in-memory)
3. **Editor Library**: Monaco vs CodeMirror? (Recommendation: Monaco for better features)
4. **NPM Scope**: `@cwbudde/dwscript` vs `dwscript-wasm`? (Recommendation: `@cwbudde/dwscript`)

## References

- [Go WebAssembly Documentation](https://github.com/golang/go/wiki/WebAssembly)
- [Monaco Editor](https://microsoft.github.io/monaco-editor/)
- [Playwright Testing](https://playwright.dev/)
- [GitHub Pages Deployment](https://docs.github.com/en/pages)
- [NPM Publishing Guide](https://docs.npmjs.com/packages-and-modules/contributing-packages-to-the-registry)
