package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ContractPass implements Pass 4: Contract Validation
//
// **Purpose**: Validate Design-by-Contract constructs (requires, ensures, invariant)
// now that all types and expressions have been validated. This is a specialized pass
// for contract-based programming features.
//
// **Responsibilities**:
// - Validate 'require' (precondition) clauses in functions
//   - Check that expressions are boolean
//   - Validate that only parameters and constants are referenced
//
// - Validate 'ensure' (postcondition) clauses in functions
//   - Check that expressions are boolean
//   - Validate 'old(expr)' expressions (save value at function entry)
//   - Check that old() only wraps valid expressions (no side effects)
//
// - Validate 'invariant' clauses in classes
//   - Check that expressions are boolean
//   - Validate that invariants only reference fields and constants
//   - Check that invariants are maintained after constructors/destructors
//
// - Validate 'assert' statements
//   - Check that expressions are boolean
//
// - Check contract inheritance rules
//   - Derived methods can weaken preconditions (OR with parent)
//   - Derived methods can strengthen postconditions (AND with parent)
//   - Class invariants are inherited
//
// **What it does NOT do**:
// - Type-check general expressions (already done in Pass 3)
// - Validate non-contract statements
//
// **Dependencies**: Pass 3 (Semantic Validation)
//
// **Inputs**:
// - Fully type-checked AST with semantic annotations
// - Complete type system with inheritance
//
// **Outputs**:
// - Errors for invalid contract expressions
// - Warnings for contract violations (if detectable at compile-time)
//
// **Example**:
//
//	function Divide(x, y: Integer): Integer;
//	require
//	  y <> 0  // Precondition: divisor must be non-zero
//	ensure
//	  Result * y = x  // Postcondition: result * divisor = dividend
//	begin
//	  Result := x div y;
//	end;
//
//	class TStack = class
//	private
//	  FCount: Integer;
//	invariant
//	  FCount >= 0  // Class invariant: count is never negative
//	public
//	  procedure Push(Item: Integer);
//	  function Pop: Integer;
//	end;
type ContractPass struct{}

// NewContractPass creates a new contract validation pass.
func NewContractPass() *ContractPass {
	return &ContractPass{}
}

// Name returns the name of this pass.
func (p *ContractPass) Name() string {
	return "Pass 4: Contract Validation"
}

// Run executes the contract validation pass.
func (p *ContractPass) Run(program *ast.Program, ctx *PassContext) error {
	validator := &contractValidator{
		ctx:     ctx,
		program: program,
	}

	// Walk all statements and validate contracts
	for _, stmt := range program.Statements {
		validator.validateStatement(stmt)
	}

	return nil
}

// contractValidator validates contract clauses (requires, ensures, invariant)
type contractValidator struct {
	ctx     *PassContext
	program *ast.Program
}

// validateStatement validates contracts in a single statement
func (v *contractValidator) validateStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.FunctionDecl:
		v.validateFunctionContracts(s)
	case *ast.ClassDecl:
		// Walk through all methods in the class
		for _, method := range s.Methods {
			v.validateFunctionContracts(method)
		}
		// TODO: Validate class invariants when they are added to the AST
	case *ast.BlockStatement:
		// Recursively validate nested statements
		for _, nested := range s.Statements {
			v.validateStatement(nested)
		}
	}
}

// validateFunctionContracts validates require and ensure clauses in a function
func (v *contractValidator) validateFunctionContracts(fn *ast.FunctionDecl) {
	if fn == nil {
		return
	}

	// Save current function context
	oldFunction := v.ctx.CurrentFunction
	v.ctx.CurrentFunction = fn
	defer func() {
		v.ctx.CurrentFunction = oldFunction
	}()

	// Validate preconditions (require)
	if fn.PreConditions != nil {
		v.validatePreConditions(fn.PreConditions, fn)
	}

	// Validate postconditions (ensure)
	if fn.PostConditions != nil {
		v.validatePostConditions(fn.PostConditions, fn)
	}
}

// validatePreConditions validates precondition clauses
func (v *contractValidator) validatePreConditions(pre *ast.PreConditions, fn *ast.FunctionDecl) {
	if pre == nil {
		return
	}

	for _, cond := range pre.Conditions {
		v.validateCondition(cond, "precondition", fn, false)
	}
}

// validatePostConditions validates postcondition clauses
func (v *contractValidator) validatePostConditions(post *ast.PostConditions, fn *ast.FunctionDecl) {
	if post == nil {
		return
	}

	for _, cond := range post.Conditions {
		v.validateCondition(cond, "postcondition", fn, true)
	}
}

// validateCondition validates a single contract condition
func (v *contractValidator) validateCondition(cond *ast.Condition, kind string, fn *ast.FunctionDecl, allowOld bool) {
	if cond == nil || cond.Test == nil {
		return
	}

	// Check that the test expression is boolean
	testType := v.checkExpression(cond.Test, fn, allowOld)
	if testType != nil && testType.String() != "Boolean" {
		v.ctx.AddError("%s must be boolean, got %s", kind, testType)
	}

	// Validate message if present (should be string, but parser already enforces this)
	if cond.Message != nil {
		msgType := v.checkExpression(cond.Message, fn, allowOld)
		if msgType != nil && msgType.String() != "String" {
			v.ctx.AddError("%s message must be string, got %s", kind, msgType)
		}
	}
}

// checkExpression type-checks an expression and validates old expressions
func (v *contractValidator) checkExpression(expr ast.Expression, fn *ast.FunctionDecl, allowOld bool) types.Type {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.OldExpression:
		return v.checkOldExpression(e, fn, allowOld)
	case *ast.BinaryExpression:
		leftType := v.checkExpression(e.Left, fn, allowOld)
		rightType := v.checkExpression(e.Right, fn, allowOld)

		// Comparison and logical operators return Boolean
		switch e.Operator {
		case "=", "<>", "<", "<=", ">", ">=", "and", "or", "xor":
			typ, _ := v.ctx.TypeRegistry.Resolve("Boolean")
			return typ
		default:
			// For arithmetic operators, return the left type
			if leftType != nil {
				return leftType
			}
			return rightType
		}
	case *ast.UnaryExpression:
		return v.checkExpression(e.Right, fn, allowOld)
	case *ast.Identifier:
		return v.checkIdentifier(e, fn)
	case *ast.IntegerLiteral:
		typ, _ := v.ctx.TypeRegistry.Resolve("Integer")
		return typ
	case *ast.FloatLiteral:
		typ, _ := v.ctx.TypeRegistry.Resolve("Float")
		return typ
	case *ast.StringLiteral:
		typ, _ := v.ctx.TypeRegistry.Resolve("String")
		return typ
	case *ast.BooleanLiteral:
		typ, _ := v.ctx.TypeRegistry.Resolve("Boolean")
		return typ
	case *ast.CallExpression:
		// Check function existence and return type
		if ident, ok := e.Function.(*ast.Identifier); ok {
			sym, _ := v.ctx.Symbols.Resolve(ident.Value)
			if sym != nil && sym.Type != nil {
				return sym.Type
			}
		}
		return nil
	case *ast.MemberAccessExpression:
		// For member access, check the target and member
		v.checkExpression(e.Object, fn, allowOld)
		// Return type would be the member's type (simplified)
		return nil
	default:
		// For other expressions, just return nil (Pass 3 already validated them)
		return nil
	}
}

// checkOldExpression validates an old expression
func (v *contractValidator) checkOldExpression(old *ast.OldExpression, fn *ast.FunctionDecl, allowOld bool) types.Type {
	if old == nil {
		return nil
	}

	// Check if old expressions are allowed in this context
	if !allowOld {
		v.ctx.AddError("old() expressions are only allowed in postconditions (ensure)")
		return nil
	}

	// Check that the referenced identifier exists
	if old.Identifier == nil {
		v.ctx.AddError("old() expression missing identifier")
		return nil
	}

	// Validate that the identifier exists in the function's scope
	idType := v.checkIdentifier(old.Identifier, fn)
	if idType == nil {
		v.ctx.AddError("old() references undefined identifier '%s'", old.Identifier.Value)
		return nil
	}

	return idType
}

// checkIdentifier checks if an identifier is defined and returns its type
func (v *contractValidator) checkIdentifier(ident *ast.Identifier, fn *ast.FunctionDecl) types.Type {
	if ident == nil {
		return nil
	}

	// Check if it's a parameter
	if fn != nil {
		for _, param := range fn.Parameters {
			if param.Name != nil && strings.EqualFold(param.Name.Value, ident.Value) {
				// Resolve parameter type
				if param.Type != nil {
					return v.resolveTypeExpression(param.Type)
				}
			}
		}

		// Check if it's 'Result' (function return value) - case-insensitive
		if strings.EqualFold(ident.Value, "Result") && fn.ReturnType != nil {
			return v.resolveTypeExpression(fn.ReturnType)
		}
	}

	// Check symbol table
	sym, _ := v.ctx.Symbols.Resolve(ident.Value)
	if sym != nil {
		return sym.Type
	}

	return nil
}

// resolveTypeExpression resolves a type expression to a types.Type
func (v *contractValidator) resolveTypeExpression(typeExpr ast.TypeExpression) types.Type {
	if typeExpr == nil {
		return nil
	}

	switch t := typeExpr.(type) {
	case *ast.TypeAnnotation:
		if t.Name != "" {
			typ, _ := v.ctx.TypeRegistry.Resolve(t.Name)
			return typ
		}
	}

	return nil
}
