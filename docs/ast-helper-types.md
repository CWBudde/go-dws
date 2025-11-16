# AST Helper Types Audit

**Date**: 2025-11-16
**Task**: 9.20.1 - Audit AST for non-Node types used in traversal

## Overview

This document catalogs all helper types in the AST that are used in node traversal but do not fully implement the Node interface. Making these types implement Node will improve type safety, enable consistent visitor traversal, and provide position information for better error messages.

## Helper Types Summary

| Type | File | Has Token/EndPos | Has String() | Has Node Methods | Status |
|------|------|------------------|--------------|------------------|--------|
| Parameter | functions.go | Partial (fields only) | ✓ | ✗ | Needs Node impl |
| CaseBranch | control_flow.go | ✗ | ✓ | ✗ | Needs full impl |
| ExceptClause | exceptions.go | ✗ | ✓ | ✗ | Needs full impl |
| ExceptionHandler | exceptions.go | ✗ | ✓ | ✗ | Needs full impl |
| FinallyClause | exceptions.go | ✓ (BaseNode) | ✓ | Partial | Needs marker only |
| FieldInitializer | records.go | ✓ (BaseNode) | ✓ | Partial | Needs marker only |
| InterfaceMethodDecl | interfaces.go | ✓ (BaseNode) | ✓ | Partial | Needs marker only |

## Detailed Analysis

### 1. Parameter (pkg/ast/functions.go)

**Current Status**: Has Token and EndPos fields, but does not implement Node interface methods.

**Fields**:
- `Token token.Token` - Present
- `EndPos token.Position` - Present
- `Name *Identifier`
- `Type *TypeAnnotation`
- `DefaultValue Expression`
- `IsLazy bool`
- `ByRef bool`
- `IsConst bool`

**What's Missing**:
- `TokenLiteral() string` method
- `Pos() token.Position` method
- `End() token.Position` method
- `statementNode()` marker method (parameters are like declarations)

**Used In**: FunctionDecl.Parameters, InterfaceMethodDecl.Parameters

**Current Visitor Handling**: Dedicated `walkParameter()` function (not in Walk() switch)

**Implementation Strategy**: Add the missing methods. Can reuse logic from BaseNode.

---

### 2. CaseBranch (pkg/ast/control_flow.go)

**Current Status**: Minimal struct with no Node interface support.

**Fields**:
- `Values []Expression`
- `Statement Statement`

**What's Missing**:
- `Token token.Token` field
- `EndPos token.Position` field
- All Node interface methods
- `statementNode()` marker method

**Used In**: CaseStatement.Cases

**Current Visitor Handling**: Dedicated `walkCaseBranch()` function (not in Walk() switch)

**Implementation Strategy**: Add Token/EndPos fields, implement Node interface. Token should be the first value token or case keyword.

---

### 3. ExceptClause (pkg/ast/exceptions.go)

**Current Status**: Helper struct with no Node interface support.

**Fields**:
- `Handlers []*ExceptionHandler`
- `ElseBlock *BlockStatement`

**What's Missing**:
- `Token token.Token` field
- `EndPos token.Position` field
- All Node interface methods
- `statementNode()` marker method

**Used In**: TryStatement.ExceptClause

**Current Visitor Handling**: Dedicated `walkExceptClause()` function (not in Walk() switch)

**Implementation Strategy**: Add Token/EndPos fields, implement Node interface. Token should be the 'except' keyword.

---

### 4. ExceptionHandler (pkg/ast/exceptions.go)

**Current Status**: Helper struct with no Node interface support.

**Fields**:
- `Variable *Identifier`
- `ExceptionType *TypeAnnotation`
- `Statement Statement`

**What's Missing**:
- `Token token.Token` field
- `EndPos token.Position` field
- All Node interface methods
- `statementNode()` marker method

**Used In**: ExceptClause.Handlers

**Current Visitor Handling**: Dedicated `walkExceptionHandler()` function (not in Walk() switch)

**Implementation Strategy**: Add Token/EndPos fields, implement Node interface. Token should be the 'on' keyword.

---

### 5. FinallyClause (pkg/ast/exceptions.go)

**Current Status**: Already has BaseNode embedded and String() method. Almost complete!

**Fields**:
- `BaseNode` - Provides Token, EndPos, TokenLiteral(), Pos(), End()
- `Block *BlockStatement`

**What's Missing**:
- `statementNode()` marker method only

**Used In**: TryStatement.FinallyClause

**Current Visitor Handling**: In Walk() switch, calls `walkFinallyClause()`

**Implementation Strategy**: Simply add `func (fc *FinallyClause) statementNode() {}` - that's it!

---

### 6. FieldInitializer (pkg/ast/records.go)

**Current Status**: Already has BaseNode embedded and String() method. Almost complete!

**Fields**:
- `BaseNode` - Provides Token, EndPos, TokenLiteral(), Pos(), End()
- `Name *Identifier`
- `Value Expression`

**What's Missing**:
- Marker method - should be `statementNode()` (it's like a mini assignment)

**Used In**: RecordLiteralExpression.Fields

**Current Visitor Handling**: In Walk() switch, calls `walkFieldInitializer()`

**Implementation Strategy**: Simply add `func (fi *FieldInitializer) statementNode() {}` - that's it!

---

### 7. InterfaceMethodDecl (pkg/ast/interfaces.go)

**Current Status**: Already has BaseNode embedded, String() method, and custom End() override. Almost complete!

**Fields**:
- `BaseNode` - Provides Token, EndPos, TokenLiteral(), Pos()
- `Name *Identifier`
- `ReturnType *TypeAnnotation`
- `Parameters []*Parameter`

**What's Missing**:
- `statementNode()` marker method only

**Used In**: InterfaceDecl.Methods

**Current Visitor Handling**: In Walk() switch, calls `walkInterfaceMethodDecl()`

**Implementation Strategy**: Simply add `func (imd *InterfaceMethodDecl) statementNode() {}` - that's it!

---

## Implementation Priority

### Phase 1: Quick Wins (add marker methods only)
1. FinallyClause - add `statementNode()`
2. FieldInitializer - add `statementNode()`
3. InterfaceMethodDecl - add `statementNode()`

### Phase 2: Implement Node Interface
4. Parameter - add Node methods (has Token/EndPos fields already)
5. CaseBranch - add Token/EndPos and Node methods
6. ExceptClause - add Token/EndPos and Node methods
7. ExceptionHandler - add Token/EndPos and Node methods

### Phase 3: Update Visitor
8. Move helper walkXXX() functions into the main Walk() switch
9. Remove manual field walking code
10. Simplify visitor implementation

### Phase 4: Update Parser
11. Ensure parser sets Token/EndPos when creating helper types
12. Verify position information is accurate

### Phase 5: Testing
13. Test visitor traversal includes all helper types
14. Verify position information in error messages

## Benefits of This Work

1. **Type Safety**: Can call Walk() on helper types just like any other node
2. **Position Info**: All traversable types can report their source location
3. **Cleaner Code**: No special-case handling in visitor
4. **Better Errors**: Error messages can point to exact helper type locations
5. **Consistency**: All AST types follow the same pattern

## Current Visitor Implementation

The visitor currently has two modes of handling types:

1. **In Walk() switch**: Types that implement Node (or are close to it)
   - FinallyClause
   - FieldInitializer
   - InterfaceMethodDecl

2. **Dedicated walkXXX() functions**: Helper types without Node interface
   - Parameter (walkParameter)
   - CaseBranch (walkCaseBranch)
   - ExceptClause (walkExceptClause)
   - ExceptionHandler (walkExceptionHandler)

After this task, all types will be in category #1, making the visitor implementation cleaner and more consistent.
