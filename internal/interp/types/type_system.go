// Package types provides the type system for the DWScript interpreter.
// This package contains type registries and type management utilities used
// during runtime execution.
package types

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	coretypes "github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// TypeSystem manages all type registries for the interpreter.
// It provides a centralized location for type information including
// classes, records, interfaces, functions, operators, and conversions.
//
// This design follows the Single Responsibility Principle by separating
// type management from execution concerns in the Interpreter.
type TypeSystem struct {
	nodeUnits            map[ast.Node]string
	operators            *OperatorRegistry                   // Operator overload registry
	conversions          *ConversionRegistry                 // Type conversion registry
	records              *ident.Map[RecordTypeValue]         // Record type registry
	interfaces           *ident.Map[InterfaceInfo]           // Interface registry
	helpers              *ident.Map[[]HelperInfo]            // Helper type registry
	arrayTypes           *ident.Map[*coretypes.ArrayType]    // Array type registry
	enumTypes            *ident.Map[EnumTypeValue]           // Enum type registry
	subrangeTypes        *ident.Map[*coretypes.SubrangeType] // Subrange type registry
	functionPointerTypes *ident.Map[coretypes.Type]          // Function/method pointer type registry
	functionRegistry     *FunctionRegistry                   // Function registry
	classTypeIDs         *ident.Map[int]                     // RTTI class type IDs
	classRegistry        *ClassRegistry                      // Class registry
	recordTypeIDs        *ident.Map[int]                     // RTTI record type IDs
	enumTypeIDs          *ident.Map[int]                     // RTTI enum type IDs
	nextRecordTypeID     int                                 // Next available record type ID
	nextEnumTypeID       int                                 // Next available enum type ID
	nextClassTypeID      int                                 // Next available class type ID
}

// NewTypeSystem creates a new TypeSystem with initialized registries.
func NewTypeSystem() *TypeSystem {
	return &TypeSystem{
		classRegistry:        NewClassRegistry(),
		functionRegistry:     NewFunctionRegistry(),
		records:              ident.NewMap[RecordTypeValue](),
		interfaces:           ident.NewMap[InterfaceInfo](),
		helpers:              ident.NewMap[[]HelperInfo](),
		arrayTypes:           ident.NewMap[*coretypes.ArrayType](),
		enumTypes:            ident.NewMap[EnumTypeValue](),
		subrangeTypes:        ident.NewMap[*coretypes.SubrangeType](),
		functionPointerTypes: ident.NewMap[coretypes.Type](),
		operators:            NewOperatorRegistry(),
		conversions:          NewConversionRegistry(),
		classTypeIDs:         ident.NewMap[int](),
		recordTypeIDs:        ident.NewMap[int](),
		enumTypeIDs:          ident.NewMap[int](),
		nextClassTypeID:      1000,   // Start class IDs at 1000
		nextRecordTypeID:     200000, // Start record IDs at 200000
		nextEnumTypeID:       300000, // Start enum IDs at 300000
	}
}

// ========== Class Registry ==========

// RegisterClass registers a new class in the type system.
// The name is stored case-insensitively (converted to lowercase).
func (ts *TypeSystem) RegisterClass(name string, class ClassInfo) {
	ts.classRegistry.Register(name, class)
}

// RegisterClassWithParent registers a class with an explicit parent name.
// This allows the ClassRegistry to track inheritance relationships.
func (ts *TypeSystem) RegisterClassWithParent(name string, class ClassInfo, parentName string) {
	ts.classRegistry.RegisterWithParent(name, class, parentName)
}

// UnregisterClass removes a class from the type system.
func (ts *TypeSystem) UnregisterClass(name string) {
	ts.classRegistry.Unregister(name)
}

// LookupClass returns the ClassInfo for the given name.
// The lookup is case-insensitive. Returns nil if not found.
func (ts *TypeSystem) LookupClass(name string) ClassInfo {
	info, ok := ts.classRegistry.Lookup(name)
	if !ok {
		return nil
	}
	return info
}

// CreateClassValue creates a metaclass reference for a registered runtime class.
func (ts *TypeSystem) CreateClassValue(className string) (runtime.Value, error) {
	classInfo := ts.LookupClass(className)
	if classInfo == nil {
		return nil, fmt.Errorf("class '%s' not found", className)
	}
	concrete, ok := classInfo.(*runtime.ClassInfo)
	if !ok {
		return nil, fmt.Errorf("class '%s' does not support metaclass references", className)
	}
	return &runtime.ClassValue{ClassInfo: concrete}, nil
}

// NewClassInfo allocates initialized runtime class metadata.
func (ts *TypeSystem) NewClassInfo(className string) (ClassInfo, error) {
	return runtime.NewClassInfo(className), nil
}

// HasClass checks if a class with the given name exists.
// The check is case-insensitive.
func (ts *TypeSystem) HasClass(name string) bool {
	return ts.classRegistry.Exists(name)
}

// AllClasses returns a map of all registered classes.
// Note: The returned map uses lowercase keys.
func (ts *TypeSystem) AllClasses() map[string]ClassInfo {
	return ts.classRegistry.GetAllClasses()
}

// LookupClassHierarchy returns all classes in the inheritance chain.
// The result is ordered from most specific (the class itself) to root.
// Returns nil if the class is not found.
func (ts *TypeSystem) LookupClassHierarchy(name string) []ClassInfo {
	return ts.classRegistry.LookupHierarchy(name)
}

// IsClassDescendantOf checks if descendantName inherits from ancestorName.
// Returns true if descendantName is derived from ancestorName (directly or indirectly).
// Also returns true if the names are equal (a class is its own descendant).
func (ts *TypeSystem) IsClassDescendantOf(descendantName, ancestorName string) bool {
	return ts.classRegistry.IsDescendantOf(descendantName, ancestorName)
}

// GetClassDepth returns the inheritance depth of a class.
// Depth 0 means no parent, depth 1 means one parent, etc.
// Returns -1 if the class is not found.
func (ts *TypeSystem) GetClassDepth(name string) int {
	return ts.classRegistry.GetDepth(name)
}

// Classes returns the ClassRegistry for direct access to advanced operations.
func (ts *TypeSystem) Classes() *ClassRegistry {
	return ts.classRegistry
}

// ========== Record Registry ==========

// RegisterRecord registers a new record type in the type system.
// The name is stored case-insensitively, with original casing preserved.
func (ts *TypeSystem) RegisterRecord(name string, record RecordTypeValue) {
	if record == nil {
		return
	}
	ts.records.Set(name, record)
}

// LookupRecord returns the RecordTypeValue for the given name.
// The lookup is case-insensitive. Returns nil if not found.
func (ts *TypeSystem) LookupRecord(name string) RecordTypeValue {
	record, _ := ts.records.Get(name)
	return record
}

// HasRecord checks if a record type with the given name exists.
// The check is case-insensitive.
func (ts *TypeSystem) HasRecord(name string) bool {
	return ts.records.Has(name)
}

// AllRecords returns a map of all registered record types.
// Note: The returned map uses normalized (lowercase) keys.
func (ts *TypeSystem) AllRecords() map[string]RecordTypeValue {
	result := make(map[string]RecordTypeValue, ts.records.Len())
	ts.records.Range(func(key string, value RecordTypeValue) bool {
		result[ident.Normalize(key)] = value
		return true
	})
	return result
}

// LookupRecordMetadata returns the runtime RecordMetadata for the given record name.
// Returns nil if the record doesn't exist or has no metadata.
func (ts *TypeSystem) LookupRecordMetadata(name string) *runtime.RecordMetadata {
	recordTypeValue := ts.LookupRecord(name)
	if recordTypeValue == nil {
		return nil
	}

	return recordTypeValue.Metadata
}

// ========== Interface Registry ==========

// RegisterInterface registers a new interface in the type system.
// The name is stored case-insensitively, with original casing preserved.
func (ts *TypeSystem) RegisterInterface(name string, iface InterfaceInfo) {
	if iface == nil {
		return
	}
	ts.interfaces.Set(name, iface)
}

// LookupInterface returns the InterfaceInfo for the given name.
// The lookup is case-insensitive. Returns nil if not found.
func (ts *TypeSystem) LookupInterface(name string) InterfaceInfo {
	iface, _ := ts.interfaces.Get(name)
	return iface
}

// HasInterface checks if an interface with the given name exists.
// The check is case-insensitive.
func (ts *TypeSystem) HasInterface(name string) bool {
	return ts.interfaces.Has(name)
}

// AllInterfaces returns a map of all registered interfaces.
// Note: The returned map uses normalized (lowercase) keys.
func (ts *TypeSystem) AllInterfaces() map[string]InterfaceInfo {
	result := make(map[string]InterfaceInfo, ts.interfaces.Len())
	ts.interfaces.Range(func(key string, value InterfaceInfo) bool {
		result[ident.Normalize(key)] = value
		return true
	})
	return result
}

// ========== Array Type Registry ==========

// RegisterArrayType registers an array type in the type system.
// The name is stored case-insensitively, with original casing preserved.
func (ts *TypeSystem) RegisterArrayType(name string, arrayType *coretypes.ArrayType) {
	if arrayType == nil {
		return
	}
	ts.arrayTypes.Set(name, arrayType)
}

// LookupArrayType returns the ArrayType for the given name.
// The lookup is case-insensitive. Returns nil if not found.
func (ts *TypeSystem) LookupArrayType(name string) *coretypes.ArrayType {
	arrayType, _ := ts.arrayTypes.Get(name)
	return arrayType
}

// HasArrayType checks if an array type with the given name exists.
// The check is case-insensitive.
func (ts *TypeSystem) HasArrayType(name string) bool {
	return ts.arrayTypes.Has(name)
}

// AllArrayTypes returns a map of all registered array types.
// Note: The returned map uses normalized (lowercase) keys.
func (ts *TypeSystem) AllArrayTypes() map[string]*coretypes.ArrayType {
	result := make(map[string]*coretypes.ArrayType, ts.arrayTypes.Len())
	ts.arrayTypes.Range(func(key string, value *coretypes.ArrayType) bool {
		result[ident.Normalize(key)] = value
		return true
	})
	return result
}

// ========== Enum Type Registry ==========

// RegisterEnumType registers an enum type in the type system.
// The name is stored case-insensitively, with original casing preserved.
func (ts *TypeSystem) RegisterEnumType(name string, enumType EnumTypeValue) {
	if enumType == nil {
		return
	}
	ts.enumTypes.Set(name, enumType)
}

// LookupEnumType returns the EnumTypeValue for the given name.
// The lookup is case-insensitive. Returns nil if not found.
func (ts *TypeSystem) LookupEnumType(name string) EnumTypeValue {
	enumType, _ := ts.enumTypes.Get(name)
	return enumType
}

// HasEnumType checks if an enum type with the given name exists.
// The check is case-insensitive.
func (ts *TypeSystem) HasEnumType(name string) bool {
	return ts.enumTypes.Has(name)
}

// AllEnumTypes returns a map of all registered enum types.
// Note: The returned map uses normalized (lowercase) keys.
func (ts *TypeSystem) AllEnumTypes() map[string]EnumTypeValue {
	result := make(map[string]EnumTypeValue, ts.enumTypes.Len())
	ts.enumTypes.Range(func(key string, value EnumTypeValue) bool {
		result[ident.Normalize(key)] = value
		return true
	})
	return result
}

// ========== Function Pointer Registry ==========

// RegisterFunctionPointerType registers a function or method pointer type.
// The name is stored case-insensitively.
func (ts *TypeSystem) RegisterFunctionPointerType(name string, funcType coretypes.Type) {
	if funcType == nil {
		return
	}
	ts.functionPointerTypes.Set(name, funcType)
}

// LookupFunctionPointerType returns the registered function or method pointer type by name.
func (ts *TypeSystem) LookupFunctionPointerType(name string) coretypes.Type {
	fp, _ := ts.functionPointerTypes.Get(name)
	return fp
}

// HasFunctionPointerType checks if a function or method pointer type exists.
func (ts *TypeSystem) HasFunctionPointerType(name string) bool {
	return ts.functionPointerTypes.Has(name)
}

// LookupEnumMetadata returns the EnumTypeValue wrapper for the given enum name.
// Returns nil if the enum doesn't exist.
func (ts *TypeSystem) LookupEnumMetadata(name string) EnumTypeValue {
	return ts.LookupEnumType(name)
}

// ========== Subrange Type Registry ==========

// RegisterSubrangeType registers a subrange type in the type system.
// The name is stored case-insensitively, with original casing preserved.
func (ts *TypeSystem) RegisterSubrangeType(name string, subrangeType *coretypes.SubrangeType) {
	if subrangeType == nil {
		return
	}
	ts.subrangeTypes.Set(name, subrangeType)
}

// LookupSubrangeType returns the SubrangeType for the given name.
// The lookup is case-insensitive. Returns nil if not found.
func (ts *TypeSystem) LookupSubrangeType(name string) *coretypes.SubrangeType {
	subrangeType, _ := ts.subrangeTypes.Get(name)
	return subrangeType
}

// HasSubrangeType checks if a subrange type with the given name exists.
// The check is case-insensitive.
func (ts *TypeSystem) HasSubrangeType(name string) bool {
	return ts.subrangeTypes.Has(name)
}

// AllSubrangeTypes returns a map of all registered subrange types.
// Note: The returned map uses normalized (lowercase) keys.
func (ts *TypeSystem) AllSubrangeTypes() map[string]*coretypes.SubrangeType {
	result := make(map[string]*coretypes.SubrangeType, ts.subrangeTypes.Len())
	ts.subrangeTypes.Range(func(key string, value *coretypes.SubrangeType) bool {
		result[ident.Normalize(key)] = value
		return true
	})
	return result
}

// ========== Function Registry ==========

// RegisterFunction registers a function overload in the type system.
// Multiple functions with the same name can be registered (overloading).
func (ts *TypeSystem) RegisterFunction(name string, fn *ast.FunctionDecl) {
	ts.functionRegistry.Register(name, fn)
}

// RegisterFunctionOrReplace registers a function, replacing declaration-only version
// if this is an implementation (has a body).
func (ts *TypeSystem) RegisterFunctionOrReplace(name string, fn *ast.FunctionDecl) {
	ts.functionRegistry.RegisterOrReplace(name, fn)
}

// RegisterFunctionWithUnit registers a function with an associated unit name.
// This allows for qualified lookups (UnitName.FunctionName).
func (ts *TypeSystem) RegisterFunctionWithUnit(unitName, functionName string, fn *ast.FunctionDecl) {
	if fn != nil {
		ts.RegisterNodeUnit(fn, unitName)
	}
	ts.functionRegistry.RegisterWithUnit(unitName, functionName, fn)
}

// RegisterFunctionWithUnitOrReplace registers a unit-owned function, replacing
// a matching declaration-only version when an implementation is imported.
func (ts *TypeSystem) RegisterFunctionWithUnitOrReplace(unitName, functionName string, fn *ast.FunctionDecl) {
	if fn != nil {
		ts.RegisterNodeUnit(fn, unitName)
	}
	ts.functionRegistry.RegisterOrReplaceWithUnit(unitName, functionName, fn)
}

// LookupFunctions returns all overloads for the given function name.
// Returns nil if no function with that name exists.
func (ts *TypeSystem) LookupFunctions(name string) []*ast.FunctionDecl {
	return ts.functionRegistry.Lookup(name)
}

// LookupQualifiedFunction returns all overloads for a qualified function name (Unit.Function).
// The lookup is case-insensitive. Returns nil if no functions found.
func (ts *TypeSystem) LookupQualifiedFunction(unitName, functionName string) []*ast.FunctionDecl {
	return ts.functionRegistry.LookupQualified(unitName, functionName)
}

// HasFunction checks if any function with the given name exists.
func (ts *TypeSystem) HasFunction(name string) bool {
	return ts.functionRegistry.Exists(name)
}

// HasQualifiedFunction checks if a qualified function exists (Unit.Function).
func (ts *TypeSystem) HasQualifiedFunction(unitName, functionName string) bool {
	return ts.functionRegistry.ExistsQualified(unitName, functionName)
}

// AllFunctions returns the entire function registry.
// The returned map should not be modified directly.
func (ts *TypeSystem) AllFunctions() map[string][]*ast.FunctionDecl {
	return ts.functionRegistry.GetAllFunctions()
}

// GetFunctionOverloadCount returns the number of overloads for a function.
func (ts *TypeSystem) GetFunctionOverloadCount(name string) int {
	return ts.functionRegistry.GetOverloadCount(name)
}

// Functions returns the FunctionRegistry for direct access to advanced operations.
func (ts *TypeSystem) Functions() *FunctionRegistry {
	return ts.functionRegistry
}

// ========== Helper Registry ==========

// RegisterHelper registers a helper method for a type.
// The type name is stored case-insensitively, with original casing preserved.
func (ts *TypeSystem) RegisterHelper(typeName string, helper HelperInfo) {
	if helper == nil {
		return
	}
	existing, _ := ts.helpers.Get(typeName)
	ts.helpers.Set(typeName, append(existing, helper))
}

// LookupHelpers returns all helper methods for the given type name.
// Returns nil if no helpers exist for the type.
func (ts *TypeSystem) LookupHelpers(typeName string) []HelperInfo {
	helpers, _ := ts.helpers.Get(typeName)
	return helpers
}

// HasHelpers checks if any helper methods exist for the given type.
func (ts *TypeSystem) HasHelpers(typeName string) bool {
	helpers, exists := ts.helpers.Get(typeName)
	return exists && len(helpers) > 0
}

// AllHelpers returns the entire helper registry.
// Note: The returned map uses normalized (lowercase) keys.
func (ts *TypeSystem) AllHelpers() map[string][]HelperInfo {
	result := make(map[string][]HelperInfo, ts.helpers.Len())
	ts.helpers.Range(func(key string, value []HelperInfo) bool {
		result[ident.Normalize(key)] = value
		return true
	})
	return result
}

// ========== Operator Registry ==========

// Operators returns the operator registry for managing operator overloads.
func (ts *TypeSystem) Operators() *OperatorRegistry {
	return ts.operators
}

// ========== Conversion Registry ==========

// Conversions returns the conversion registry for managing type conversions.
func (ts *TypeSystem) Conversions() *ConversionRegistry {
	return ts.conversions
}

// ========== Type ID Registry (RTTI) ==========

// GetOrAllocateClassTypeID returns the RTTI type ID for a class.
// If the class doesn't have an ID yet, a new one is allocated.
func (ts *TypeSystem) GetOrAllocateClassTypeID(className string) int {
	if id, exists := ts.classTypeIDs.Get(className); exists {
		return id
	}
	id := ts.nextClassTypeID
	ts.classTypeIDs.Set(className, id)
	ts.nextClassTypeID++
	return id
}

// GetClassTypeID returns the RTTI type ID for a class if it exists.
// Returns 0 if the class doesn't have an allocated type ID.
func (ts *TypeSystem) GetClassTypeID(className string) int {
	id, _ := ts.classTypeIDs.Get(className)
	return id
}

// GetOrAllocateRecordTypeID returns the RTTI type ID for a record.
// If the record doesn't have an ID yet, a new one is allocated.
func (ts *TypeSystem) GetOrAllocateRecordTypeID(recordName string) int {
	if id, exists := ts.recordTypeIDs.Get(recordName); exists {
		return id
	}
	id := ts.nextRecordTypeID
	ts.recordTypeIDs.Set(recordName, id)
	ts.nextRecordTypeID++
	return id
}

// GetRecordTypeID returns the RTTI type ID for a record if it exists.
// Returns 0 if the record doesn't have an allocated type ID.
func (ts *TypeSystem) GetRecordTypeID(recordName string) int {
	id, _ := ts.recordTypeIDs.Get(recordName)
	return id
}

// GetOrAllocateEnumTypeID returns the RTTI type ID for an enum.
// If the enum doesn't have an ID yet, a new one is allocated.
func (ts *TypeSystem) GetOrAllocateEnumTypeID(enumName string) int {
	if id, exists := ts.enumTypeIDs.Get(enumName); exists {
		return id
	}
	id := ts.nextEnumTypeID
	ts.enumTypeIDs.Set(enumName, id)
	ts.nextEnumTypeID++
	return id
}

// GetEnumTypeID returns the RTTI type ID for an enum if it exists.
// Returns 0 if the enum doesn't have an allocated type ID.
func (ts *TypeSystem) GetEnumTypeID(enumName string) int {
	id, _ := ts.enumTypeIDs.Get(enumName)
	return id
}

// ========== Type Information ==========

// ClassInfo provides runtime class metadata to the type registry.
type ClassInfo = runtime.IClassInfo

// RecordTypeValue provides runtime record metadata to the type registry.
type RecordTypeValue = *runtime.RecordTypeValue

// InterfaceInfo provides runtime interface metadata to the type registry.
type InterfaceInfo = *runtime.MutableInterfaceInfo

// HelperInfo provides runtime helper metadata to the type registry.
type HelperInfo = *runtime.MutableHelperInfo

// EnumTypeValue provides runtime enum metadata to the type registry.
type EnumTypeValue = *runtime.EnumTypeValue

// ========== Operator Registry ==========

// OperatorRegistry manages operator overloads.
// Uses ident.Map for case-insensitive operator lookup (e.g., "and" vs "AND").
type OperatorRegistry struct {
	entries *ident.Map[[]*OperatorEntry]
}

// OperatorEntry represents a registered operator overload.
type OperatorEntry struct {
	Class         runtime.IClassInfo
	Operator      string
	BindingName   string
	OperandTypes  []coretypes.Type
	SelfIndex     int
	IsClassMethod bool
}

// NewOperatorRegistry creates a new operator registry.
func NewOperatorRegistry() *OperatorRegistry {
	return &OperatorRegistry{
		entries: ident.NewMap[[]*OperatorEntry](),
	}
}

// Register registers a new operator overload.
// Returns an error if an operator with the same signature is already registered.
func (r *OperatorRegistry) Register(entry *OperatorEntry) error {
	if entry == nil {
		return fmt.Errorf("operator entry cannot be nil")
	}

	// Check for duplicate signatures
	existing, _ := r.entries.Get(entry.Operator)
	for _, e := range existing {
		if coretypes.OperatorOperandsEqual(e.OperandTypes, entry.OperandTypes) {
			return fmt.Errorf("operator already registered")
		}
	}

	r.entries.Set(entry.Operator, append(existing, entry))
	return nil
}

// Lookup finds an operator overload matching the given operator and operand types.
// Returns the entry and true if found, nil and false otherwise.
func (r *OperatorRegistry) Lookup(operator string, operandTypes []coretypes.Type) (*OperatorEntry, bool) {
	if r == nil {
		return nil, false
	}

	entries, ok := r.entries.Get(operator)
	if !ok {
		return nil, false
	}

	return coretypes.SelectOperatorOverload(operandTypes, entries, func(entry *OperatorEntry) []coretypes.Type { return entry.OperandTypes })
}

// Clone creates a deep copy of the operator registry.
func (r *OperatorRegistry) Clone() *OperatorRegistry {
	if r == nil {
		return NewOperatorRegistry()
	}
	clone := NewOperatorRegistry()
	r.entries.Range(func(op string, list []*OperatorEntry) bool {
		copied := make([]*OperatorEntry, len(list))
		copy(copied, list)
		clone.entries.Set(op, copied)
		return true
	})
	return clone
}

// ConversionRegistry stores conversions in declaration order for deterministic chaining.
type ConversionRegistry struct{ implicit, explicit []*ConversionEntry }

// ConversionEntry binds a conversion between resolved language types.
type ConversionEntry struct {
	From        coretypes.Type
	To          coretypes.Type
	BindingName string
	Implicit    bool
}

// NewConversionRegistry creates an empty conversion registry.
func NewConversionRegistry() *ConversionRegistry { return &ConversionRegistry{} }

// Register adds a conversion, rejecting duplicate typed signatures.
func (r *ConversionRegistry) Register(entry *ConversionEntry) error {
	if entry == nil || entry.From == nil || entry.To == nil {
		return fmt.Errorf("conversion requires resolved source and target types")
	}
	entries := &r.explicit
	if entry.Implicit {
		entries = &r.implicit
	}
	for _, old := range *entries {
		if coretypes.OperatorTypesEqual(old.From, entry.From) && coretypes.OperatorTypesEqual(old.To, entry.To) {
			return fmt.Errorf("conversion already registered")
		}
	}
	*entries = append(*entries, entry)
	return nil
}

// FindImplicit returns the conversion with the requested resolved signature.
func (r *ConversionRegistry) FindImplicit(from, to coretypes.Type) (*ConversionEntry, bool) {
	if r == nil {
		return nil, false
	}
	for _, entry := range r.implicit {
		if coretypes.OperatorTypesEqual(entry.From, from) && coretypes.OperatorTypesEqual(entry.To, to) {
			return entry, true
		}
	}
	return nil, false
}

// FindConversionPath finds the shortest implicit conversion chain, up to maxDepth steps.
func (r *ConversionRegistry) FindConversionPath(from, to coretypes.Type, maxDepth int) []coretypes.Type {
	if r == nil || from == nil || to == nil || maxDepth <= 0 {
		return nil
	}
	queue := [][]coretypes.Type{{from}}
	visited := []coretypes.Type{from}
	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		if len(path) > maxDepth {
			continue
		}
		current := path[len(path)-1]
		for _, entry := range r.implicit {
			if !coretypes.OperatorTypesEqual(entry.From, current) {
				continue
			}
			next := append(append([]coretypes.Type(nil), path...), entry.To)
			if coretypes.OperatorTypesEqual(entry.To, to) {
				return next
			}
			seen := false
			for _, old := range visited {
				if coretypes.OperatorTypesEqual(old, entry.To) {
					seen = true
					break
				}
			}
			if !seen {
				visited = append(visited, entry.To)
				queue = append(queue, next)
			}
		}
	}
	return nil
}
