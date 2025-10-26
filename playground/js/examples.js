/**
 * Playground example loader.
 *
 * Instead of embedding DWScript snippets directly in JavaScript, we now
 * reference the canonical example programs stored under ../examples/scripts.
 * When the playground is hosted standalone (without repository root), the
 * loader will fall back to files placed under playground/examples/.
 */

const EXAMPLE_BASE_PATHS = [
    '../examples/scripts/',
    'examples/'
];

function createExample(key, name, description, filename) {
    return {
        key,
        name,
        description,
        file: filename,
        paths: EXAMPLE_BASE_PATHS.map((base) => `${base}${filename}`),
        code: null,
        source: null
    };
}

const EXAMPLES = {
    hello: createExample(
        'hello',
        'Hello World',
        'A simple Hello World program',
        'hello_world.dws'
    ),
    fibonacci: createExample(
        'fibonacci',
        'Fibonacci Sequence',
        'Calculate Fibonacci numbers using recursion',
        'fibonacci.dws'
    ),
    factorial: createExample(
        'factorial',
        'Factorial',
        'Calculate factorial using recursion and iteration',
        'factorial.dws'
    ),
    loops: createExample(
        'loops',
        'Loops',
        'Demonstrate different loop structures',
        'loops.dws'
    ),
    functions: createExample(
        'functions',
        'Functions',
        'Functions, procedures, and parameters',
        'functions.dws'
    ),
    classes: createExample(
        'classes',
        'Classes (OOP)',
        'Object-oriented programming with classes',
        'classes.dws'
    ),
    math: createExample(
        'math',
        'Math Operations',
        'Mathematical calculations and operators',
        'math_operations.dws'
    ),
    caseStatement: createExample(
        'caseStatement',
        'Case Statements',
        'Map numeric values to labels using case-of blocks',
        'case_statement.dws'
    ),
    palindrome: createExample(
        'palindrome',
        'Palindrome Checker',
        'Test words using custom string helpers',
        'palindrome_checker.dws'
    ),
    primes: createExample(
        'primes',
        'Prime Numbers',
        'Generate primes with helper functions',
        'prime_numbers.dws'
    ),
    table: createExample(
        'table',
        'Multiplication Table',
        'Build a formatted multiplication grid with nested loops',
        'multiplication_table.dws'
    )
};

function getExample(key) {
    return EXAMPLES[key] || null;
}

function getExampleKeys() {
    return Object.keys(EXAMPLES);
}

function getExampleList() {
    return getExampleKeys().map((key) => {
        const { name, description } = EXAMPLES[key];
        return { key, name, description };
    });
}

async function loadExampleCode(key) {
    const example = getExample(key);
    if (!example) {
        throw new Error(`Unknown example: ${key}`);
    }

    if (example.code) {
        return example;
    }

    example.code = await fetchExampleSource(example);
    return example;
}

async function fetchExampleSource(example) {
    const errors = [];

    for (const path of example.paths) {
        try {
            const response = await fetch(path, { cache: 'no-cache' });
            if (!response.ok) {
                errors.push(`${path} (${response.status})`);
                continue;
            }

            const text = await response.text();
            example.source = path;
            return text;
        } catch (err) {
            errors.push(`${path} (${err.message})`);
        }
    }

    throw new Error(
        `Unable to load ${example.file}. Tried: ${errors.join('; ')}`
    );
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        EXAMPLES,
        getExample,
        getExampleKeys,
        getExampleList,
        loadExampleCode
    };
} else {
    window.EXAMPLES = EXAMPLES;
    window.getExample = getExample;
    window.getExampleKeys = getExampleKeys;
    window.getExampleList = getExampleList;
    window.loadExampleCode = loadExampleCode;
}
