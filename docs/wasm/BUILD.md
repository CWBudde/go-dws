# Building DWScript for WebAssembly

This document describes how to build the DWScript interpreter for WebAssembly (WASM).

## Prerequisites

- **Go 1.21+** with WebAssembly support
- **just** command runner (`cargo install just` or `brew install just`)
- **bc** calculator (for size calculations, usually pre-installed)
- **wasm-opt** (optional, for optimization): `npm install -g binaryen`

## Quick Start

```bash
# Build WASM binary (default: monolithic mode)
just wasm

# Build with optimization
just wasm-opt

# Test that WASM compiles
just wasm-test

# Clean build artifacts
just wasm-clean

# Show binary size
just wasm-size
```

## Build Modes

DWScript supports three build modes:

### 1. Monolithic (Default)
Single WASM binary with all features included.

```bash
just wasm monolithic
```

- **Size**: ~4-5 MB uncompressed
- **Complexity**: Low
- **Use case**: Simple deployments, single-file distribution

### 2. Modular
Core WASM with optional modules loaded on demand.

```bash
just wasm modular
```

- **Size**: ~500KB-1MB initial
- **Complexity**: Medium
- **Use case**: Progressive loading, feature-based splitting

### 3. Hybrid
WASM core with JavaScript glue for I/O and browser APIs.

```bash
just wasm hybrid
```

- **Size**: ~1-2 MB
- **Complexity**: Medium
- **Use case**: Tight browser integration, custom I/O

## Build Outputs

After building, files are placed in `build/wasm/dist/`:

```
build/wasm/dist/
├── dwscript.wasm    # Compiled WASM binary
└── wasm_exec.js     # Go WASM runtime support
```

## Build Process Details

### Manual Build

If you want to build manually without `just`:

```bash
# Set environment for WASM target
export GOOS=js
export GOARCH=wasm

# Build the binary
go build -o dwscript.wasm ./cmd/dwscript-wasm

# Copy wasm_exec.js from Go distribution
cp "$(go env GOROOT)/go/lib/wasm/wasm_exec.js" .
```

### Build Script

The build script `build/wasm/build.sh` handles:

1. Setting `GOOS=js GOARCH=wasm` environment
2. Building with appropriate flags: `-ldflags=-s`
3. Copying `wasm_exec.js` from Go distribution
4. Size checking and warnings (fails if > 3MB without optimization)
5. Optional optimization with `wasm-opt`

**Usage:**

```bash
./build/wasm/build.sh [mode]

# With optimization
OPTIMIZE=true ./build/wasm/build.sh
```

### Build Flags

The build uses the following Go build flags:

- `-ldflags=-s`: Strip debug symbols to reduce size
- `-tags=wasm_*`: Build tags for different modes (future)

## Optimization

### Using wasm-opt

The `wasm-opt` tool from Binaryen can significantly reduce binary size:

```bash
# Optimize during build
just wasm-opt

# Optimize existing binary
just wasm-optimize build/wasm/dist/dwscript.wasm

# Manual optimization
wasm-opt -O3 -o dwscript.opt.wasm dwscript.wasm
```

**Expected results:**
- Original: ~4.2 MB
- Optimized: ~2.5-3 MB (30-40% reduction)
- Gzipped: ~1-1.5 MB

### Size Monitoring

The build process automatically checks binary size:

```bash
just wasm-size
```

If the uncompressed binary exceeds 3 MB, a warning is displayed.

## Testing the Build

### Compile-Only Test

Quick test that WASM compiles without errors:

```bash
just wasm-test
```

### Browser Test

To test in a browser, you need an HTML file:

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>DWScript WASM Test</title>
</head>
<body>
    <h1>DWScript WASM</h1>
    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(
            fetch("dwscript.wasm"),
            go.importObject
        ).then((result) => {
            go.run(result.instance);

            // Test the API
            const dws = new DWScript();
            const result = dws.eval('PrintLn("Hello from WASM!");');
            console.log(result);
        });
    </script>
</body>
</html>
```

Serve with a local HTTP server:

```bash
# Using Python
cd build/wasm/dist
python3 -m http.server 8080

# Using Node.js
npx http-server build/wasm/dist -p 8080
```

Then open http://localhost:8080 in your browser.

### Node.js Test

Test in Node.js environment:

```bash
# Install Node.js WASM polyfill
npm install -g wasm-exec

# Run test
cd build/wasm/dist
node wasm_exec.js dwscript.wasm
```

## Troubleshooting

### Build fails with "undefined: js"

Make sure you're using the correct build tags:

```bash
GOOS=js GOARCH=wasm go build ./cmd/dwscript-wasm
```

### wasm_exec.js not found

The script searches multiple locations. If it still fails:

```bash
# Find wasm_exec.js manually
find $(go env GOROOT) -name "wasm_exec.js"

# Copy manually
cp /path/to/wasm_exec.js build/wasm/dist/
```

### Binary too large

Try these optimization strategies:

1. **Use wasm-opt**: `just wasm-opt`
2. **Strip more aggressively**: Add `-ldflags=-w` (removes DWARF)
3. **Use modular mode**: `just wasm modular`
4. **Analyze size**: `wasm-objdump -h dwscript.wasm`

### WASM execution errors in browser

Check browser console for errors:

- **CORS errors**: Serve via HTTP server, not `file://`
- **Memory errors**: WASM needs sufficient memory
- **Module errors**: Ensure wasm_exec.js is loaded first

## Performance Expectations

### Binary Size

| Build Mode | Uncompressed | Optimized | Gzipped |
|------------|--------------|-----------|---------|
| Monolithic | 4-5 MB | 2.5-3 MB | 1-1.5 MB |
| Modular | 1-2 MB | 800KB-1.2 MB | 400-600 KB |
| Hybrid | 2-3 MB | 1.5-2 MB | 700KB-1 MB |

### Execution Speed

- **WASM vs Native**: 50-80% of native Go performance
- **Startup Time**: < 100ms to initialize WASM module
- **Compilation**: ~1ms for simple programs (<100 LOC)

### Network Performance

- **First Load** (uncached): 1-2 seconds on broadband (5 Mbps)
- **Subsequent Loads** (cached): < 500ms
- **With gzip**: ~50-70% size reduction

## CI/CD Integration

### GitHub Actions

```yaml
name: Build WASM

on: [push]

jobs:
  wasm:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install just
        run: cargo install just

      - name: Build WASM
        run: just wasm

      - name: Check size
        run: just wasm-size

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: dwscript-wasm
          path: build/wasm/dist/
```

## Next Steps

- **Web Playground**: See `docs/wasm/PLAYGROUND.md` (coming soon)
- **NPM Package**: See `docs/wasm/NPM.md` (coming soon)
- **JavaScript API**: See `docs/wasm/API.md` (coming soon)

## References

- [Go WebAssembly Documentation](https://github.com/golang/go/wiki/WebAssembly)
- [Binaryen wasm-opt](https://github.com/WebAssembly/binaryen)
- [WebAssembly Specification](https://webassembly.github.io/spec/)
