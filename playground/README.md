# DWScript Web Playground

An interactive, browser-based environment for writing, testing, and sharing DWScript code.

## Quick Start

### Option 1: Using Python

```bash
# From the playground directory
python3 -m http.server 8080
```

Then open: http://localhost:8080

### Option 2: Using Node.js

```bash
# Install http-server globally (once)
npm install -g http-server

# From the playground directory
http-server -p 8080
```

Then open: http://localhost:8080

### Option 3: Using Go

```bash
# From the playground directory
go run -tags=dev << 'EOF'
package main
import (
    "fmt"
    "net/http"
)
func main() {
    fs := http.FileServer(http.Dir("."))
    http.Handle("/", fs)
    fmt.Println("Playground running at http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}
EOF
```

## Prerequisites

Before running the playground, ensure the WASM module is built:

```bash
# From the project root
just wasm
```

This creates `build/wasm/dist/dwscript.wasm` and `wasm_exec.js`.

## Features

- **Monaco Editor**: Full VS Code-style editor with DWScript syntax highlighting
- **Real-time Execution**: Run DWScript code in your browser via WebAssembly
- **Code Sharing**: Share code via URL (base64-encoded)
- **Auto-save**: Code persists in localStorage
- **Themes**: Light and dark themes
- **Examples**: Pre-loaded example programs
- **Error Highlighting**: Visual markers for compilation errors

## Browser Requirements

- Chrome 57+ / Firefox 52+ / Safari 11+ / Edge 16+
- WebAssembly support
- JavaScript enabled
- localStorage (for auto-save feature)

## File Structure

```
playground/
├── index.html              # Main HTML page
├── css/
│   └── playground.css      # Styling
├── js/
│   ├── dwscript-lang.js    # Monaco language definition
│   ├── examples.js         # Example programs
│   └── playground.js       # Main application logic
└── README.md               # This file
```

## Keyboard Shortcuts

- `Ctrl+Enter` (or `Cmd+Enter`): Run code
- `Alt+Shift+F`: Format code
- `Ctrl+F` (or `Cmd+F`): Find
- `Ctrl+H` (or `Cmd+H`): Find and replace

Plus all standard Monaco Editor shortcuts!

## Documentation

For detailed documentation, see:
- [PLAYGROUND.md](../docs/wasm/PLAYGROUND.md) - Complete playground documentation
- [API.md](../docs/wasm/API.md) - JavaScript API reference
- [BUILD.md](../docs/wasm/BUILD.md) - Building the WASM module

## Troubleshooting

### WASM module not loading

1. Ensure you built the WASM module: `just wasm`
2. Check that `../build/wasm/dist/` contains `dwscript.wasm` and `wasm_exec.js`
3. Use a proper HTTP server (not `file://` protocol)

### Code not running

1. Wait for status bar to say "Ready"
2. Check browser console for errors
3. Verify your browser supports WebAssembly

### CORS errors

Must serve via HTTP server, not `file://` protocol. Use one of the methods above.

## Development

To modify the playground:

1. Edit files in `playground/`
2. Refresh browser to see changes (no build step for HTML/CSS/JS)
3. If modifying WASM: rebuild with `just wasm`

## Deployment

The playground can be deployed to:
- GitHub Pages (automated via `.github/workflows/deploy-playground.yml`)
- Any static hosting service (Netlify, Vercel, etc.)
- Your own web server

Just copy the `playground/` directory and `build/wasm/dist/` to your host.

## License

Part of the go-dws project. See main repository for license details.
