package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

func (i *Interpreter) fullClassNameFromDecl(cd *ast.ClassDecl) string {
	if cd.EnclosingClass != nil && cd.EnclosingClass.Value != "" {
		return cd.EnclosingClass.Value + "." + cd.Name.Value
	}
	return cd.Name.Value
}

// evalFunctionDeclaration evaluates a function declaration.
// It registers the function in the function registry without executing its body.
// For method implementations (fn.ClassName != nil), it updates the class's or record's Methods map.
func (i *Interpreter) evalFunctionDeclaration(fn *ast.FunctionDecl) Value {
	// Check if this is a method implementation (has a class/record name like TExample.Method)
	if fn.ClassName != nil {
		typeName := fn.ClassName.Value

		// Task 9.14.2 & PR #147: DWScript is case-insensitive
		// Use normalized key for O(1) lookup instead of O(n) linear search
		classInfo, isClass := i.classes[ident.Normalize(typeName)]

		if isClass {
			// Handle class method implementation
			i.evalClassMethodImplementation(fn, classInfo)
			return &NilValue{}
		}

		// PR #147: Use normalized key for O(1) lookup instead of O(n) linear search
		recordInfo, isRecord := i.records[ident.Normalize(typeName)]

		if isRecord {
			// Handle record method implementation
			i.evalRecordMethodImplementation(fn, recordInfo)
			return &NilValue{}
		}

		return i.newErrorWithLocation(fn, "type '%s' not found for method '%s'", typeName, fn.Name.Value)
	}

	// Task 3.5.7: Register global function in both legacy map and TypeSystem
	// The legacy map (i.functions) is kept for backward compatibility during migration.
	// TypeSystem.RegisterFunctionOrReplace handles the interface/implementation pattern:
	// - Forward declarations (no body) are appended
	// - Implementations (with body) replace matching declarations by signature

	// Register in TypeSystem (new path, used by Evaluator)
	i.typeSystem.RegisterFunctionOrReplace(fn.Name.Value, fn)

	// Also maintain legacy registry for backward compatibility
	// DWScript is case-insensitive, so normalize the function name
	funcName := ident.Normalize(fn.Name.Value)

	// If this function has a body, it may be an implementation that should
	// replace a previous interface declaration (which has no body).
	// This happens when units have separate interface and implementation sections.
	if fn.Body != nil {
		// Replace any existing declaration without a body
		existingOverloads := i.functions[funcName]
		i.functions[funcName] = i.replaceMethodInOverloadList(existingOverloads, fn)
	} else {
		// Interface declaration - append it
		i.functions[funcName] = append(i.functions[funcName], fn)
	}

	return &NilValue{}
}

// evalClassMethodImplementation handles class method implementation registration
func (i *Interpreter) evalClassMethodImplementation(fn *ast.FunctionDecl, classInfo *ClassInfo) {
	// Update the method in the class (replacing the declaration with the implementation)
	// Support method overloading by storing multiple methods per name
	// We need to replace the declaration with the implementation in the overload list
	normalizedMethodName := ident.Normalize(fn.Name.Value)

	if fn.IsClassMethod {
		classInfo.ClassMethods[normalizedMethodName] = fn
		// Replace declaration with implementation in overload list
		overloads := classInfo.ClassMethodOverloads[normalizedMethodName]
		classInfo.ClassMethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)
	} else {
		classInfo.Methods[normalizedMethodName] = fn
		// Replace declaration with implementation in overload list
		overloads := classInfo.MethodOverloads[normalizedMethodName]
		classInfo.MethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)
	}

	// Also store constructors
	if fn.IsConstructor {
		normalizedCtorName := ident.Normalize(fn.Name.Value)
		classInfo.Constructors[normalizedCtorName] = fn
		// Replace declaration with implementation in constructor overload list
		overloads := classInfo.ConstructorOverloads[normalizedCtorName]
		classInfo.ConstructorOverloads[normalizedCtorName] = i.replaceMethodInOverloadList(overloads, fn)
		// Always update Constructor to use the implementation (which has the body)
		// This replaces the declaration that was set during class parsing
		classInfo.Constructor = fn
	}

	// Store destructor
	if fn.IsDestructor {
		classInfo.Destructor = fn
	}

	// Task 9.14: Rebuild the VMT after adding method implementation
	// This ensures the VMT has references to methods with bodies, not just declarations
	classInfo.buildVirtualMethodTable()

	// PR #147 follow-up: Propagate implementation to descendants' method tables.
	// Child classes copy parent Methods/ClassMethods during declaration.
	// When a parent implementation arrives later (common for separate interface/implementation),
	// the child maps still point to the declaration (body=nil). Update them unless the child
	// overrides the method itself.
	i.propagateMethodImplementationToDescendants(classInfo, normalizedMethodName, fn, fn.IsClassMethod)

	// PR #147 Fix: Rebuild VMT for all descendant classes to propagate the change.
	// When a parent class method implementation is added after child classes exist,
	// child classes have stale VMT entries pointing to declaration-only methods.
	// We must rebuild their VMTs to get the new implementation.
	i.rebuildDescendantVMTs(classInfo)
}

// rebuildDescendantVMTs rebuilds the virtual method table for all classes that
// inherit from the given class. This is necessary when a parent class method
// implementation is added after child classes have been created.
// PR #147 Fix: Prevents child classes from keeping stale VMT entries.
func (i *Interpreter) rebuildDescendantVMTs(parentClass *ClassInfo) {
	// Iterate through all registered classes
	for _, classInfo := range i.classes {
		// Check if this class is a descendant of parentClass
		if i.isDescendantOf(classInfo, parentClass) {
			// Rebuild this descendant's VMT to pick up the new implementation
			classInfo.buildVirtualMethodTable()
		}
	}
}

// isDescendantOf checks if childClass is a descendant of ancestorClass.
// Returns true if childClass inherits from ancestorClass (directly or indirectly).
func (i *Interpreter) isDescendantOf(childClass, ancestorClass *ClassInfo) bool {
	// Walk up the parent chain from childClass
	current := childClass.Parent
	for current != nil {
		if current == ancestorClass {
			return true
		}
		current = current.Parent
	}
	return false
}

// propagateMethodImplementationToDescendants updates descendant classes' method maps
// to point to the latest implementation from an ancestor (when the descendant hasn't
// provided its own override). This fixes stale declaration-only entries in child
// Classes when parent method implementations are registered after the child class
// was already built.
func (i *Interpreter) propagateMethodImplementationToDescendants(parentClass *ClassInfo, normalizedMethodName string, fn *ast.FunctionDecl, isClassMethod bool) {
	for _, classInfo := range i.classes {
		if !i.isDescendantOf(classInfo, parentClass) {
			continue
		}

		if isClassMethod {
			if existing, ok := classInfo.ClassMethods[normalizedMethodName]; ok {
				// Skip if descendant defines/overrides its own method
				if existing.ClassName != nil && ident.Equal(existing.ClassName.Value, classInfo.Name) {
					continue
				}
				classInfo.ClassMethods[normalizedMethodName] = fn
			}
		} else {
			if existing, ok := classInfo.Methods[normalizedMethodName]; ok {
				// Skip if descendant defines/overrides its own method
				if existing.ClassName != nil && ident.Equal(existing.ClassName.Value, classInfo.Name) {
					continue
				}
				classInfo.Methods[normalizedMethodName] = fn
			}
		}
	}
}

// evalRecordMethodImplementation handles record method implementation registration
func (i *Interpreter) evalRecordMethodImplementation(fn *ast.FunctionDecl, recordInfo *RecordTypeValue) {
	// Update the method in the record (replacing the declaration with the implementation)
	// Support method overloading by storing multiple methods per name
	normalizedMethodName := ident.Normalize(fn.Name.Value)

	if fn.IsClassMethod {
		// Static method
		recordInfo.ClassMethods[normalizedMethodName] = fn
		// Replace declaration with implementation in overload list
		overloads := recordInfo.ClassMethodOverloads[normalizedMethodName]
		recordInfo.ClassMethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)
	} else {
		// Instance method
		recordInfo.Methods[normalizedMethodName] = fn
		// Replace declaration with implementation in overload list
		overloads := recordInfo.MethodOverloads[normalizedMethodName]
		recordInfo.MethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)
	}
}

// evalClassDeclaration evaluates a class declaration.
// It builds a ClassInfo from the AST and registers it in the class registry.
// Handles inheritance by copying parent fields and methods to the child class.
func (i *Interpreter) evalClassDeclaration(cd *ast.ClassDecl) Value {
	className := i.fullClassNameFromDecl(cd)

	// Check if this is a partial class declaration
	var classInfo *ClassInfo
	// PR #147: Use normalized key for O(1) case-insensitive lookup
	existingClass, exists := i.classes[ident.Normalize(className)]

	if exists && existingClass.IsPartial && cd.IsPartial {
		// Merging partial classes - reuse existing ClassInfo
		classInfo = existingClass
	} else if exists {
		// Non-partial class already exists - error (semantic analyzer should catch this)
		return i.newErrorWithLocation(cd, "class '%s' already declared", className)
	} else {
		// New class declaration
		classInfo = NewClassInfo(className)
	}

	// Mark as partial if this declaration is partial
	if cd.IsPartial {
		classInfo.IsPartial = true
		classInfo.Metadata.IsPartial = true // Task 3.5.39
	}

	// Set abstract flag (only if not already set)
	if cd.IsAbstract {
		classInfo.IsAbstract = true
		classInfo.Metadata.IsAbstract = true // Task 3.5.39
	}

	// Set external flags (only if not already set)
	if cd.IsExternal {
		classInfo.IsExternal = true
		classInfo.ExternalName = cd.ExternalName
		classInfo.Metadata.IsExternal = true              // Task 3.5.39
		classInfo.Metadata.ExternalName = cd.ExternalName // Task 3.5.39
	}

	// Provide current class context for nested type resolution during declaration processing
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)
	tempEnv.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: classInfo})
	i.env = tempEnv
	defer func() { i.env = savedEnv }()

	// Handle inheritance if parent class is specified
	var parentClass *ClassInfo
	if cd.Parent != nil {
		// Explicit parent specified
		// Task 9.14.2 & PR #147: DWScript is case-insensitive
		// Use normalized key for O(1) lookup instead of O(n) linear search
		parentName := cd.Parent.Value
		var exists bool
		parentClass, exists = i.classes[ident.Normalize(parentName)]
		if !exists {
			return i.newErrorWithLocation(cd, "parent class '%s' not found", parentName)
		}
	} else {
		// If no explicit parent, implicitly inherit from TObject
		// (unless this IS TObject or it's an external class)
		if !ident.Equal(className, "TObject") && !cd.IsExternal {
			// PR #147: Use normalized key for O(1) lookup instead of O(n) linear search
			var exists bool
			parentClass, exists = i.classes[ident.Normalize("TObject")]
			if !exists {
				return i.newErrorWithLocation(cd, "implicit parent class 'TObject' not found")
			}
		}
	}

	// Set parent reference and inherit members (only if not already set for partial classes)
	if parentClass != nil && classInfo.Parent == nil {
		classInfo.Parent = parentClass
		classInfo.Metadata.Parent = parentClass.Metadata          // Task 3.5.39
		classInfo.Metadata.ParentName = parentClass.Metadata.Name // Task 3.5.39

		// Copy parent fields (child inherits all parent fields)
		for fieldName, fieldType := range parentClass.Fields {
			classInfo.Fields[fieldName] = fieldType
		}
		// Copy parent field declarations (for initializers)
		for fieldName, fieldDecl := range parentClass.FieldDecls {
			classInfo.FieldDecls[fieldName] = fieldDecl
		}

		// Copy parent methods (child inherits all parent methods)
		// Keep Methods and ClassMethods for backward compatibility (direct lookups)
		for methodName, methodDecl := range parentClass.Methods {
			classInfo.Methods[methodName] = methodDecl
		}
		for methodName, methodDecl := range parentClass.ClassMethods {
			classInfo.ClassMethods[methodName] = methodDecl
		}

		// DON'T copy MethodOverloads/ClassMethodOverloads from parent
		// Each class should only store its OWN method overloads, not inherited ones.
		// getMethodOverloadsInHierarchy will walk the hierarchy to collect them at call time.
		// This prevents duplication when a child class overrides a parent method.

		// Copy constructors
		for name, constructor := range parentClass.Constructors {
			normalizedName := ident.Normalize(name)
			classInfo.Constructors[normalizedName] = constructor
		}
		for name, overloads := range parentClass.ConstructorOverloads {
			normalizedName := ident.Normalize(name)
			classInfo.ConstructorOverloads[normalizedName] = append([]*ast.FunctionDecl(nil), overloads...)
		}

		// Task 9.3: Inherit default constructor if parent has one
		if parentClass.DefaultConstructor != "" {
			classInfo.DefaultConstructor = parentClass.DefaultConstructor
		}

		// Task 3.5.39: Copy parent constructors to metadata
		for name, constructor := range parentClass.Metadata.Constructors {
			if classInfo.Metadata.Constructors == nil {
				classInfo.Metadata.Constructors = make(map[string]*runtime.MethodMetadata)
			}
			classInfo.Metadata.Constructors[name] = constructor
		}
		for name, overloads := range parentClass.Metadata.ConstructorOverloads {
			if classInfo.Metadata.ConstructorOverloads == nil {
				classInfo.Metadata.ConstructorOverloads = make(map[string][]*runtime.MethodMetadata)
			}
			classInfo.Metadata.ConstructorOverloads[name] = append([]*runtime.MethodMetadata(nil), overloads...)
		}

		// Task 3.5.39: Inherit default constructor name into metadata
		if parentClass.Metadata.DefaultConstructor != "" {
			classInfo.Metadata.DefaultConstructor = parentClass.Metadata.DefaultConstructor
		}

		// Copy operator overloads
		classInfo.Operators = parentClass.Operators.clone()
	}

	// Process implemented interfaces
	// Validate that each interface exists and store references
	for _, ifaceIdent := range cd.Interfaces {
		ifaceName := ifaceIdent.Value
		// Look up interface in registry (case-insensitive)
		// Task 3.5.184: Use TypeSystem lookup instead of i.interfaces map
		iface := i.lookupInterfaceInfo(ifaceName)
		if iface == nil {
			return i.newErrorWithLocation(cd, "interface '%s' not found", ifaceName)
		}

		// Add interface to class's interface list
		classInfo.Interfaces = append(classInfo.Interfaces, iface)

		// Task 3.5.39: Add interface name to metadata
		classInfo.Metadata.Interfaces = append(classInfo.Metadata.Interfaces, ifaceName)

		// Note: Method implementation validation is deferred until the class methods
		// are fully processed. For now, we just register that the class claims to
		// implement the interface. Full validation would happen in semantic analysis.
	}

	// Register class constants BEFORE processing fields
	// This allows class vars to reference constants in their initialization expressions
	// Evaluate constants eagerly in order so they can reference earlier constants
	for _, constDecl := range cd.Constants {
		if constDecl == nil {
			continue
		}
		// Store the constant declaration
		classInfo.Constants[constDecl.Name.Value] = constDecl

		// Evaluate the constant value immediately
		// Create temporary environment with previously evaluated constants
		savedEnv := i.env
		tempEnv := NewEnclosedEnvironment(i.env)
		for cName, cValue := range classInfo.ConstantValues {
			tempEnv.Define(cName, cValue)
		}
		i.env = tempEnv

		constValue := i.Eval(constDecl.Value)
		i.env = savedEnv

		if isError(constValue) {
			return constValue
		}

		// Cache the evaluated value
		classInfo.ConstantValues[constDecl.Name.Value] = constValue
	}

	// Copy parent constants (child inherits all parent constants)
	if classInfo.Parent != nil {
		for constName, constDecl := range classInfo.Parent.Constants {
			// Only copy if not already defined in child class
			if _, exists := classInfo.Constants[constName]; !exists {
				classInfo.Constants[constName] = constDecl
			}
		}
		// Also copy parent constant values
		for constName, constValue := range classInfo.Parent.ConstantValues {
			// Only copy if not already defined in child class
			if _, exists := classInfo.ConstantValues[constName]; !exists {
				classInfo.ConstantValues[constName] = constValue
			}
		}
	}

	// Task 9.6: Register class BEFORE processing fields
	// This allows field initializers to reference the class name (e.g., FField := TObj.Value)
	// Task 3.5.46: Register in legacy map early for field initializers that reference the class
	// Note: TypeSystem registration happens at end of function after VMT is built
	i.classes[ident.Normalize(classInfo.Name)] = classInfo

	// Evaluate nested type declarations before processing fields so nested classes
	// can be referenced by name inside the outer class.
	for _, nested := range cd.NestedTypes {
		switch n := nested.(type) {
		case *ast.ClassDecl:
			if n.EnclosingClass == nil {
				n.EnclosingClass = &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: cd.Name.Token,
						},
					},
					Value: classInfo.Name,
				}
			}
			if result := i.evalClassDeclaration(n); isError(result) {
				return result
			}
			if nestedInfo, ok := i.classes[ident.Normalize(i.fullClassNameFromDecl(n))]; ok {
				classInfo.NestedClasses[ident.Normalize(n.Name.Value)] = nestedInfo
			}
		default:
			if result := i.Eval(n); isError(result) {
				return result
			}
		}
	}

	// Add own fields to ClassInfo
	for _, field := range cd.Fields {
		var fieldType types.Type
		var cachedInitValue Value // Cache evaluated init value to avoid double evaluation

		// Get the field type from the type annotation or infer from initialization
		if field.Type != nil {
			// Explicit type annotation
			fieldType = i.resolveTypeFromExpression(field.Type)
			if fieldType == nil {
				return i.newErrorWithLocation(field, "unknown or invalid type for field '%s'", field.Name.Value)
			}
		} else if field.InitValue != nil {
			// Type inference from initialization value
			// Create temporary environment with class constants available
			savedEnv := i.env
			tempEnv := NewEnclosedEnvironment(i.env)
			for cName, cValue := range classInfo.ConstantValues {
				tempEnv.Define(cName, cValue)
			}
			i.env = tempEnv

			// Evaluate the init value to infer the type
			initVal := i.Eval(field.InitValue)
			i.env = savedEnv

			if isError(initVal) {
				return initVal
			}
			// Cache the evaluated value to reuse for class var initialization
			cachedInitValue = initVal

			// Infer type from the value
			fieldType = i.inferTypeFromValue(initVal)
			if fieldType == nil {
				return i.newErrorWithLocation(field, "cannot infer type for field '%s'", field.Name.Value)
			}
		} else {
			// No type and no initialization
			return i.newErrorWithLocation(field, "field '%s' has no type annotation", field.Name.Value)
		}

		// Check if this is a class variable (static field) or instance field
		if field.IsClassVar {
			var classVarValue Value

			// Check if there's an initialization expression
			if field.InitValue != nil {
				// Reuse cached value if available (from type inference)
				// This avoids double evaluation which would run side effects twice
				if cachedInitValue != nil {
					classVarValue = cachedInitValue
				} else {
					// Need to evaluate (explicit type annotation case)
					// Create temporary environment with class constants available
					savedEnv := i.env
					tempEnv := NewEnclosedEnvironment(i.env)
					for cName, cValue := range classInfo.ConstantValues {
						tempEnv.Define(cName, cValue)
					}
					i.env = tempEnv

					// Evaluate the initialization expression
					val := i.Eval(field.InitValue)
					i.env = savedEnv

					if isError(val) {
						return val
					}
					classVarValue = val
				}
			} else {
				// Initialize class variable with default value based on type
				switch fieldType {
				case types.INTEGER:
					classVarValue = &IntegerValue{Value: 0}
				case types.FLOAT:
					classVarValue = &FloatValue{Value: 0.0}
				case types.STRING:
					classVarValue = &StringValue{Value: ""}
				case types.BOOLEAN:
					classVarValue = &BooleanValue{Value: false}
				default:
					classVarValue = &NilValue{}
				}
			}
			classInfo.ClassVars[field.Name.Value] = classVarValue
		} else {
			// Store instance field type in ClassInfo
			classInfo.Fields[field.Name.Value] = fieldType
			// Store field declaration for initializer access
			classInfo.FieldDecls[field.Name.Value] = field

			// Task 3.5.39: Add to metadata
			fieldMeta := runtime.FieldMetadataFromAST(field)
			fieldMeta.Type = fieldType
			runtime.AddFieldToClass(classInfo.Metadata, fieldMeta)
		}
	}

	// Add own methods to ClassInfo (these override parent methods if same name)
	// Support method overloading by storing multiple methods per name
	for _, method := range cd.Methods {
		// Normalize method name for case-insensitive lookup
		// This matches the semantic analyzer behavior (types.go AddMethodOverload)
		normalizedMethodName := ident.Normalize(method.Name.Value)

		// Auto-detect constructors: methods named "Create" that return the class type
		// This handles inline constructor declarations like: function Create(...): TClass;
		// Matches semantic analyzer behavior (analyze_classes_decl.go:576-580)
		if !method.IsConstructor && ident.Equal(method.Name.Value, "Create") && method.ReturnType != nil {
			returnTypeName := method.ReturnType.String()
			if ident.Equal(returnTypeName, cd.Name.Value) {
				method.IsConstructor = true
			}
		}

		// Task 3.5.39: Create MethodMetadata once for this method
		methodMeta := runtime.MethodMetadataFromAST(method)
		i.methodRegistry.RegisterMethod(methodMeta)

		// Check if this is a class method (static method) or instance method
		if method.IsClassMethod {
			// Store in ClassMethods map
			classInfo.ClassMethods[normalizedMethodName] = method
			// Add to overload list
			classInfo.ClassMethodOverloads[normalizedMethodName] = append(classInfo.ClassMethodOverloads[normalizedMethodName], method)

			// Task 3.5.39: Add to metadata (unless it's a constructor/destructor - those go separately)
			if !method.IsConstructor && !method.IsDestructor {
				runtime.AddMethodToClass(classInfo.Metadata, methodMeta, true)
			}
		} else {
			// Store in instance Methods map
			classInfo.Methods[normalizedMethodName] = method
			// Add to overload list
			classInfo.MethodOverloads[normalizedMethodName] = append(classInfo.MethodOverloads[normalizedMethodName], method)

			// Task 3.5.39: Add to metadata (unless it's a constructor/destructor - those go separately)
			if !method.IsConstructor && !method.IsDestructor {
				runtime.AddMethodToClass(classInfo.Metadata, methodMeta, false)
			}
		}

		// Task 3.5.39: Handle destructor
		if method.IsDestructor {
			classInfo.Metadata.Destructor = methodMeta
		}

		if method.IsConstructor {
			normalizedName := ident.Normalize(method.Name.Value)
			classInfo.Constructors[normalizedName] = method

			// Task 3.5.39: Add constructor to metadata (reuse methodMeta)
			runtime.AddConstructorToClass(classInfo.Metadata, methodMeta)

			// Task 9.3: Capture default constructor
			if method.IsDefault {
				classInfo.DefaultConstructor = method.Name.Value
			}

			// In DWScript, a child constructor with the same name and signature HIDES the parent's,
			// regardless of whether it has the `override` keyword or not
			existingOverloads := classInfo.ConstructorOverloads[normalizedName]
			replaced := false
			for i, existingMethod := range existingOverloads {
				// Check if signatures match (same number and types of parameters)
				if parametersMatch(existingMethod.Parameters, method.Parameters) {
					// Replace the parent constructor with this child constructor (hiding)
					existingOverloads[i] = method
					replaced = true
					break
				}
			}
			if !replaced {
				// No matching parent constructor found (different signature), just append
				existingOverloads = append(existingOverloads, method)
			}
			// Write the modified slice back to the map
			classInfo.ConstructorOverloads[normalizedName] = existingOverloads
		}
	}

	// Identify constructor (method named "Create")
	if constructor, exists := classInfo.Methods["create"]; exists {
		classInfo.Constructor = constructor
	}
	if cd.Constructor != nil {
		normalizedName := ident.Normalize(cd.Constructor.Name.Value)
		classInfo.Constructors[normalizedName] = cd.Constructor

		// Task 3.5.39: Register constructor in metadata
		// Create and register method metadata for explicit constructor
		constructorMeta := runtime.MethodMetadataFromAST(cd.Constructor)
		i.methodRegistry.RegisterMethod(constructorMeta)
		runtime.AddConstructorToClass(classInfo.Metadata, constructorMeta)

		// In DWScript, a child constructor with the same name and signature HIDES the parent's,
		// regardless of whether it has the `override` keyword or not
		existingOverloads := classInfo.ConstructorOverloads[normalizedName]
		replaced := false
		for i, existingMethod := range existingOverloads {
			// Check if signatures match (same number and types of parameters)
			if parametersMatch(existingMethod.Parameters, cd.Constructor.Parameters) {
				// Replace the parent constructor with this child constructor (hiding)
				existingOverloads[i] = cd.Constructor
				replaced = true
				break
			}
		}
		if !replaced {
			// No matching parent constructor found (different signature), just append
			existingOverloads = append(existingOverloads, cd.Constructor)
		}
		// Write the modified slice back to the map
		classInfo.ConstructorOverloads[normalizedName] = existingOverloads
	}

	// Identify destructor (method named "Destroy")
	// Task 3.5.39: Destructor metadata is now set during the method loop above
	if destructor, exists := classInfo.Methods["destroy"]; exists {
		classInfo.Destructor = destructor
	}

	// Task 3.5.39: Inherit destructor from parent if no local destructor declared
	if classInfo.Metadata.Destructor == nil && classInfo.Parent != nil && classInfo.Parent.Metadata.Destructor != nil {
		classInfo.Metadata.Destructor = classInfo.Parent.Metadata.Destructor
	}

	// Synthesize implicit parameterless constructor if any constructor has 'overload'
	i.synthesizeImplicitParameterlessConstructor(classInfo)

	// Register properties
	// Properties are registered after fields and methods so they can reference them
	for _, propDecl := range cd.Properties {
		if propDecl == nil {
			continue
		}

		// Convert AST property to PropertyInfo
		propInfo := i.convertPropertyDecl(propDecl)
		if propInfo != nil {
			classInfo.Properties[propDecl.Name.Value] = propInfo
		}
	}

	// Copy parent properties (child inherits all parent properties)
	if classInfo.Parent != nil {
		for propName, propInfo := range classInfo.Parent.Properties {
			// Only copy if not already defined in child class
			if _, exists := classInfo.Properties[propName]; !exists {
				classInfo.Properties[propName] = propInfo
			}
		}
	}

	// Register class operators (Stage 8)
	for _, opDecl := range cd.Operators {
		if opDecl == nil {
			continue
		}
		if errVal := i.registerClassOperator(classInfo, opDecl); isError(errVal) {
			return errVal
		}
	}

	// Build virtual method table after all methods and fields are processed
	classInfo.buildVirtualMethodTable()

	// Register class in TypeSystem after VMT is built
	// Note: Legacy map registration already (done early) for field initializers
	parentName2 := ""
	if classInfo.Parent != nil {
		parentName2 = classInfo.Parent.Name
	}
	i.typeSystem.RegisterClassWithParent(classInfo.Name, classInfo, parentName2)

	return &NilValue{}
}

// convertPropertyDecl converts an AST property declaration to a PropertyInfo struct.
// This extracts the property metadata for runtime property access handling.
// Note: We mark all identifiers as field access for now and will check at runtime
// whether they're actually fields or methods.
func (i *Interpreter) convertPropertyDecl(propDecl *ast.PropertyDecl) *types.PropertyInfo {
	// Resolve property type
	var propType types.Type
	switch propDecl.Type.String() {
	case "Integer":
		propType = types.INTEGER
	case "Float":
		propType = types.FLOAT
	case "String":
		propType = types.STRING
	case "Boolean":
		propType = types.BOOLEAN
	default:
		// Try to resolve known class types; fall back to NIL if unknown
		if classInfo := i.resolveClassInfoByName(propDecl.Type.String()); classInfo != nil {
			propType = types.NewClassType(classInfo.Name, nil)
		} else {
			propType = types.NIL
		}
	}

	propInfo := &types.PropertyInfo{
		Name:            propDecl.Name.Value,
		Type:            propType,
		IsIndexed:       len(propDecl.IndexParams) > 0,
		IsDefault:       propDecl.IsDefault,
		IsClassProperty: propDecl.IsClassProperty,
	}

	if propDecl.IndexValue != nil {
		if val, ok := ast.ExtractIntegerLiteral(propDecl.IndexValue); ok {
			propInfo.HasIndexValue = true
			propInfo.IndexValue = val
			propInfo.IndexValueType = types.INTEGER
		}
	}

	// Determine read access kind and spec
	if propDecl.ReadSpec != nil {
		if ident, ok := propDecl.ReadSpec.(*ast.Identifier); ok {
			// It's an identifier - store the name, we'll check if it's a field or method at access time
			propInfo.ReadSpec = ident.Value
			// Mark as field for now - evalPropertyRead will check both fields and methods
			propInfo.ReadKind = types.PropAccessField
		} else {
			// It's an expression
			propInfo.ReadKind = types.PropAccessExpression
			propInfo.ReadSpec = propDecl.ReadSpec.String()
			propInfo.ReadExpr = propDecl.ReadSpec // Store AST node for evaluation
		}
	} else {
		propInfo.ReadKind = types.PropAccessNone
	}

	// Determine write access kind and spec
	if propDecl.WriteSpec != nil {
		if ident, ok := propDecl.WriteSpec.(*ast.Identifier); ok {
			// It's an identifier - store the name, we'll check if it's a field or method at access time
			propInfo.WriteSpec = ident.Value
			// Mark as field for now - evalPropertyWrite will check both fields and methods
			propInfo.WriteKind = types.PropAccessField
		} else {
			propInfo.WriteKind = types.PropAccessNone
		}
	} else {
		propInfo.WriteKind = types.PropAccessNone
	}

	return propInfo
}

// evalInterfaceDeclaration evaluates an interface declaration.
// It builds an InterfaceInfo from the AST and registers it in the interface registry.
// Handles inheritance by linking to parent interface and inheriting its methods.
func (i *Interpreter) evalInterfaceDeclaration(id *ast.InterfaceDecl) Value {
	// Create new InterfaceInfo
	interfaceInfo := NewInterfaceInfo(id.Name.Value)

	// Handle inheritance if parent interface is specified
	// Task 3.5.184: Use TypeSystem lookup instead of i.interfaces map
	if id.Parent != nil {
		parentName := id.Parent.Value
		parentInterface := i.lookupInterfaceInfo(parentName)
		if parentInterface == nil {
			return i.newErrorWithLocation(id, "parent interface '%s' not found", parentName)
		}

		// Set parent reference
		interfaceInfo.Parent = parentInterface

		// Note: We don't copy parent methods here because InterfaceInfo.GetMethod()
		// and AllMethods() already handle parent interface traversal
	}

	// Add methods to InterfaceInfo
	// Convert InterfaceMethodDecl nodes to FunctionDecl nodes for consistency
	for _, methodDecl := range id.Methods {
		// Create a FunctionDecl from the InterfaceMethodDecl
		// Interface methods are declarations only (no body)
		funcDecl := &ast.FunctionDecl{
			BaseNode: ast.BaseNode{
				Token: methodDecl.Token,
			},
			Name:       methodDecl.Name,
			Parameters: methodDecl.Parameters,
			ReturnType: methodDecl.ReturnType,
			Body:       nil, // Interface methods have no body
		}

		// Use normalized key for case-insensitive method lookups
		interfaceInfo.Methods[ident.Normalize(methodDecl.Name.Value)] = funcDecl
	}

	// Register properties declared on the interface
	for _, propDecl := range id.Properties {
		if propDecl == nil {
			continue
		}
		if propInfo := i.convertPropertyDecl(propDecl); propInfo != nil {
			interfaceInfo.Properties[ident.Normalize(propDecl.Name.Value)] = propInfo
		}
	}

	// Register interface in TypeSystem (Task 3.5.184c: only TypeSystem now, legacy map removed)
	i.typeSystem.RegisterInterface(interfaceInfo.Name, interfaceInfo)

	return &NilValue{}
}

// synthesizeImplicitParameterlessConstructor generates an implicit parameterless constructor
// when at least one constructor has the 'overload' directive.
//
// In DWScript, when a constructor is marked with 'overload', the runtime implicitly provides
// a parameterless constructor if one doesn't already exist. This allows code like:
//
//	type TObj = class
//	  constructor Create(x: Integer); overload;
//	end;
//	var o := TObj.Create;  // Calls implicit parameterless constructor
//	var p := TObj.Create(5);  // Calls explicit overload with parameter
func (i *Interpreter) synthesizeImplicitParameterlessConstructor(classInfo *ClassInfo) {
	// For each constructor name, check if it has the 'overload' directive
	// If so, ensure there's a parameterless overload
	for ctorName, overloads := range classInfo.ConstructorOverloads {
		hasOverloadDirective := false
		hasParameterlessOverload := false

		// Check if any overload has the 'overload' directive
		// and if a parameterless overload already exists
		for _, ctor := range overloads {
			if ctor.IsOverload {
				hasOverloadDirective = true
			}
			if len(ctor.Parameters) == 0 {
				hasParameterlessOverload = true
			}
		}

		// If this constructor set has 'overload' but no parameterless version, synthesize one
		if hasOverloadDirective && !hasParameterlessOverload {
			// Create a minimal constructor AST node (just for runtime - no actual body needed)
			// The interpreter will initialize fields with default values when no constructor body exists
			implicitConstructor := &ast.FunctionDecl{
				BaseNode:      ast.BaseNode{},
				Name:          &ast.Identifier{Value: ctorName},
				Parameters:    []*ast.Parameter{}, // No parameters
				ReturnType:    nil,                // Constructors don't have explicit return types
				Body:          nil,                // No body - just field initialization
				IsConstructor: true,
				IsOverload:    true,
			}

			// Add to class constructor maps
			normalizedName := ident.Normalize(ctorName)
			if _, exists := classInfo.Constructors[normalizedName]; !exists {
				classInfo.Constructors[normalizedName] = implicitConstructor
			}
			classInfo.ConstructorOverloads[normalizedName] = append(
				classInfo.ConstructorOverloads[normalizedName],
				implicitConstructor,
			)
		}
	}
}

func (i *Interpreter) evalOperatorDeclaration(decl *ast.OperatorDecl) Value {
	if decl.Kind == ast.OperatorKindClass {
		// Class operators are registered during class declaration evaluation
		return &NilValue{}
	}

	if decl.Binding == nil {
		return i.newErrorWithLocation(decl, "operator '%s' missing binding", decl.OperatorSymbol)
	}

	operandTypes := make([]string, len(decl.OperandTypes))
	for idx, operand := range decl.OperandTypes {
		opRand := operand.String()
		operandTypes[idx] = NormalizeTypeAnnotation(opRand)
	}

	if decl.Kind == ast.OperatorKindConversion {
		if len(operandTypes) != 1 {
			return i.newErrorWithLocation(decl, "conversion operator '%s' requires exactly one operand", decl.OperatorSymbol)
		}
		if decl.ReturnType == nil {
			return i.newErrorWithLocation(decl, "conversion operator '%s' requires a return type", decl.OperatorSymbol)
		}
		targetType := NormalizeTypeAnnotation(decl.ReturnType.String())
		entry := &interptypes.ConversionEntry{
			From: operandTypes[0],
			To:   targetType,
			// DWScript is case-insensitive, so normalize the binding name
			BindingName: ident.Normalize(decl.Binding.Value),
			Implicit:    ident.Equal(decl.OperatorSymbol, "implicit"),
		}
		if err := i.typeSystem.Conversions().Register(entry); err != nil {
			return i.newErrorWithLocation(decl, "conversion from %s to %s already defined", operandTypes[0], targetType)
		}
		return &NilValue{}
	}

	entry := &runtimeOperatorEntry{
		Operator:     decl.OperatorSymbol,
		OperandTypes: operandTypes,
		// DWScript is case-insensitive, so normalize the binding name
		BindingName:   ident.Normalize(decl.Binding.Value),
		Class:         nil,  // Global operator, not tied to a class
		SelfIndex:     -1,   // No self parameter for global operators
		IsClassMethod: false,
	}

	if err := i.globalOperators.register(entry); err != nil {
		return i.newErrorWithLocation(decl, "operator '%s' already defined for operand types (%s)", decl.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	return &NilValue{}
}

func (i *Interpreter) registerClassOperator(classInfo *ClassInfo, opDecl *ast.OperatorDecl) Value {
	if opDecl.Binding == nil {
		return i.newErrorWithLocation(opDecl, "class operator '%s' missing binding", opDecl.OperatorSymbol)
	}

	bindingName := opDecl.Binding.Value
	normalizedBindingName := ident.Normalize(bindingName)
	method, isClassMethod := classInfo.ClassMethods[normalizedBindingName]
	if !isClassMethod {
		var ok bool
		method, ok = classInfo.Methods[normalizedBindingName]
		if !ok {
			return i.newErrorWithLocation(opDecl, "binding '%s' for class operator '%s' not found in class '%s'", bindingName, opDecl.OperatorSymbol, classInfo.Name)
		}
	}

	classKey := NormalizeTypeAnnotation(classInfo.Name)
	operandTypes := make([]string, 0, len(opDecl.OperandTypes)+1)
	includesClass := false
	for _, operand := range opDecl.OperandTypes {
		// Resolve type aliases before normalizing
		// This ensures that "toa" (alias for "array of const") is resolved to "ARRAY OF CONST"
		typeName := operand.String()
		resolvedType, err := i.resolveType(typeName)
		var key string
		if err == nil {
			// Successfully resolved - use the resolved type's string representation
			key = NormalizeTypeAnnotation(resolvedType.String())
		} else {
			// Failed to resolve - use the raw type name (might be a forward reference)
			key = NormalizeTypeAnnotation(typeName)
		}
		if key == classKey {
			includesClass = true
		}
		operandTypes = append(operandTypes, key)
	}
	if !includesClass {
		if ident.Equal(opDecl.OperatorSymbol, "in") {
			operandTypes = append(operandTypes, classKey)
		} else {
			operandTypes = append([]string{classKey}, operandTypes...)
		}
	}

	selfIndex := -1
	if !isClassMethod {
		for idx, key := range operandTypes {
			if key == classKey {
				selfIndex = idx
				break
			}
		}
		if selfIndex == -1 {
			return i.newErrorWithLocation(opDecl, "unable to determine self operand for class operator '%s'", opDecl.OperatorSymbol)
		}
	}

	entry := &runtimeOperatorEntry{
		Operator:      opDecl.OperatorSymbol,
		OperandTypes:  operandTypes,
		BindingName:   normalizedBindingName,
		Class:         classInfo,
		IsClassMethod: isClassMethod,
		SelfIndex:     selfIndex,
	}

	if err := classInfo.Operators.register(entry); err != nil {
		return i.newErrorWithLocation(opDecl, "class operator '%s' already defined for operand types (%s)", opDecl.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	if method.IsConstructor {
		classInfo.Constructors[normalizedBindingName] = method
	}

	return &NilValue{}
}

// replaceMethodInOverloadList replaces a method declaration with its implementation in the overload list.
//
// This function finds a method with matching signature and replaces it, or appends if not found.
// parametersMatch checks if two parameter lists have matching signatures
// (same count and same parameter types)
func parametersMatch(params1, params2 []*ast.Parameter) bool {
	if len(params1) != len(params2) {
		return false
	}
	for i := range params1 {
		// Compare parameter types
		if params1[i].Type != nil && params2[i].Type != nil {
			if params1[i].Type.String() != params2[i].Type.String() {
				return false
			}
		} else if params1[i].Type != params2[i].Type {
			// One has type, other doesn't
			return false
		}
	}
	return true
}

func (i *Interpreter) replaceMethodInOverloadList(list []*ast.FunctionDecl, impl *ast.FunctionDecl) []*ast.FunctionDecl {
	// Check if we already have a declaration for this overload signature
	for idx, decl := range list {
		// Match by parameter count and types
		if parametersMatch(decl.Parameters, impl.Parameters) {
			// Task 9.14: Preserve virtual/override/reintroduce flags from declaration
			// The implementation doesn't have these keywords, but we need to preserve them
			impl.IsVirtual = decl.IsVirtual
			impl.IsOverride = decl.IsOverride
			impl.IsReintroduce = decl.IsReintroduce
			impl.IsAbstract = decl.IsAbstract

			// Replace the declaration with the implementation
			list[idx] = impl
			return list
		}
	}
	// No matching declaration found - append the implementation
	return append(list, impl)
}

// inferTypeFromValue infers the type from a runtime value.
// This is used for type inference when a variable or field is declared without an explicit type.
func (i *Interpreter) inferTypeFromValue(val Value) types.Type {
	switch val := val.(type) {
	case *IntegerValue:
		return types.INTEGER
	case *FloatValue:
		return types.FLOAT
	case *StringValue:
		return types.STRING
	case *BooleanValue:
		return types.BOOLEAN
	case *ArrayValue:
		// For arrays, we could try to infer the element type
		arrVal := val
		if len(arrVal.Elements) > 0 {
			elemType := i.inferTypeFromValue(arrVal.Elements[0])
			if elemType != nil {
				lowBound := 0
				highBound := len(arrVal.Elements) - 1
				return &types.ArrayType{
					ElementType: elemType,
					LowBound:    &lowBound,
					HighBound:   &highBound,
				}
			}
		}
		return nil
	case *ObjectInstance:
		// For object instances, type inference is complex
		// Return nil for now (type inference for objects may not be common for class vars)
		return nil
	case *NilValue:
		// Nil doesn't have a specific type
		return nil
	default:
		return nil
	}
}
