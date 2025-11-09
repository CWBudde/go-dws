# REFACTOR.md

## Overview

This document provides a comprehensive strategy for refactoring the go-dws codebase to improve maintainability, readability, and organization. The project has grown significantly, with 20+ files exceeding 30KB and several directories containing 60-90+ files.

**Current Issues:**
- Large files (up to 80KB) are difficult to navigate and maintain
- Flat directory structures make it hard to find related code
- Test files mixed with implementation files
- No clear organizational pattern for builtin functions

**Goals:**
- Split large files into focused, logical components
- Organize files into subdirectories by feature area
- Maintain parallel structure between `internal/interp/` and `internal/semantic/`
- Improve code discoverability and reduce cognitive load
- Preserve all functionality and test coverage

---

## File Splitting Strategy

### Priority 1: Critical Large Files (>50KB)

#### 1.1 internal/interp/objects.go (79KB, 2,357 lines)

**Current Content:** Object instantiation, member access, property operations, method calls, constructor/method lookup, inherited expressions, class constants

**Proposed Split:**

```
objects.go (REMOVE) →
├── objects/instantiation.go    // Object creation and constructors
├── objects/properties.go        // Property read/write operations
├── objects/methods.go          // Method calls and overload resolution
└── objects/hierarchy.go        // Hierarchy lookup and inheritance
```

**Split Details:**

**objects/instantiation.go** (~400 lines)
- `evalNewExpression()` - Object instantiation
- `executeConstructor()` - Constructor execution
- `createObjectInstance()` - Instance creation
- Constructor parameter handling
- Field initialization logic

**objects/properties.go** (~700 lines)
- `evalPropertyReadExpression()` - Instance property reads
- `evalPropertyWriteExpression()` - Instance property writes
- `evalClassPropertyReadExpression()` - Class property reads
- `evalClassPropertyWriteExpression()` - Class property writes
- `evalIndexedPropertyRead()` - Indexed property reads
- `evalIndexedPropertyWrite()` - Indexed property writes
- Property getter/setter invocation

**objects/methods.go** (~900 lines)
- `evalMethodCall()` - Method invocation
- `findMethod()` - Method lookup in class hierarchy
- `resolveMethodOverload()` - Overload resolution
- Method parameter matching
- Virtual method dispatch
- Method result handling

**objects/hierarchy.go** (~350 lines)
- `lookupInHierarchy()` - General hierarchy lookup
- `findFieldInHierarchy()` - Field search up the chain
- `findMethodInHierarchy()` - Method search up the chain
- `evalInheritedExpression()` - Inherited keyword handling
- `evalClassConstantExpression()` - Class constant access
- Ancestor traversal helpers

**Migration Notes:**
- Create `internal/interp/objects/` directory
- Update imports in `interpreter.go` and `expressions.go`
- Move related test code to `tests/objects/`

---

#### 1.2 internal/interp/functions.go (67KB, 2,048 lines)

**Current Content:** Call expression evaluation, builtin function dispatch, user function calls, function pointers, lambdas, record methods, type casting, overload resolution

**Proposed Split:**

```
functions.go (REMOVE) →
├── functions/calls.go          // Main call routing and dispatch
├── functions/builtins.go       // Builtin function handling
├── functions/user.go           // User-defined function execution
├── functions/pointers.go       // Function pointers and lambdas
├── functions/records.go        // Record method calls
└── functions/typecast.go       // Type casting operations
```

**Split Details:**

**functions/calls.go** (~500 lines)
- `evalCallExpression()` - Main call dispatcher
- `resolveCallTarget()` - Determine what's being called
- `prepareArguments()` - Argument evaluation and preparation
- Call type routing logic
- Error handling for invalid calls

**functions/builtins.go** (~600 lines)
- `callBuiltinFunction()` - Builtin dispatch
- `normalizeBuiltinName()` - Name normalization
- `validateBuiltinArgs()` - Argument validation
- `resolveBuiltinOverload()` - Overload selection
- Builtin function registry
- Parameter type checking

**functions/user.go** (~400 lines)
- `callUserFunction()` - User function execution
- `createFunctionScope()` - Scope creation for function
- `bindParameters()` - Parameter binding
- `handleVarParameters()` - Variable parameter handling
- `handleDefaultParameters()` - Default value handling
- Return value processing

**functions/pointers.go** (~350 lines)
- `callFunctionPointer()` - Function pointer invocation
- `evalLambdaExpression()` - Lambda evaluation
- `captureLambdaEnvironment()` - Closure capture
- `callLambda()` - Lambda invocation
- Function pointer type checking

**functions/records.go** (~400 lines)
- `callRecordMethod()` - Record method dispatch
- `findRecordMethod()` - Method lookup in record type
- `bindRecordSelf()` - Self parameter binding
- Record method parameter handling
- Record method result handling

**functions/typecast.go** (~300 lines)
- `evalTypeCastExpression()` - Type casting
- `castToType()` - Cast operation
- `isTypeCastable()` - Cast validity check
- Conversion functions (int→float, string→int, etc.)
- Dynamic cast support

**Migration Notes:**
- Create `internal/interp/functions/` directory
- Update imports throughout interpreter
- Consolidate function-related tests

---

#### 1.3 internal/interp/statements.go (52KB, 1,689 lines)

**Current Content:** Program evaluation, variable declarations, constant declarations, assignments (simple, member, index, compound), control flow (if, while, repeat, for, for-in, case), break/continue/return/exit

**Proposed Split:**

```
statements.go (REMOVE) →
├── statements/declarations.go  // Variable and constant declarations
├── statements/assignments.go   // All assignment types
├── statements/control.go       // if and case statements
└── statements/loops.go         // while, repeat, for, for-in, break, continue
```

**Split Details:**

**statements/declarations.go** (~400 lines)
- `evalVarDeclaration()` - Variable declarations
- `evalConstDeclaration()` - Constant declarations
- `evalTypeDeclaration()` - Type alias declarations
- Type resolution for declarations
- Initial value evaluation
- Multiple variable declaration handling

**statements/assignments.go** (~600 lines)
- `evalAssignStatement()` - Simple assignment (`:=`)
- `evalMemberAssignment()` - Member assignment (`obj.field := x`)
- `evalIndexAssignment()` - Index assignment (`arr[i] := x`)
- `evalCompoundAssignment()` - Compound assignment (`+=`, `-=`, etc.)
- Assignment type checking
- Left-hand side resolution
- Right-hand side evaluation

**statements/control.go** (~500 lines)
- `evalIfStatement()` - If/then/else
- `evalCaseStatement()` - Case/switch
- `evalCaseMatch()` - Case condition matching
- `evalCaseRange()` - Case range matching
- Control flow branching
- Condition evaluation

**statements/loops.go** (~600 lines)
- `evalWhileStatement()` - While loops
- `evalRepeatStatement()` - Repeat-until loops
- `evalForStatement()` - For loops (integer range)
- `evalForInStatement()` - For-in loops (collection iteration)
- `evalBreakStatement()` - Break handling
- `evalContinueStatement()` - Continue handling
- Loop control flow management
- Iterator handling

**Migration Notes:**
- Create `internal/interp/statements/` directory
- Keep `evalProgram()` and `evalBlockStatement()` in main `interpreter.go`
- Update statement evaluation dispatch

---

#### 1.4 internal/interp/builtins_datetime.go (40KB, 1,133 lines)

**Current Content:** 54 date/time builtin functions including Now, Date, Time, DateOf, TimeOf, EncodeDate, DecodeDate, FormatDateTime, parsing, calculations, etc.

**Proposed Split:**

```
builtins_datetime.go (REMOVE) →
├── builtins/datetime_format.go  // Date/time formatting and parsing
├── builtins/datetime_calc.go    // Date/time calculations and manipulation
└── builtins/datetime_info.go    // Date/time information extraction
```

**Split Details:**

**builtins/datetime_format.go** (~400 lines)
- `builtinFormatDateTime()` - Format date/time to string
- `builtinDateTimeToStr()` - Standard conversion
- `builtinStrToDateTime()` - Parse string to date/time
- `builtinTryStrToDateTime()` - Safe parsing
- Format string parsing logic
- Date/time formatting helpers

**builtins/datetime_calc.go** (~400 lines)
- `builtinIncMonth()` - Add/subtract months
- `builtinIncYear()` - Add/subtract years
- `builtinIncDay()` - Add/subtract days
- `builtinIncHour()`, `builtinIncMinute()`, `builtinIncSecond()`
- `builtinDaysBetween()`, `builtinMonthsBetween()`, `builtinYearsBetween()`
- `builtinDateTimeAddDays()`, `builtinDateTimeAddMonths()`
- Date/time arithmetic helpers

**builtins/datetime_info.go** (~333 lines)
- `builtinNow()` - Current date/time
- `builtinDate()` - Current date
- `builtinTime()` - Current time
- `builtinYearOf()`, `builtinMonthOf()`, `builtinDayOf()`
- `builtinHourOf()`, `builtinMinuteOf()`, `builtinSecondOf()`, `builtinMillisecondOf()`
- `builtinDayOfWeek()`, `builtinDayOfYear()`
- `builtinWeekOfYear()`, `builtinIsLeapYear()`
- `builtinDaysInMonth()`, `builtinDaysInYear()`
- Date/time component extraction

**Migration Notes:**
- Move to `internal/interp/builtins/` directory
- Update builtin registry in main interpreter
- Consolidate date/time test files

---

#### 1.5 internal/interp/builtins_math.go (35KB, 1,123 lines)

**Current Content:** 40 math builtin functions including trigonometry, logarithms, rounding, min/max, power, sqrt, etc.

**Proposed Split:**

```
builtins_math.go (REMOVE) →
├── builtins/math_basic.go       // Basic arithmetic and utility functions
├── builtins/math_trig.go        // Trigonometric and hyperbolic functions
└── builtins/math_convert.go     // Rounding, truncation, and conversions
```

**Split Details:**

**builtins/math_basic.go** (~350 lines)
- `builtinAbs()` - Absolute value
- `builtinMin()`, `builtinMax()` - Min/max functions
- `builtinSqr()` - Square
- `builtinSqrt()` - Square root
- `builtinPower()` - Exponentiation
- `builtinExp()` - Exponential
- `builtinLn()`, `builtinLog10()`, `builtinLog2()` - Logarithms
- `builtinSign()` - Sign function
- `builtinRandom()`, `builtinRandomize()` - Random numbers

**builtins/math_trig.go** (~400 lines)
- `builtinSin()`, `builtinCos()`, `builtinTan()` - Trigonometric
- `builtinArcSin()`, `builtinArcCos()`, `builtinArcTan()`, `builtinArcTan2()`
- `builtinSinh()`, `builtinCosh()`, `builtinTanh()` - Hyperbolic
- `builtinArcSinh()`, `builtinArcCosh()`, `builtinArcTanh()`
- `builtinDegToRad()`, `builtinRadToDeg()` - Angle conversions
- Trigonometry helpers

**builtins/math_convert.go** (~373 lines)
- `builtinRound()` - Round to nearest integer
- `builtinTrunc()` - Truncate decimal part
- `builtinFloor()` - Round down
- `builtinCeil()` - Round up
- `builtinFrac()` - Fractional part
- `builtinInt()` - Integer part
- `builtinIntPower()` - Integer exponentiation
- `builtinMod()` - Modulo
- `builtinDivMod()` - Division with remainder
- Type conversion helpers

**Migration Notes:**
- Move to `internal/interp/builtins/` directory
- Update builtin registry
- Organize math test files

---

### Priority 2: Large Files (40-50KB)

#### 2.1 internal/semantic/analyze_classes.go (48KB, 1,272 lines)

**Proposed Split:**

```
analyze_classes.go (REMOVE) →
├── classes/decl.go              // Class declaration analysis
├── classes/methods.go           // Method analysis and overrides
├── classes/properties.go        // Property and member analysis
└── classes/constructors.go      // Constructor synthesis and validation
```

**Split Details:**

**classes/decl.go** (~350 lines)
- `analyzeClassDeclaration()` - Main class analysis entry
- `validateClassInheritance()` - Inheritance validation
- `checkForwardReferences()` - Forward declaration handling
- `registerClassSymbol()` - Symbol table registration
- `validateAbstractClass()` - Abstract class rules
- Interface implementation checking

**classes/methods.go** (~400 lines)
- `analyzeMethodDeclaration()` - Method analysis
- `checkMethodOverride()` - Override validation
- `validateVirtualMethod()` - Virtual method rules
- `checkMethodSignatureMatch()` - Signature compatibility
- `analyzeMethodImplementation()` - Method body analysis
- Overload resolution checking

**classes/properties.go** (~300 lines)
- `analyzePropertyDeclaration()` - Property analysis
- `validatePropertyAccessors()` - Getter/setter validation
- `analyzeMemberAccess()` - Member access type checking
- `checkPropertyVisibility()` - Visibility rules
- Property type resolution

**classes/constructors.go** (~222 lines)
- `synthesizeDefaultConstructor()` - Default constructor creation
- `analyzeConstructorCall()` - Constructor invocation checking
- `validateConstructorChain()` - Constructor chaining
- `checkConstructorParameters()` - Parameter validation
- Inherited constructor handling

**Migration Notes:**
- Create `internal/semantic/classes/` directory
- Update analyzer imports
- Move class-related tests

---

#### 2.2 internal/bytecode/vm.go (47KB, 1,785 lines)

**Proposed Split:**

```
vm.go (REMOVE) →
├── vm/core.go                   // VM struct, Run loop, frame management
├── vm/stack.go                  // Stack operations
├── vm/builtins.go               // Builtin function implementations
└── vm/exceptions.go             // Exception handling
```

**Split Details:**

**vm/core.go** (~500 lines)
- `type VM struct` - VM structure
- `NewVM()` - VM initialization
- `Run()` - Main execution loop
- `executeInstruction()` - Instruction dispatch
- `pushFrame()`, `popFrame()` - Call frame management
- Program counter manipulation

**vm/stack.go** (~400 lines)
- `push()`, `pop()`, `peek()` - Basic stack operations
- `pushInt()`, `pushFloat()`, `pushString()`, etc. - Typed push
- `popInt()`, `popFloat()`, `popString()`, etc. - Typed pop
- Stack frame allocation
- Stack overflow checking
- Local variable access

**vm/builtins.go** (~600 lines)
- `callBuiltin()` - Builtin dispatch
- Implementation of each builtin in VM context
- Builtin parameter extraction from stack
- Builtin result pushing
- Builtin error handling

**vm/exceptions.go** (~285 lines)
- `raiseException()` - Exception raising
- `handleException()` - Exception catching
- `unwindStack()` - Stack unwinding
- Try/catch/finally frame management
- Exception object creation

**Migration Notes:**
- Create `internal/bytecode/vm/` directory
- Update compiler references
- Organize VM tests

---

#### 2.3 internal/bytecode/compiler.go (42KB, 1,744 lines)

**Proposed Split:**

```
compiler.go (REMOVE) →
├── compiler/core.go             // Compiler struct, compile entry points
├── compiler/statements.go       // Statement compilation
├── compiler/expressions.go      // Expression compilation
└── compiler/functions.go        // Function/lambda compilation
```

**Split Details:**

**compiler/core.go** (~400 lines)
- `type Compiler struct` - Compiler structure
- `NewCompiler()` - Compiler initialization
- `Compile()` - Main compilation entry
- `compileProgram()` - Program compilation
- Constant pool management
- Symbol table management
- Scope management

**compiler/statements.go** (~500 lines)
- `compileStatement()` - Statement dispatch
- `compileVarDeclaration()`, `compileConstDeclaration()`
- `compileAssignment()` - All assignment types
- `compileIfStatement()`, `compileCaseStatement()`
- `compileWhileLoop()`, `compileForLoop()`, `compileForInLoop()`
- `compileBreak()`, `compileContinue()`, `compileReturn()`
- Control flow and jump handling

**compiler/expressions.go** (~600 lines)
- `compileExpression()` - Expression dispatch
- `compileIntegerLiteral()`, `compileFloatLiteral()`, `compileStringLiteral()`
- `compileIdentifier()` - Variable access
- `compileBinaryExpression()`, `compileUnaryExpression()`
- `compileCallExpression()` - Function calls
- `compileIndexExpression()`, `compileMemberExpression()`
- Expression optimization

**compiler/functions.go** (~244 lines)
- `compileFunctionDeclaration()` - Function compilation
- `compileLambda()` - Lambda expression compilation
- `compileFunctionCall()` - Call site compilation
- Parameter handling
- Closure capture
- Nested function compilation

**Migration Notes:**
- Create `internal/bytecode/compiler/` directory
- Update VM and optimizer imports
- Consolidate compiler tests

---

#### 2.4 internal/parser/expressions.go (34KB, 1,232 lines)

**Current Content:** Expression parsing with Pratt parser - literals, identifiers, unary, binary, call, index, member access, type cast, lambda, etc.

**Possible Split** (Optional - may be fine as-is):

```
expressions.go (could split if desired) →
├── expressions/literals.go      // Literal parsing
├── expressions/operators.go     // Binary/unary operations
├── expressions/calls.go         // Call expressions
└── expressions/complex.go       // Lambda, type cast, etc.
```

**Recommendation:** This file may be acceptable as-is since it's a coherent Pratt parser. Consider splitting only if adding more expression types pushes it past 50KB.

---

### Priority 3: Medium Files (32-40KB)

The following files are moderately large but may not require immediate splitting:

- `internal/interp/builtins_core.go` (32KB) - Could be split if it grows
- `internal/interp/value.go` (34KB) - 16+ type definitions, natural size
- `internal/interp/expressions.go` (38KB) - Expression evaluation dispatch
- `internal/semantic/analyze_builtin_math.go` (36KB) - Mirrors interp version
- `internal/semantic/analyze_builtin_datetime.go` (37KB) - Mirrors interp version
- `internal/types/types.go` (36KB) - Type system definitions

**Recommendation:** Monitor these files. If they grow beyond 50KB or become difficult to navigate, apply similar splitting strategies.

---

## Subdirectory Organization Strategy

### Overview

Many directories have become cluttered with 60-90+ files. Organizing them into subdirectories will improve navigability and maintainability.

**Key Principles:**
1. **Feature-based grouping** - Group related functionality together
2. **Parallel structures** - Mirror organization between `interp/` and `semantic/`
3. **Test co-location** - Move test files to `tests/` subdirectories
4. **Shallow hierarchies** - Prefer 2-3 levels max, avoid over-nesting

---

### Strategy 1: internal/interp/ (Currently 97 files)

**Current State:**
- 97 total files (34 implementation, 63 test files)
- Flat structure with all files in root
- Builtin functions spread across multiple files
- OOP features scattered
- Tests mixed with implementation

**Proposed Structure:**

```
internal/interp/
├── builtins/                    # All builtin function implementations
│   ├── core.go                  # Core builtins (Print, Length, etc.)
│   ├── math_basic.go            # Basic math functions
│   ├── math_trig.go             # Trigonometric functions
│   ├── math_convert.go          # Rounding/conversion functions
│   ├── datetime_format.go       # Date/time formatting
│   ├── datetime_calc.go         # Date/time calculations
│   ├── datetime_info.go         # Date/time information
│   ├── strings.go               # String manipulation
│   ├── arrays.go                # Array builtins
│   ├── json.go                  # JSON functions
│   ├── variant.go               # Variant functions
│   ├── ordinals.go              # Ord, Chr, Succ, Pred
│   ├── collections.go           # Collection helpers
│   └── registry.go              # Builtin function registry
├── objects/                     # Object-oriented features
│   ├── instantiation.go         # Object creation
│   ├── properties.go            # Property access
│   ├── methods.go               # Method calls
│   └── hierarchy.go             # Inheritance/lookup
├── functions/                   # Function call handling
│   ├── calls.go                 # Call dispatch
│   ├── builtins.go              # Builtin call handling
│   ├── user.go                  # User function calls
│   ├── pointers.go              # Function pointers/lambdas
│   ├── records.go               # Record method calls
│   └── typecast.go              # Type casting
├── statements/                  # Statement evaluation
│   ├── declarations.go          # var, const, type declarations
│   ├── assignments.go           # Assignment statements
│   ├── control.go               # if, case statements
│   └── loops.go                 # while, repeat, for loops
├── values/                      # Value type definitions
│   ├── types.go                 # ValueType enum, Value interface
│   ├── primitives.go            # IntValue, FloatValue, StringValue, BoolValue
│   ├── array.go                 # ArrayValue
│   ├── record.go                # RecordValue
│   ├── set.go                   # SetValue
│   ├── enum.go                  # EnumValue
│   ├── class.go                 # ClassValue
│   ├── function.go              # FunctionValue
│   ├── variant.go               # VariantValue
│   └── helpers.go               # Value conversion helpers
├── tests/                       # All test files
│   ├── integration/
│   │   ├── fixture_test.go      # Fixture test runner
│   │   └── interpreter_test.go  # Full interpreter tests
│   ├── builtins/
│   │   ├── math_test.go
│   │   ├── string_test.go
│   │   ├── datetime_test.go
│   │   ├── json_test.go
│   │   ├── variant_test.go
│   │   └── ...
│   ├── objects/
│   │   ├── class_test.go
│   │   ├── property_test.go
│   │   ├── method_test.go
│   │   ├── inheritance_test.go
│   │   └── ...
│   ├── statements/
│   │   ├── assignment_test.go
│   │   ├── control_flow_test.go
│   │   ├── loops_test.go
│   │   └── ...
│   ├── types/
│   │   ├── array_test.go
│   │   ├── record_test.go
│   │   ├── set_test.go
│   │   └── ...
│   └── ...
├── interpreter.go               # Main interpreter struct
├── expressions.go               # Expression evaluation dispatch
├── helpers.go                   # Helper functions
├── environment.go               # Environment/scope management
├── exceptions.go                # Exception handling
├── unit_loader.go               # Unit loading
└── doc.go                       # Package documentation
```

**Benefits:**
- Reduces root directory from 97 to ~15 files
- Clear separation of concerns
- Easy to find related functionality
- Tests organized by feature area
- Builtin functions grouped logically

**Migration Steps:**
1. Create subdirectories: `builtins/`, `objects/`, `functions/`, `statements/`, `values/`, `tests/`
2. Move/split large files according to Priority 1 plan
3. Move remaining builtin files to `builtins/`
4. Move value type files to `values/`
5. Move all `*_test.go` files to appropriate `tests/` subdirectories
6. Update imports throughout codebase
7. Update `CLAUDE.md` with new structure
8. Run full test suite to verify

---

### Strategy 2: internal/semantic/ (Currently 88 files)

**Current State:**
- 88 total files (40 implementation, 48 test files)
- Similar issues to `interp/`
- Should mirror `interp/` structure for consistency

**Proposed Structure:**

```
internal/semantic/
├── builtins/                    # Builtin function type checking
│   ├── math.go                  # Math builtin signatures
│   ├── datetime.go              # DateTime builtin signatures
│   ├── string.go                # String builtin signatures
│   ├── array.go                 # Array builtin signatures
│   ├── json.go                  # JSON builtin signatures
│   ├── variant.go               # Variant builtin signatures
│   ├── convert.go               # Conversion builtin signatures
│   └── functions.go             # Builtin function signature registry
├── classes/                     # Class analysis
│   ├── decl.go                  # Class declaration analysis
│   ├── methods.go               # Method analysis
│   ├── properties.go            # Property analysis
│   └── constructors.go          # Constructor analysis
├── functions/                   # Function analysis
│   ├── calls.go                 # Call expression analysis
│   ├── pointers.go              # Function pointer analysis
│   ├── lambdas.go               # Lambda analysis
│   └── overloads.go             # Overload resolution
├── tests/                       # All test files
│   ├── builtins/
│   ├── classes/
│   ├── functions/
│   ├── types/
│   └── ...
├── analyzer.go                  # Main analyzer struct
├── type_resolution.go           # Type resolution logic
├── symbol_table.go              # Symbol table management
├── errors.go                    # Semantic error reporting
├── analyze_expressions.go       # Expression analysis
├── analyze_statements.go        # Statement analysis
├── analyze_types.go             # Type declaration analysis
└── doc.go                       # Package documentation
```

**Benefits:**
- Mirrors `internal/interp/` structure
- Easy to find corresponding analysis code for interpreter features
- Builtin analysis grouped together
- Tests organized by feature

**Migration Steps:**
1. Create subdirectories: `builtins/`, `classes/`, `functions/`, `tests/`
2. Split `analyze_classes.go` according to Priority 2 plan
3. Move/organize builtin analysis files to `builtins/`
4. Move function analysis files to `functions/`
5. Move all test files to `tests/` subdirectories
6. Update imports
7. Run tests

---

### Strategy 3: internal/parser/ (Currently 61 files)

**Current State:**
- 61 total files (21 implementation, 40 test files)
- Parsing logic is naturally organized by language feature
- Main issue is test file clutter

**Proposed Structure:**

```
internal/parser/
├── tests/                       # All test files (move 40 files here)
│   ├── expressions/
│   │   ├── literals_test.go
│   │   ├── operators_test.go
│   │   ├── calls_test.go
│   │   └── ...
│   ├── statements/
│   │   ├── declarations_test.go
│   │   ├── control_flow_test.go
│   │   └── ...
│   ├── types/
│   │   ├── arrays_test.go
│   │   ├── records_test.go
│   │   └── ...
│   └── ...
├── parser.go                    # Core parser struct
├── expressions.go               # Expression parsing (may split)
├── statements.go                # Statement parsing
├── functions.go                 # Function declaration parsing
├── classes.go                   # Class declaration parsing
├── interfaces.go                # Interface declaration parsing
├── types.go                     # Type declaration parsing
├── arrays.go                    # Array type parsing
├── records.go                   # Record type parsing
├── enums.go                     # Enum type parsing
├── sets.go                      # Set type parsing
├── control_flow.go              # Control flow statement parsing
├── properties.go                # Property declaration parsing
├── exceptions.go                # Exception handling parsing
├── operators.go                 # Operator precedence
├── helpers.go                   # Parser helper functions
└── doc.go                       # Package documentation
```

**Benefits:**
- Keeps parsing logic in root (less nesting, easier to navigate)
- Moves 40 test files to organized subdirectory
- Simple, flat structure for implementation

**Recommendation:**
- Parser structure is fairly clean already
- Main benefit is moving test files
- Consider splitting `expressions.go` only if it grows significantly

**Migration Steps:**
1. Create `tests/` subdirectory with feature subdirectories
2. Move all `*_test.go` files to appropriate locations
3. Split `expressions.go` if desired
4. Update imports
5. Run tests

---

### Strategy 4: internal/bytecode/ (Currently 16 files)

**Current State:**
- 16 total files (7 implementation, 9 test files)
- Already fairly well organized
- Main candidates: `vm.go` and `compiler.go`

**Proposed Structure:**

```
internal/bytecode/
├── vm/                          # Virtual machine components
│   ├── core.go                  # VM struct, Run loop
│   ├── stack.go                 # Stack operations
│   ├── builtins.go              # Builtin implementations
│   └── exceptions.go            # Exception handling
├── compiler/                    # Compiler components
│   ├── core.go                  # Compiler struct
│   ├── statements.go            # Statement compilation
│   ├── expressions.go           # Expression compilation
│   └── functions.go             # Function compilation
├── tests/                       # Test files
│   ├── vm_test.go
│   ├── compiler_test.go
│   ├── optimizer_test.go
│   └── ...
├── bytecode.go                  # Bytecode format, Value types
├── instruction.go               # Instruction definitions
├── optimizer.go                 # Bytecode optimizer
├── disasm.go                    # Disassembler
└── doc.go                       # Package documentation
```

**Benefits:**
- Clear VM vs Compiler separation
- Tests organized
- Room for future growth (JIT, more optimizations)

**Migration Steps:**
1. Create `vm/`, `compiler/`, `tests/` subdirectories
2. Split `vm.go` and `compiler.go` according to Priority 2 plan
3. Move test files to `tests/`
4. Update imports
5. Run tests

---

### Strategy 5: internal/types/ (Currently ~25 files)

**Current State:**
- Core type system definitions
- Reasonably well organized already
- No immediate restructuring needed

**Recommendation:**
- Keep current structure
- Monitor file sizes
- Consider subdirectories if package grows significantly (e.g., `types/classes/`, `types/builtins/`)

---

### Strategy 6: internal/lexer/ (Currently ~10 files)

**Current State:**
- Small, focused package
- Well organized

**Recommendation:**
- No changes needed
- Keep as-is

---

## Implementation Priorities

### Phase 1: Critical File Splits (Week 1-2)

**Goal:** Split the largest, most unwieldy files

1. ✅ Split `internal/interp/objects.go` (79KB → 4 files)
2. ✅ Split `internal/interp/functions.go` (67KB → 6 files)
3. ✅ Split `internal/interp/statements.go` (52KB → 4 files)
4. ✅ Run full test suite after each split
5. ✅ Update imports as needed

**Deliverable:** Three critical files split into logical components

---

### Phase 2: Builtin Organization (Week 2-3)

**Goal:** Organize builtin functions into dedicated subdirectory

1. ✅ Create `internal/interp/builtins/` directory
2. ✅ Split `builtins_datetime.go` → 3 files
3. ✅ Split `builtins_math.go` → 3 files
4. ✅ Move remaining builtin files to `builtins/`
5. ✅ Create builtin registry file
6. ✅ Update imports and tests

**Deliverable:** All builtin functions organized in `builtins/` subdirectory

---

### Phase 3: Semantic Package Refactor (Week 3-4)

**Goal:** Mirror interp organization in semantic package

1. ✅ Create `internal/semantic/builtins/` directory
2. ✅ Split `analyze_classes.go` (48KB → 4 files)
3. ✅ Organize builtin analysis files
4. ✅ Create `internal/semantic/classes/` directory
5. ✅ Create `internal/semantic/functions/` directory
6. ✅ Run tests

**Deliverable:** Semantic package organized to mirror interp structure

---

### Phase 4: Bytecode Package Refactor (Week 4-5)

**Goal:** Organize VM and compiler into subdirectories

1. ✅ Create `internal/bytecode/vm/` directory
2. ✅ Split `vm.go` (47KB → 4 files)
3. ✅ Create `internal/bytecode/compiler/` directory
4. ✅ Split `compiler.go` (42KB → 4 files)
5. ✅ Update imports
6. ✅ Run tests

**Deliverable:** Bytecode package with clear VM/compiler separation

---

### Phase 5: Test Organization (Week 5-6)

**Goal:** Move all test files to organized subdirectories

1. ✅ Create `internal/interp/tests/` with subdirectories
2. ✅ Move all interp test files
3. ✅ Create `internal/semantic/tests/` with subdirectories
4. ✅ Move all semantic test files
5. ✅ Create `internal/parser/tests/` with subdirectories
6. ✅ Move all parser test files
7. ✅ Create `internal/bytecode/tests/` with subdirectories
8. ✅ Move all bytecode test files
9. ✅ Update test imports
10. ✅ Run full test suite

**Deliverable:** All test files organized by feature area

---

### Phase 6: Value Types Organization (Week 6)

**Goal:** Organize value type definitions

1. ✅ Create `internal/interp/values/` directory
2. ✅ Move value type files
3. ✅ Update imports
4. ✅ Run tests

**Deliverable:** Value types in dedicated subdirectory

---

### Phase 7: Documentation Updates (Week 7)

**Goal:** Update all documentation to reflect new structure

1. ✅ Update `CLAUDE.md` with new directory structure
2. ✅ Update `README.md` with new package layout
3. ✅ Update `CONTRIBUTING.md` with refactoring guidelines
4. ✅ Add `doc.go` files to new subdirectories
5. ✅ Update architecture diagrams in `goal.md`

**Deliverable:** Complete, accurate documentation

---

## Migration Guidelines

### General Principles

1. **One change at a time** - Split one file or move one group at a time
2. **Test after every change** - Run full test suite after each modification
3. **Maintain git history** - Use `git mv` when moving files
4. **Update imports immediately** - Don't let broken imports accumulate
5. **Document as you go** - Update comments and docs alongside code changes

### File Splitting Process

For each file to be split:

1. **Create target directory** (if needed)
   ```bash
   mkdir -p internal/interp/objects
   ```

2. **Create new files** with appropriate headers
   ```go
   // Package objects contains object-oriented feature implementations.
   package objects

   import (
       "github.com/MeKo-Tech/go-dws/internal/ast"
       // ... other imports
   )
   ```

3. **Copy functions to new files** based on logical grouping

4. **Update package references**
   - Change `package interp` to `package objects` (or appropriate name)
   - Adjust visibility (exported vs unexported)

5. **Update imports in other files**
   ```go
   import (
       "github.com/MeKo-Tech/go-dws/internal/interp/objects"
   )
   ```

6. **Update function calls**
   ```go
   // Old: evalNewExpression(...)
   // New: objects.EvalNewExpression(...)
   ```

7. **Run tests**
   ```bash
   go test ./internal/interp/... -v
   ```

8. **Delete original file** only after all tests pass
   ```bash
   git rm internal/interp/objects.go
   ```

9. **Commit**
   ```bash
   git add .
   git commit -m "refactor: split objects.go into objects/ subdirectory"
   ```

### Import Path Updates

When creating subdirectories, import paths change:

**Before:**
```go
import "github.com/MeKo-Tech/go-dws/internal/interp"

interp.NewInterpreter()
```

**After:**
```go
import (
    "github.com/MeKo-Tech/go-dws/internal/interp"
    "github.com/MeKo-Tech/go-dws/internal/interp/builtins"
    "github.com/MeKo-Tech/go-dws/internal/interp/objects"
)

interp.NewInterpreter()
builtins.RegisterAll()
objects.Instantiate()
```

### Test File Organization

When moving test files:

1. **Preserve test package names**
   ```go
   // Can use either:
   package interp_test  // Black-box testing
   package interp       // White-box testing
   ```

2. **Update relative paths** in test data
   ```go
   // Old: testdata/simple.dws
   // New: ../../testdata/simple.dws (if tests moved to tests/ subdirectory)
   ```

3. **Consider test helper consolidation**
   - Create shared test helpers in `tests/helpers.go`
   - Reduce duplication across test files

### Visibility Considerations

When splitting files, consider function visibility:

**Keep unexported** (lowercase) if:
- Function is only used within the package
- Function is an implementation detail

**Make exported** (uppercase) if:
- Function needs to be called from other packages
- Function is part of public API

**Example:**
```go
// objects/instantiation.go

// EvalNewExpression evaluates object instantiation (exported - called from interp)
func EvalNewExpression(node *ast.NewExpression, interp *Interpreter) Value {
    return createObjectInstance(node, interp)
}

// createObjectInstance is an internal helper (unexported)
func createObjectInstance(node *ast.NewExpression, interp *Interpreter) Value {
    // ...
}
```

### Common Pitfalls to Avoid

❌ **Don't split arbitrarily** - Split based on logical boundaries
❌ **Don't break circular dependencies** - Be mindful of package dependencies
❌ **Don't skip tests** - Always run tests after changes
❌ **Don't batch too many changes** - Commit frequently
❌ **Don't forget documentation** - Update docs alongside code

✅ **Do split by feature** - Group related functions
✅ **Do maintain package cohesion** - Keep related code together
✅ **Do test incrementally** - Run tests after each change
✅ **Do commit frequently** - Small, focused commits
✅ **Do update documentation** - Keep docs in sync

---

## Testing Strategy

### After Each File Split

1. **Unit tests** - Run package tests
   ```bash
   go test ./internal/interp/objects -v
   ```

2. **Integration tests** - Run interpreter tests
   ```bash
   go test ./internal/interp -v -run TestDWScriptFixtures
   ```

3. **Full suite** - Run all tests
   ```bash
   go test ./... -v
   ```

4. **Coverage check** - Ensure coverage doesn't drop
   ```bash
   go test -cover ./internal/interp/...
   ```

### After Directory Reorganization

1. **Import checks** - Verify no broken imports
   ```bash
   go build ./...
   ```

2. **Test discovery** - Ensure all tests still run
   ```bash
   go test ./... -v | grep -c "PASS"
   ```

3. **Fixture tests** - Run complete DWScript test suite
   ```bash
   go test ./internal/interp -run TestDWScriptFixtures -v
   ```

### Regression Prevention

1. **Run tests before and after** - Compare results
2. **Check test coverage** - Should remain stable or increase
3. **Verify fixture test pass rate** - Should not decrease
4. **Test CLI tool** - Ensure `dwscript` commands still work

---

## Expected Outcomes

### Metrics Before Refactoring

- **Files over 50KB:** 6 files
- **Files over 30KB:** 20 files
- **Total files in internal/interp:** 97
- **Total files in internal/semantic:** 88
- **Total files in internal/parser:** 61
- **Total files in internal/bytecode:** 16
- **Average file size:** ~15KB
- **Largest file:** 79KB (objects.go)

### Metrics After Refactoring

- **Files over 50KB:** 0 files (target)
- **Files over 30KB:** <5 files (target)
- **Root files in internal/interp:** ~15 (from 97)
- **Root files in internal/semantic:** ~15 (from 88)
- **Root files in internal/parser:** ~20 (from 61)
- **Root files in internal/bytecode:** ~10 (from 16)
- **Average file size:** ~8-10KB (target)
- **Largest file:** <40KB (target)

### Qualitative Improvements

✅ **Improved navigation** - Easier to find related code
✅ **Better organization** - Clear feature groupings
✅ **Reduced cognitive load** - Smaller, focused files
✅ **Parallel structures** - Consistent organization across packages
✅ **Test organization** - Tests grouped by feature
✅ **Easier onboarding** - New contributors can navigate more easily
✅ **Better maintainability** - Changes localized to specific areas
✅ **Clearer dependencies** - Package relationships more explicit

---

## Future Considerations

### As the Project Grows

1. **Monitor file sizes** - Split files before they exceed 50KB
2. **Consistent patterns** - Apply same organization to new packages
3. **Regular refactoring** - Don't let technical debt accumulate
4. **Documentation** - Keep CLAUDE.md updated with structure changes

### Potential Future Subdirectories

If packages continue to grow, consider:

- `internal/interp/vm/` - Alternative VM implementation
- `internal/interp/optimizer/` - AST optimization passes
- `internal/semantic/inference/` - Type inference
- `internal/codegen/` - Code generation (if transpilation added)
- `pkg/stdlib/` - Standard library modules

### WebAssembly Considerations (Stage 10.15)

The planned `pkg/platform/` and `pkg/wasm/` packages should follow these same organizational principles:

```
pkg/
├── platform/
│   ├── filesystem/
│   ├── console/
│   └── runtime/
└── wasm/
    ├── bindings/
    ├── interop/
    └── stdlib/
```

---

## Conclusion

This refactoring plan provides a comprehensive strategy for organizing the go-dws codebase. By splitting large files into logical components and organizing code into feature-based subdirectories, we'll create a more maintainable, navigable, and scalable codebase.

**Key Takeaways:**

1. **Logical splits** - Based on functionality, not arbitrary size limits
2. **Parallel organization** - Mirror structure between related packages
3. **Incremental approach** - Phase-by-phase implementation with testing
4. **Clear priorities** - Focus on highest-impact changes first
5. **Comprehensive testing** - Verify functionality after every change

This refactoring will position the project well for future growth and make it easier for contributors to understand and modify the codebase.

---

**Document Version:** 1.0
**Last Updated:** 2025-11-10
**Status:** Ready for Implementation
