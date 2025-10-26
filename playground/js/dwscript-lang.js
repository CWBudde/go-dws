/**
 * DWScript Language Definition for Monaco Editor
 *
 * Defines syntax highlighting, tokenization, and language features
 * for DWScript in Monaco Editor.
 */

// DWScript Language Configuration
const dwscriptLanguageConfig = {
    comments: {
        lineComment: '//',
        blockComment: ['{', '}']
    },
    brackets: [
        ['{', '}'],
        ['[', ']'],
        ['(', ')'],
        ['begin', 'end']
    ],
    autoClosingPairs: [
        { open: '{', close: '}' },
        { open: '[', close: ']' },
        { open: '(', close: ')' },
        { open: "'", close: "'", notIn: ['string', 'comment'] },
        { open: '"', close: '"', notIn: ['string'] }
    ],
    surroundingPairs: [
        { open: '{', close: '}' },
        { open: '[', close: ']' },
        { open: '(', close: ')' },
        { open: "'", close: "'" },
        { open: '"', close: '"' }
    ],
    folding: {
        markers: {
            start: new RegExp('^\\s*\\{|\\bbegin\\b'),
            end: new RegExp('^\\s*\\}|\\bend\\b')
        }
    }
};

// DWScript Monarch Tokenizer
const dwscriptMonarchDefinition = {
    defaultToken: '',
    tokenPostfix: '.dws',
    ignoreCase: true,

    keywords: [
        'and', 'array', 'as', 'begin', 'break', 'case', 'class', 'const',
        'constructor', 'continue', 'destructor', 'div', 'do', 'downto',
        'else', 'end', 'enum', 'except', 'exit', 'external', 'false',
        'finally', 'for', 'forward', 'function', 'if', 'implementation',
        'in', 'inherited', 'interface', 'is', 'mod', 'nil', 'not',
        'object', 'of', 'or', 'procedure', 'program', 'property',
        'raise', 'record', 'repeat', 'set', 'shl', 'shr', 'static',
        'then', 'to', 'true', 'try', 'type', 'unit', 'until', 'uses',
        'var', 'virtual', 'while', 'with', 'xor', 'override', 'abstract',
        'private', 'protected', 'public', 'published', 'operator', 'lazy',
        'deprecated', 'default', 'helper', 'strict', 'sealed', 'partial',
        'async', 'await', 'delegate', 'lambda', 'yield'
    ],

    typeKeywords: [
        'Integer', 'Float', 'String', 'Boolean', 'Variant', 'TObject',
        'TClass', 'TDateTime', 'Currency', 'Extended', 'Double',
        'Single', 'Byte', 'Word', 'Cardinal', 'Int64', 'UInt64'
    ],

    operators: [
        '=', '>', '<', '<=', '>=', '<>', ':=',
        '+', '-', '*', '/', 'div', 'mod',
        'and', 'or', 'xor', 'not', 'shl', 'shr',
        '+=', '-=', '*=', '/='
    ],

    symbols: /[=><!~?:&|+\-*\/\^%]+/,
    escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,

    tokenizer: {
        root: [
            // Identifiers and keywords
            [/[a-z_$][\w$]*/, {
                cases: {
                    '@keywords': 'keyword',
                    '@typeKeywords': 'type',
                    '@default': 'identifier'
                }
            }],

            // Whitespace
            { include: '@whitespace' },

            // Delimiters and operators
            [/[{}()\[\]]/, '@brackets'],
            [/[<>](?!@symbols)/, '@brackets'],
            [/@symbols/, {
                cases: {
                    '@operators': 'operator',
                    '@default': ''
                }
            }],

            // Numbers
            [/\$[0-9A-Fa-f]+/, 'number.hex'],
            [/%[01]+/, 'number.binary'],
            [/\d*\.\d+([eE][\-+]?\d+)?/, 'number.float'],
            [/\d+/, 'number'],

            // Delimiter: after number because of .\d floats
            [/[;,.]/, 'delimiter'],

            // Strings
            [/'([^'\\]|\\.)*$/, 'string.invalid'],  // non-terminated string
            [/'/, 'string', '@string'],
            [/"([^"\\]|\\.)*$/, 'string.invalid'],  // non-terminated string
            [/"/, 'string', '@string_double'],
        ],

        whitespace: [
            [/[ \t\r\n]+/, ''],
            [/\/\/.*$/, 'comment'],
            [/\{/, 'comment', '@comment'],
            [/\(\*/, 'comment', '@comment_paren'],
        ],

        comment: [
            [/[^\}]+/, 'comment'],
            [/\}/, 'comment', '@pop'],
        ],

        comment_paren: [
            [/[^\*\)]+/, 'comment'],
            [/\*\)/, 'comment', '@pop'],
        ],

        string: [
            [/[^\\']+/, 'string'],
            [/@escapes/, 'string.escape'],
            [/''/, 'string.escape'],  // doubled quote escaping
            [/'/, 'string', '@pop']
        ],

        string_double: [
            [/[^\\"]+/, 'string'],
            [/@escapes/, 'string.escape'],
            [/""/, 'string.escape'],  // doubled quote escaping
            [/"/, 'string', '@pop']
        ],
    },
};

// Theme definition
const dwscriptTheme = {
    base: 'vs',
    inherit: true,
    rules: [
        { token: 'keyword', foreground: '0000FF', fontStyle: 'bold' },
        { token: 'type', foreground: '267F99', fontStyle: 'bold' },
        { token: 'comment', foreground: '008000', fontStyle: 'italic' },
        { token: 'string', foreground: 'A31515' },
        { token: 'number', foreground: '098658' },
        { token: 'number.hex', foreground: '098658' },
        { token: 'number.binary', foreground: '098658' },
        { token: 'number.float', foreground: '098658' },
        { token: 'operator', foreground: '000000' },
        { token: 'delimiter', foreground: '000000' },
    ],
    colors: {
        'editor.foreground': '#000000',
        'editor.background': '#FFFFFF',
        'editorLineNumber.foreground': '#237893',
        'editorCursor.foreground': '#000000',
        'editor.selectionBackground': '#ADD6FF',
        'editor.inactiveSelectionBackground': '#E5EBF1',
    }
};

// Dark theme definition
const dwscriptDarkTheme = {
    base: 'vs-dark',
    inherit: true,
    rules: [
        { token: 'keyword', foreground: '569CD6', fontStyle: 'bold' },
        { token: 'type', foreground: '4EC9B0', fontStyle: 'bold' },
        { token: 'comment', foreground: '6A9955', fontStyle: 'italic' },
        { token: 'string', foreground: 'CE9178' },
        { token: 'number', foreground: 'B5CEA8' },
        { token: 'number.hex', foreground: 'B5CEA8' },
        { token: 'number.binary', foreground: 'B5CEA8' },
        { token: 'number.float', foreground: 'B5CEA8' },
        { token: 'operator', foreground: 'D4D4D4' },
        { token: 'delimiter', foreground: 'D4D4D4' },
    ],
    colors: {
        'editor.foreground': '#D4D4D4',
        'editor.background': '#1E1E1E',
        'editorLineNumber.foreground': '#858585',
        'editorCursor.foreground': '#AEAFAD',
        'editor.selectionBackground': '#264F78',
        'editor.inactiveSelectionBackground': '#3A3D41',
    }
};

// Register language with Monaco
function registerDWScriptLanguage(monaco) {
    // Register the language
    monaco.languages.register({ id: 'dwscript' });

    // Set the configuration
    monaco.languages.setLanguageConfiguration('dwscript', dwscriptLanguageConfig);

    // Set the tokenizer
    monaco.languages.setMonarchTokensProvider('dwscript', dwscriptMonarchDefinition);

    // Define the themes
    monaco.editor.defineTheme('dwscript-light', dwscriptTheme);
    monaco.editor.defineTheme('dwscript-dark', dwscriptDarkTheme);
}

// Export for use in playground
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { registerDWScriptLanguage };
}
