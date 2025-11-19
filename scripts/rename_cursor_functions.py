#!/usr/bin/env python3
"""
Script to rename parse*Cursor() functions to parse*() and remove dispatchers.

This implements Subtask 2.7.4.2 Step 2: Rename all parse*Cursor() to parse*()
"""

import re
from pathlib import Path

def find_dispatchers(content):
    """Find all dispatcher functions that just delegate to Cursor versions."""
    dispatchers = []

    # Pattern: func (p *Parser) parseSomething(...) ... {
    #              return p.parseSomethingCursor(...)
    #          }
    lines = content.split('\n')
    i = 0
    while i < len(lines):
        line = lines[i]

        # Look for function declaration
        func_match = re.match(r'^func \(p \*Parser\) (parse\w+)\((.*?)\)(.*?)\{', line)
        if func_match:
            func_name = func_match.group(1)
            params = func_match.group(2)
            return_type = func_match.group(3).strip()

            # Check if next non-empty line is a return statement calling Cursor version
            j = i + 1
            while j < len(lines) and not lines[j].strip():
                j += 1

            if j < len(lines):
                return_match = re.match(r'\s*return p\.(\w+Cursor)\(', lines[j])
                if return_match:
                    cursor_func = return_match.group(1)
                    expected_cursor = func_name + "Cursor"

                    # Verify it's calling the corresponding Cursor function
                    if cursor_func == expected_cursor:
                        # Find the closing brace
                        k = j + 1
                        while k < len(lines) and lines[k].strip() != '}':
                            k += 1

                        if k < len(lines):
                            dispatchers.append({
                                'name': func_name,
                                'cursor_name': cursor_func,
                                'start_line': i,
                                'end_line': k,
                                'params': params,
                                'return_type': return_type
                            })
                            i = k  # Skip to end of this function
        i += 1

    return dispatchers

def remove_dispatchers(content, dispatchers):
    """Remove dispatcher functions from content."""
    lines = content.split('\n')

    # Sort dispatchers by start_line in reverse order (remove from bottom up)
    dispatchers_sorted = sorted(dispatchers, key=lambda d: d['start_line'], reverse=True)

    for dispatcher in dispatchers_sorted:
        start = dispatcher['start_line']
        end = dispatcher['end_line']

        # Remove the function and any blank lines immediately before it
        # (but keep at least one blank line)
        while start > 0 and not lines[start - 1].strip():
            start -= 1
        if start > 0:
            start += 1  # Keep one blank line

        # Remove lines
        del lines[start:end + 1]

    return '\n'.join(lines)

def rename_cursor_functions(content, dispatchers):
    """Rename parse*Cursor functions to parse*."""
    # Create a set of cursor function names to rename
    cursor_names = {d['cursor_name'] for d in dispatchers}

    lines = content.split('\n')
    new_lines = []

    for line in lines:
        new_line = line

        # Rename function declarations: func (p *Parser) parseSomethingCursor( -> parseSomething(
        for cursor_name in cursor_names:
            base_name = cursor_name[:-6]  # Remove "Cursor" suffix
            # Function declaration
            new_line = re.sub(
                r'\bfunc \(p \*Parser\) ' + cursor_name + r'\(',
                f'func (p *Parser) {base_name}(',
                new_line
            )

        new_lines.append(new_line)

    return '\n'.join(new_lines)

def process_file(filepath):
    """Process a single Go file."""
    print(f"Processing {filepath}...")

    with open(filepath, 'r') as f:
        content = f.read()

    # Find all dispatchers
    dispatchers = find_dispatchers(content)

    if not dispatchers:
        print(f"  No dispatchers found in {filepath}")
        return 0

    print(f"  Found {len(dispatchers)} dispatchers:")
    for d in dispatchers:
        print(f"    - {d['name']} -> {d['cursor_name']}")

    # Step 1: Remove dispatchers
    content = remove_dispatchers(content, dispatchers)

    # Step 2: Rename Cursor functions
    content = rename_cursor_functions(content, dispatchers)

    # Write back
    with open(filepath, 'w') as f:
        f.write(content)

    print(f"  ✓ Renamed {len(dispatchers)} functions in {filepath}")
    return len(dispatchers)

def main():
    # Use path relative to script location for portability
    script_dir = Path(__file__).parent
    parser_dir = script_dir.parent / "internal" / "parser"

    # Get all .go files except tests
    go_files = sorted([
        f for f in parser_dir.glob("*.go")
        if not f.name.endswith("_test.go")
    ])

    total_renamed = 0
    for filepath in go_files:
        total_renamed += process_file(filepath)

    print(f"\n✓ Total functions renamed: {total_renamed}")

if __name__ == "__main__":
    main()
