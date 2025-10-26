# DWScript Web Playground

The DWScript Web Playground is an interactive, browser-based environment for writing, testing, and sharing DWScript code. It features the Monaco Editor (VS Code's editor), WebAssembly-powered execution, and a rich set of tools for learning and experimenting with DWScript.

## Features

### Core Functionality

- **Monaco Editor Integration**: Full-featured code editor with:
  - Syntax highlighting for DWScript
  - Line numbers and code folding
  - Minimap for code navigation
  - Multi-cursor editing
  - Find and replace
  - Bracket matching
  - Auto-indentation

- **Real-time Execution**: Run DWScript code directly in the browser using WebAssembly

- **Output Console**: Split-pane view showing:
  - Program output (PrintLn, etc.)
  - Compilation errors with line numbers
  - Runtime errors with stack traces
  - Execution statistics (time, memory)

- **Example Programs**: Pre-loaded examples demonstrating:
  - Hello World
  - Fibonacci sequence
  - Factorial (recursive and iterative)
  - Loop structures
  - Functions and procedures
  - Object-oriented programming
  - Math operations

### User Experience Features

- **Code Sharing**: Share code via URL using base64-encoded fragments
- **Auto-save**: Automatically saves code to localStorage
- **Theme Support**: Light and dark themes with Monaco integration
- **Keyboard Shortcuts**:
  - `Ctrl+Enter` (or `Cmd+Enter`): Run code
  - `Alt+Shift+F`: Format code
  - Monaco's built-in shortcuts (Ctrl+F for find, etc.)

- **Responsive Design**: Works on desktop and mobile devices
- **Error Highlighting**: Visual markers for compilation errors in the editor
- **Split Pane Resizing**: Adjustable editor/output split

## Architecture

### Directory Structure

```
playground/
├── index.html              # Main HTML page
├── css/
│   └── playground.css      # Styling for the playground
├── js/
│   ├── dwscript-lang.js    # Monaco language definition
│   ├── examples.js         # Example programs
│   └── playground.js       # Main application logic
└── examples/
    └── (future: standalone example files)
```

### Component Overview

```
┌─────────────────────────────────────────────────────┐
│                    Browser                           │
│  ┌───────────────────────────────────────────────┐  │
│  │           Monaco Editor (JavaScript)          │  │
│  │  ┌─────────────────────────────────────────┐  │  │
│  │  │   DWScript Language Definition          │  │  │
│  │  │   - Syntax highlighting                 │  │  │
│  │  │   - Tokenization rules                  │  │  │
│  │  │   - Auto-completion (future)            │  │  │
│  │  └─────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────┘  │
│                       ↓                              │
│  ┌───────────────────────────────────────────────┐  │
│  │        Playground Controller (JS)             │  │
│  │  - Event handling                             │  │
│  │  - State management                           │  │
│  │  - localStorage integration                   │  │
│  │  - URL sharing                                │  │
│  └───────────────────────────────────────────────┘  │
│                       ↓                              │
│  ┌───────────────────────────────────────────────┐  │
│  │         DWScript WASM Module                  │  │
│  │  ┌─────────────────────────────────────────┐  │  │
│  │  │  Go WASM Runtime (wasm_exec.js)         │  │  │
│  │  └─────────────────────────────────────────┘  │  │
│  │  ┌─────────────────────────────────────────┐  │  │
│  │  │  DWScript Engine (Go → WASM)            │  │  │
│  │  │  - Lexer                                │  │  │
│  │  │  - Parser                               │  │  │
│  │  │  - Semantic Analyzer                    │  │  │
│  │  │  - Interpreter                          │  │  │
│  │  └─────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────┘  │
│                       ↓                              │
│  ┌───────────────────────────────────────────────┐  │
│  │            Output Console                     │  │
│  │  - Formatted output                           │  │
│  │  - Error messages                             │  │
│  │  - Execution stats                            │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Data Flow

1. **User Input**: User types code in Monaco Editor
2. **Auto-save**: Code is saved to localStorage on change
3. **Run**: User clicks Run button or presses Ctrl+Enter
4. **Compilation**: Code is sent to WASM module for compilation
5. **Error Handling**:
   - If compilation fails, errors are displayed with line markers
   - If compilation succeeds, execution begins
6. **Execution**: WASM module executes the compiled code
7. **Output**: Results are sent back to JavaScript and displayed in console
8. **Stats**: Execution time and other metrics are shown

### Monaco Editor Integration

The playground uses Monaco Editor (the editor that powers VS Code) with a custom DWScript language definition:

**Language Definition** (`dwscript-lang.js`):
- **Keywords**: All DWScript keywords (begin, end, var, function, etc.)
- **Type Keywords**: Built-in types (Integer, String, Float, etc.)
- **Operators**: All DWScript operators (=, :=, +, -, div, mod, etc.)
- **Comments**: Line (`//`) and block (`{ }`, `(* *)`)
- **Strings**: Single and double quotes with escape sequences
- **Numbers**: Integer, hex ($), binary (%), and float literals

**Themes**:
- **Light Theme**: VS Code-style light theme optimized for DWScript
- **Dark Theme**: VS Code-style dark theme with DWScript syntax colors

### WASM Integration

The playground loads the DWScript WASM module built from Go:

**Loading Process**:
1. Load `wasm_exec.js` from Go distribution
2. Instantiate WASM module using `WebAssembly.instantiateStreaming()`
3. Create `DWScript` instance via JavaScript API
4. Initialize with output/error callbacks
5. Ready to compile and execute code

**JavaScript API** (see [API.md](API.md)):
- `dws.compile(code)`: Compile code and return program object
- `dws.run(program)`: Execute a compiled program
- `dws.eval(code)`: Compile and run in one step
- `dws.on(event, callback)`: Register event listeners

## Usage Guide

### Running the Playground Locally

1. **Build WASM module**:
   ```bash
   just wasm
   ```

2. **Start a local server**:
   ```bash
   # Python
   cd playground
   python3 -m http.server 8080

   # Node.js
   npx http-server playground -p 8080

   # Go
   cd playground && go run -tags=dev server.go
   ```

3. **Open in browser**:
   ```
   http://localhost:8080
   ```

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+Enter` (or `Cmd+Enter`) | Run code |
| `Ctrl+S` (or `Cmd+S`) | Save (auto-saves to localStorage) |
| `Alt+Shift+F` | Format code |
| `Ctrl+F` (or `Cmd+F`) | Find |
| `Ctrl+H` (or `Cmd+H`) | Find and replace |
| `Ctrl+/` (or `Cmd+/`) | Toggle line comment |
| `Alt+Up/Down` | Move line up/down |
| `Ctrl+D` (or `Cmd+D`) | Add selection to next match |

### URL Sharing

To share code:
1. Write or load code in the editor
2. Click the "Share" button
3. URL is copied to clipboard
4. Share URL with others

The URL contains base64-encoded code in the fragment (`#`), so the code is stored client-side (not on any server).

**Example**:
```
https://example.com/playground#dmFyIG1zZzogU3RyaW5nIDo9ICdIZWxsbyc7ClByaW50TG4obXNnKTs=
```

### localStorage Features

The playground automatically saves:
- **Code**: Current editor content (on every change)
- **Theme**: Light or dark theme preference

**Storage Keys**:
- `dwscript_playground_code`: Editor content
- `dwscript_playground_theme`: Theme preference ('light' or 'dark')

**Privacy**: All data is stored locally in the browser. Nothing is sent to any server.

## Deployment

### GitHub Pages

The playground is automatically deployed to GitHub Pages via GitHub Actions:

**Workflow**: `.github/workflows/deploy-playground.yml`

**Triggers**:
- Push to `main` branch (when playground files change)
- Manual workflow dispatch

**Process**:
1. Checkout code
2. Set up Go 1.21+
3. Install `just` command runner
4. Build WASM module (`just wasm`)
5. Copy playground files to `_site/`
6. Copy WASM build to `_site/wasm/`
7. Upload to GitHub Pages
8. Deploy

**Deployment URL**: `https://<username>.github.io/go-dws/`

### Custom Domain

To deploy on a custom domain:

1. **Update CNAME** (if using GitHub Pages):
   ```bash
   echo "playground.example.com" > playground/CNAME
   ```

2. **Configure DNS**:
   - Add CNAME record pointing to `<username>.github.io`

3. **Update links in code** (if needed):
   - Update WASM paths in `index.html` if different from default

## Customization

### Adding New Examples

Edit `playground/js/examples.js`:

```javascript
const EXAMPLES = {
    myExample: {
        name: 'My Example',
        description: 'Description of what this example does',
        code: `// DWScript code here
var x: Integer := 42;
PrintLn(IntToStr(x));`
    },
    // ... other examples
};
```

### Modifying Themes

Edit theme definitions in `playground/js/dwscript-lang.js`:

```javascript
const dwscriptTheme = {
    base: 'vs',
    inherit: true,
    rules: [
        { token: 'keyword', foreground: '0000FF', fontStyle: 'bold' },
        // ... more token rules
    ],
    colors: {
        'editor.background': '#FFFFFF',
        // ... more editor colors
    }
};
```

### Changing Editor Settings

Modify editor options in `playground/js/playground.js`:

```javascript
editor = monaco.editor.create(document.getElementById('editor'), {
    fontSize: 14,              // Font size
    minimap: { enabled: true }, // Show/hide minimap
    wordWrap: 'on',            // Word wrap
    // ... other options
});
```

## Browser Compatibility

### Supported Browsers

| Browser | Version | Status |
|---------|---------|--------|
| Chrome | 57+ | ✅ Fully supported |
| Firefox | 52+ | ✅ Fully supported |
| Safari | 11+ | ✅ Fully supported |
| Edge | 16+ | ✅ Fully supported |

### Required Features

- **WebAssembly**: Core requirement for running DWScript
- **ES6 Modules**: For loading Monaco Editor
- **localStorage**: For code persistence
- **Clipboard API**: For share functionality (optional)

### Checking Support

The playground automatically detects WebAssembly support and shows an error if unavailable.

## Performance

### Load Time

- **First Load** (uncached):
  - Monaco Editor: ~500ms
  - WASM Module: ~200-500ms (depending on size)
  - Total: ~700-1000ms

- **Subsequent Loads** (cached):
  - Monaco Editor: ~100ms
  - WASM Module: ~50ms
  - Total: ~150ms

### Execution Speed

- **Simple Programs** (<100 lines): < 10ms
- **Medium Programs** (100-500 lines): 10-50ms
- **Complex Programs** (500+ lines): 50-200ms

WASM execution is typically 50-80% of native Go speed.

### Optimization Tips

1. **Use CDN for Monaco**: Already configured (jsDelivr)
2. **Enable Browser Caching**: Configured via GitHub Pages
3. **Compress WASM**: Use `just wasm-opt` for smaller binaries
4. **Lazy Load Examples**: Load on demand rather than bundling

## Troubleshooting

### Playground Won't Load

**Problem**: White screen or "Loading..." never completes

**Solutions**:
1. Check browser console for errors
2. Ensure WASM file is accessible (check Network tab)
3. Verify browser supports WebAssembly
4. Clear browser cache and reload

### Code Won't Run

**Problem**: Clicking "Run" does nothing or shows error

**Solutions**:
1. Check that WASM initialized (status bar should say "Ready")
2. Look for compilation errors in output console
3. Verify code syntax is correct
4. Check browser console for JavaScript errors

### Share URL Doesn't Work

**Problem**: Shared URL doesn't load code correctly

**Solutions**:
1. Verify URL wasn't truncated when copying
2. Check that URL contains `#` followed by encoded code
3. Try manually encoding/decoding with base64 tools
4. Use shorter code snippets (URL length limits)

### Theme Not Persisting

**Problem**: Theme resets on page reload

**Solutions**:
1. Check that localStorage is enabled in browser
2. Verify no browser extensions are blocking storage
3. Check browser privacy settings (some modes disable localStorage)

## Future Enhancements

Planned features for future versions:

- [ ] **Auto-completion**: Intelligent code completion
- [ ] **Code snippets**: Quick insertion of common patterns
- [ ] **Multi-file support**: Work with multiple files/modules
- [ ] **Debugging**: Breakpoints and step-through execution
- [ ] **Mobile optimization**: Better touch support
- [ ] **Offline mode**: Service Worker for offline usage
- [ ] **Export/Import**: Download/upload code files
- [ ] **Collaboration**: Real-time collaborative editing
- [ ] **Version history**: Undo/redo beyond editor default
- [ ] **Performance profiling**: Visual profiler for code

## Contributing

To contribute to the playground:

1. Fork the repository
2. Create a feature branch
3. Make your changes in `playground/`
4. Test locally (see "Running Locally" above)
5. Submit a pull request

**Areas for Contribution**:
- New example programs
- Theme improvements
- UI/UX enhancements
- Documentation
- Bug fixes

## See Also

- [API.md](API.md) - JavaScript API documentation
- [BUILD.md](BUILD.md) - Building the WASM module
- [../../README.md](../../README.md) - Project overview
- [Monaco Editor Documentation](https://microsoft.github.io/monaco-editor/)
- [WebAssembly](https://webassembly.org/)

## License

The DWScript Playground is part of the go-dws project and is licensed under the same terms as the main project.
