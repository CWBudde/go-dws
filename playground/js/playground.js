/**
 * DWScript Playground Main Script
 *
 * Manages the Monaco Editor, WASM integration, UI interactions,
 * and all playground functionality.
 */

// Global state
let editor = null;
let dws = null;
let isWasmReady = false;
let currentTheme = 'light';
let decorations = [];

// Storage keys
const STORAGE_KEY_CODE = 'dwscript_playground_code';
const STORAGE_KEY_THEME = 'dwscript_playground_theme';

// Default code
const DEFAULT_CODE = `// Welcome to DWScript Playground!
// Try some code below:

var message: String := 'Hello from DWScript!';
var x: Integer := 42;

PrintLn(message);
PrintLn('The answer is ' + IntToStr(x));

// Try changing the code and click Run
`;

/**
 * Initialize Monaco Editor
 */
function initMonaco() {
    require.config({ paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.44.0/min/vs' } });

    require(['vs/editor/editor.main'], function () {
        // Register DWScript language
        registerDWScriptLanguage(monaco);

        // Load theme preference
        currentTheme = localStorage.getItem(STORAGE_KEY_THEME) || 'light';
        applyTheme(currentTheme, false);

        // Get initial code (from URL, localStorage, or default)
        const initialCode = getInitialCode();

        // Create the editor
        editor = monaco.editor.create(document.getElementById('editor'), {
            value: initialCode,
            language: 'dwscript',
            theme: currentTheme === 'dark' ? 'dwscript-dark' : 'dwscript-light',
            fontSize: 14,
            lineNumbers: 'on',
            minimap: { enabled: true },
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 4,
            insertSpaces: true,
            wordWrap: 'on',
            contextmenu: true,
            quickSuggestions: false
        });

        // Set up event listeners
        setupEditorEvents();
        setupUIEvents();
        setupResizer();

        // Update editor info
        updateEditorInfo();

        // Auto-save on change
        editor.onDidChangeModelContent(() => {
            saveCodeToStorage();
            updateEditorInfo();
        });

        // Keyboard shortcuts
        editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, runCode);
        editor.addCommand(monaco.KeyMod.Alt | monaco.KeyMod.Shift | monaco.KeyCode.KeyF, formatCode);

        updateStatus('Monaco Editor loaded', 'ready');
    });
}

/**
 * Initialize WASM
 */
async function initWasm() {
    try {
        updateStatus('Loading WebAssembly module...', 'loading');

        const go = new Go();
        const result = await WebAssembly.instantiateStreaming(
            fetch('wasm/dwscript.wasm'),
            go.importObject
        );

        // Start the Go program
        go.run(result.instance);

        // Wait for initialization
        await new Promise(resolve => setTimeout(resolve, 100));

        // Create DWScript instance
        dws = new DWScript();

        // Initialize with callbacks
        await dws.init({
            onOutput: (text) => {
                appendOutput(text);
            },
            onError: (error) => {
                console.error('DWScript error:', error);
            }
        });

        // Get version info
        const version = dws.version();
        document.getElementById('version').textContent = `v${version.version}`;

        isWasmReady = true;
        updateStatus('Ready to run DWScript code', 'ready');
        enableControls();

    } catch (error) {
        console.error('WASM initialization error:', error);
        updateStatus('Failed to load WebAssembly: ' + error.message, 'error');
        appendOutput('❌ Failed to initialize DWScript WASM module\n' + error.message, 'error');
    }
}

/**
 * Get initial code from URL, localStorage, or default
 */
function getInitialCode() {
    // Check URL fragment for shared code
    const hash = window.location.hash.substr(1);
    if (hash) {
        try {
            return atob(hash);
        } catch (e) {
            console.warn('Failed to decode URL fragment:', e);
        }
    }

    // Check localStorage
    const saved = localStorage.getItem(STORAGE_KEY_CODE);
    if (saved) {
        return saved;
    }

    // Use default
    return DEFAULT_CODE;
}

/**
 * Save code to localStorage
 */
function saveCodeToStorage() {
    if (editor) {
        localStorage.setItem(STORAGE_KEY_CODE, editor.getValue());
    }
}

/**
 * Run the code in the editor
 */
function runCode() {
    if (!isWasmReady || !dws) {
        appendOutput('⚠️ DWScript WASM not ready yet\n', 'warning');
        return;
    }

    const code = editor.getValue();
    if (!code.trim()) {
        appendOutput('⚠️ No code to run\n', 'warning');
        return;
    }

    // Clear previous decorations
    clearErrorMarkers();

    // Clear output
    clearOutput();

    try {
        updateStatus('Running...', 'loading');
        const startTime = performance.now();

        const result = dws.eval(code);
        const endTime = performance.now();
        const executionTime = (endTime - startTime).toFixed(2);

        if (result.success) {
            if (result.output) {
                appendOutput(result.output);
            } else {
                appendOutput('// Program completed successfully (no output)\n', 'info');
            }
            updateStats(`Execution time: ${result.executionTime}ms (total: ${executionTime}ms)`);
            updateStatus('Execution completed', 'ready');
            document.getElementById('outputInfo').textContent = 'Success';
        } else {
            appendOutput('❌ Runtime Error:\n', 'error');
            appendOutput(result.error.message + '\n', 'error');
            updateStatus('Runtime error', 'error');
            document.getElementById('outputInfo').textContent = 'Error';
        }

    } catch (error) {
        if (error.type === 'CompileError') {
            appendOutput('❌ Compilation Error:\n', 'error');
            appendOutput(error.message + '\n', 'error');

            // Try to parse error location and add markers
            addErrorMarkers(error.message);

            updateStatus('Compilation error', 'error');
            document.getElementById('outputInfo').textContent = 'Error';
        } else {
            appendOutput('❌ Error: ' + error.message + '\n', 'error');
            updateStatus('Error', 'error');
        }
        console.error('Execution error:', error);
    }
}

/**
 * Clear output console
 */
function clearOutput() {
    const output = document.getElementById('output');
    output.textContent = '';
    document.getElementById('outputInfo').textContent = 'Ready';
    updateStats('');
}

/**
 * Append text to output console
 */
function appendOutput(text, type = '') {
    const output = document.getElementById('output');
    const span = document.createElement('span');
    span.className = type ? `output-${type}` : 'output-line';
    span.textContent = text;
    output.appendChild(span);
    output.scrollTop = output.scrollHeight;
}

/**
 * Update status bar
 */
function updateStatus(message, status = 'loading') {
    const statusBar = document.getElementById('statusBar');
    const statusText = document.getElementById('statusText');
    statusText.textContent = message;
    statusBar.className = `status-bar ${status}`;
}

/**
 * Update stats in status bar
 */
function updateStats(text) {
    document.getElementById('statsText').textContent = text;
}

/**
 * Update editor info
 */
function updateEditorInfo() {
    if (!editor) return;

    const model = editor.getModel();
    const lineCount = model.getLineCount();
    const position = editor.getPosition();

    document.getElementById('editorInfo').textContent =
        `Line ${position.lineNumber}, Column ${position.column} | ${lineCount} lines`;
}

/**
 * Load example code
 */
function loadExample(key) {
    const example = getExample(key);
    if (example && editor) {
        editor.setValue(example.code);
        clearOutput();
        appendOutput(`// Loaded: ${example.name}\n// ${example.description}\n\n`, 'info');
    }
}

/**
 * Share code via URL
 */
function shareCode() {
    if (!editor) return;

    const code = editor.getValue();
    const encoded = btoa(code);
    const url = window.location.origin + window.location.pathname + '#' + encoded;

    // Copy to clipboard
    if (navigator.clipboard) {
        navigator.clipboard.writeText(url).then(() => {
            showNotification('Share URL copied to clipboard!');
        }).catch(err => {
            console.error('Failed to copy:', err);
            showNotification('Failed to copy URL');
        });
    } else {
        // Fallback
        const input = document.createElement('input');
        input.value = url;
        document.body.appendChild(input);
        input.select();
        document.execCommand('copy');
        document.body.removeChild(input);
        showNotification('Share URL copied to clipboard!');
    }
}

/**
 * Toggle theme
 */
function toggleTheme() {
    currentTheme = currentTheme === 'light' ? 'dark' : 'light';
    applyTheme(currentTheme, true);
    localStorage.setItem(STORAGE_KEY_THEME, currentTheme);
}

/**
 * Apply theme
 */
function applyTheme(theme, updateEditor = true) {
    document.body.setAttribute('data-theme', theme);

    if (updateEditor && editor) {
        monaco.editor.setTheme(theme === 'dark' ? 'dwscript-dark' : 'dwscript-light');
    }
}

/**
 * Format code
 */
function formatCode() {
    if (editor) {
        editor.getAction('editor.action.formatDocument').run();
    }
}

/**
 * Add error markers to editor
 */
function addErrorMarkers(errorMessage) {
    if (!editor) return;

    // Try to parse line number from error message
    // Format: "Error at line X: ..." or "Line X: ..."
    const lineMatch = errorMessage.match(/line (\d+)/i);
    if (lineMatch) {
        const lineNumber = parseInt(lineMatch[1], 10);

        decorations = editor.deltaDecorations(decorations, [
            {
                range: new monaco.Range(lineNumber, 1, lineNumber, 1),
                options: {
                    isWholeLine: true,
                    className: 'error-line',
                    glyphMarginClassName: 'error-glyph',
                    marginClassName: 'error-margin'
                }
            }
        ]);

        // Add marker
        monaco.editor.setModelMarkers(editor.getModel(), 'dwscript', [
            {
                startLineNumber: lineNumber,
                startColumn: 1,
                endLineNumber: lineNumber,
                endColumn: Number.MAX_VALUE,
                message: errorMessage,
                severity: monaco.MarkerSeverity.Error
            }
        ]);

        // Jump to error line
        editor.revealLineInCenter(lineNumber);
    }
}

/**
 * Clear error markers
 */
function clearErrorMarkers() {
    if (!editor) return;

    decorations = editor.deltaDecorations(decorations, []);
    monaco.editor.setModelMarkers(editor.getModel(), 'dwscript', []);
}

/**
 * Show notification
 */
function showNotification(message) {
    const notification = document.createElement('div');
    notification.className = 'copy-notification';
    notification.textContent = message;
    document.body.appendChild(notification);

    setTimeout(() => {
        document.body.removeChild(notification);
    }, 2000);
}

/**
 * Enable controls after WASM is ready
 */
function enableControls() {
    document.getElementById('btnRun').disabled = false;
}

/**
 * Set up editor events
 */
function setupEditorEvents() {
    // Position changes
    if (editor) {
        editor.onDidChangeCursorPosition(updateEditorInfo);
    }
}

/**
 * Set up UI event listeners
 */
function setupUIEvents() {
    // Run button
    document.getElementById('btnRun').addEventListener('click', runCode);

    // Clear button
    document.getElementById('btnClear').addEventListener('click', clearOutput);

    // Examples dropdown
    document.getElementById('selExamples').addEventListener('change', (e) => {
        const key = e.target.value;
        if (key) {
            loadExample(key);
            e.target.value = ''; // Reset dropdown
        }
    });

    // Share button
    document.getElementById('btnShare').addEventListener('click', shareCode);

    // Theme button
    document.getElementById('btnTheme').addEventListener('click', toggleTheme);

    // Format button
    document.getElementById('btnFormat').addEventListener('click', formatCode);
}

/**
 * Set up split pane resizer
 */
function setupResizer() {
    const resizer = document.getElementById('resizer');
    const container = document.querySelector('.container');
    const editorPanel = document.querySelector('.panel-editor');
    const outputPanel = document.querySelector('.panel-output');

    let isResizing = false;
    let startX = 0;
    let startWidthEditor = 0;
    let startWidthOutput = 0;

    resizer.addEventListener('mousedown', (e) => {
        isResizing = true;
        startX = e.clientX;
        startWidthEditor = editorPanel.offsetWidth;
        startWidthOutput = outputPanel.offsetWidth;
        resizer.classList.add('active');
        e.preventDefault();
    });

    document.addEventListener('mousemove', (e) => {
        if (!isResizing) return;

        const deltaX = e.clientX - startX;
        const newWidthEditor = startWidthEditor + deltaX;
        const newWidthOutput = startWidthOutput - deltaX;

        const minWidth = 300;
        if (newWidthEditor >= minWidth && newWidthOutput >= minWidth) {
            editorPanel.style.flex = `0 0 ${newWidthEditor}px`;
            outputPanel.style.flex = `0 0 ${newWidthOutput}px`;
        }
    });

    document.addEventListener('mouseup', () => {
        if (isResizing) {
            isResizing = false;
            resizer.classList.remove('active');
        }
    });
}

// Initialize playground on page load
window.addEventListener('DOMContentLoaded', () => {
    initMonaco();
    initWasm();
});

// Handle page visibility changes
document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') {
        // Auto-save when page becomes visible
        saveCodeToStorage();
    }
});
