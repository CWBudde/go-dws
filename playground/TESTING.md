# Playground Testing Checklist

This document provides a comprehensive testing checklist for the DWScript Playground.

## Browser Compatibility Testing

### Chrome/Chromium (57+)

- [ ] Playground loads without errors
- [ ] Monaco Editor renders correctly
- [ ] WASM module loads successfully
- [ ] Code execution works
- [ ] Syntax highlighting is correct
- [ ] Error markers appear for compilation errors
- [ ] Share button copies URL to clipboard
- [ ] Theme toggle works (light/dark)
- [ ] Split pane resizer works
- [ ] localStorage persists code
- [ ] All examples load correctly
- [ ] Keyboard shortcuts work (Ctrl+Enter, Alt+Shift+F)

### Firefox (52+)

- [ ] Playground loads without errors
- [ ] Monaco Editor renders correctly
- [ ] WASM module loads successfully
- [ ] Code execution works
- [ ] Syntax highlighting is correct
- [ ] Error markers appear for compilation errors
- [ ] Share button copies URL to clipboard
- [ ] Theme toggle works (light/dark)
- [ ] Split pane resizer works
- [ ] localStorage persists code
- [ ] All examples load correctly
- [ ] Keyboard shortcuts work (Ctrl+Enter, Alt+Shift+F)

### Safari (11+)

- [ ] Playground loads without errors
- [ ] Monaco Editor renders correctly
- [ ] WASM module loads successfully
- [ ] Code execution works
- [ ] Syntax highlighting is correct
- [ ] Error markers appear for compilation errors
- [ ] Share button copies URL to clipboard
- [ ] Theme toggle works (light/dark)
- [ ] Split pane resizer works
- [ ] localStorage persists code
- [ ] All examples load correctly
- [ ] Keyboard shortcuts work (Cmd+Enter, Alt+Shift+F)

### Edge (16+)

- [ ] Playground loads without errors
- [ ] Monaco Editor renders correctly
- [ ] WASM module loads successfully
- [ ] Code execution works
- [ ] Syntax highlighting is correct
- [ ] Error markers appear for compilation errors
- [ ] Share button copies URL to clipboard
- [ ] Theme toggle works (light/dark)
- [ ] Split pane resizer works
- [ ] localStorage persists code
- [ ] All examples load correctly
- [ ] Keyboard shortcuts work (Ctrl+Enter, Alt+Shift+F)

## Functional Testing

### Code Execution

- [ ] Simple PrintLn statements work
- [ ] Variable declarations work
- [ ] Arithmetic operations work
- [ ] String concatenation works
- [ ] Control flow (if/else) works
- [ ] Loops (for/while/repeat) work
- [ ] Functions/procedures work
- [ ] Classes and OOP work
- [ ] Compilation errors are caught
- [ ] Runtime errors are caught
- [ ] Execution time is displayed

### Editor Features

- [ ] Code typing is responsive
- [ ] Syntax highlighting updates in real-time
- [ ] Line numbers are correct
- [ ] Code folding works
- [ ] Minimap displays correctly
- [ ] Find/Replace works (Ctrl+F, Ctrl+H)
- [ ] Multi-cursor editing works
- [ ] Auto-indentation works
- [ ] Bracket matching works
- [ ] Comments are highlighted correctly

### UI Features

- [ ] Status bar updates correctly
- [ ] Output console displays text correctly
- [ ] Output console scrolls to bottom on new output
- [ ] Clear button clears output
- [ ] Run button triggers execution
- [ ] Examples dropdown loads examples
- [ ] Toolbar buttons are responsive
- [ ] Panel resizer provides visual feedback

### Data Persistence

- [ ] Code saves to localStorage on change
- [ ] Code restores from localStorage on reload
- [ ] Theme preference persists
- [ ] URL fragment sharing works
- [ ] Shared URLs load code correctly
- [ ] Multiple tabs work independently

### Error Handling

- [ ] Compilation errors show in output
- [ ] Compilation errors create editor markers
- [ ] Runtime errors show in output
- [ ] WASM initialization errors are handled
- [ ] Network errors are handled gracefully
- [ ] Invalid examples don't crash the app

## Performance Testing

### Load Time

- [ ] Initial load completes in < 2 seconds (broadband)
- [ ] Monaco Editor loads in < 500ms
- [ ] WASM module loads in < 500ms
- [ ] Subsequent loads are faster (caching)

### Execution Performance

- [ ] Simple programs (< 10 lines) execute in < 10ms
- [ ] Medium programs (10-100 lines) execute in < 50ms
- [ ] Complex programs (100+ lines) execute in < 200ms
- [ ] No noticeable lag when typing
- [ ] No noticeable lag when switching themes

### Memory Usage

- [ ] No memory leaks after multiple executions
- [ ] Memory usage stays reasonable (< 100MB)
- [ ] No performance degradation over time

## Responsive Design Testing

### Desktop (1920x1080)

- [ ] Layout is well-proportioned
- [ ] All buttons are visible
- [ ] Split pane works correctly
- [ ] No horizontal scrolling

### Laptop (1366x768)

- [ ] Layout adapts correctly
- [ ] All buttons are visible
- [ ] Split pane works correctly
- [ ] No horizontal scrolling

### Tablet (768x1024)

- [ ] Layout adapts to portrait/landscape
- [ ] Touch interactions work
- [ ] Buttons are touch-friendly
- [ ] Virtual keyboard doesn't break layout

### Mobile (375x667)

- [ ] Layout is usable on small screens
- [ ] Touch interactions work
- [ ] Buttons are accessible
- [ ] Virtual keyboard integration works

## Accessibility Testing

- [ ] Keyboard navigation works
- [ ] Focus indicators are visible
- [ ] Color contrast is sufficient
- [ ] Screen reader compatibility (basic)
- [ ] No reliance on color alone for information

## Security Testing

- [ ] No XSS vulnerabilities in code execution
- [ ] No code injection vulnerabilities
- [ ] localStorage data is scoped correctly
- [ ] No sensitive data in URLs
- [ ] CORS policies are correct

## Testing Instructions

### Manual Testing Steps

1. **Start Local Server**:
   ```bash
   cd playground
   python3 -m http.server 8080
   ```

2. **Open in Browser**:
   - Navigate to http://localhost:8080
   - Open browser DevTools console

3. **Test Basic Functionality**:
   - Wait for "Ready" status
   - Type a simple program
   - Click Run
   - Verify output appears

4. **Test Examples**:
   - Load each example from dropdown
   - Run each example
   - Verify expected output

5. **Test Error Handling**:
   - Enter invalid syntax
   - Click Run
   - Verify error message appears
   - Verify error marker in editor

6. **Test Sharing**:
   - Write some code
   - Click Share button
   - Verify URL copied
   - Open URL in new tab
   - Verify code loads

7. **Test Persistence**:
   - Write some code
   - Reload page
   - Verify code persists

8. **Test Theme**:
   - Click Theme button
   - Verify theme changes
   - Reload page
   - Verify theme persists

### Automated Testing (Future)

Future plans include:
- Playwright tests for cross-browser automation
- Unit tests for JavaScript modules
- Integration tests for WASM communication
- Visual regression tests for UI

## Known Issues

Document any known issues here:

- [ ] Issue 1: Description
- [ ] Issue 2: Description

## Browser-Specific Issues

### Chrome
- None known

### Firefox
- None known

### Safari
- May require explicit user gesture for clipboard API

### Edge
- None known

## Reporting Issues

If you find issues during testing:

1. Check browser console for errors
2. Note exact browser version
3. Note exact steps to reproduce
4. Take screenshot if UI issue
5. Create GitHub issue with details

## Test Results

Last tested: (Date)

| Browser | Version | Status | Notes |
|---------|---------|--------|-------|
| Chrome | | ⏳ Pending | |
| Firefox | | ⏳ Pending | |
| Safari | | ⏳ Pending | |
| Edge | | ⏳ Pending | |
