// Package types provides the type system for the DWScript interpreter.
// This package contains type registries and type management utilities used
// during runtime execution.
//
// Task 3.4.1: Extract type registries from Interpreter into TypeSystem
package types

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
)

// TypeSystem manages all type registries for the interpreter.
// It provides a centralized location for type information including
// classes, records, interfaces, functions, operators, and conversions.
//
// This design follows the Single Responsibility Principle by separating
// type management from execution concerns in the Interpreter.
type TypeSystem struct {
	// Class registry: Task 3.4.2 - using ClassRegistry abstraction
	classRegistry *ClassRegistry

	// Record registry: maps lowercase record names to RecordTypeValue
	records map[string]*RecordTypeValue

	// Interface registry: maps lowercase interface names to InterfaceInfo
	interfaces map[string]*InterfaceInfo

	// Function registry: maps function names to overload lists
	// Key is function name (case-insensitive), value is list of overloads
	functions map[string][]*ast.FunctionDecl

	// Helper registry: maps type names to helper method lists
	// Key is type name (uppercase), value is list of helper methods
	helpers map[string][]*HelperInfo

	// Operator registry: manages operator overloads
	operators *OperatorRegistry

	// Conversion registry: manages type conversions
	conversions *ConversionRegistry

	// Type ID registries for RTTI (Task 9.25)
	classTypeIDs  map[string]int
	recordTypeIDs map[string]int
	enumTypeIDs   map[string]int

	// Next available type IDs
	nextClassTypeID  int
	nextRecordTypeID int
	nextEnumTypeID   int
}

// NewTypeSystem creates a new TypeSystem with initialized registries.
func NewTypeSystem() *TypeSystem {
	return &TypeSystem{
		classRegistry:    NewClassRegistry(), // Task 3.4.2
		records:          make(map[string]*RecordTypeValue),
		interfaces:       make(map[string]*InterfaceInfo),
		functions:        make(map[string][]*ast.FunctionDecl),
		helpers:          make(map[string][]*HelperInfo),
		operators:        NewOperatorRegistry(),
		conversions:      NewConversionRegistry(),
		classTypeIDs:     make(map[string]int),
		recordTypeIDs:    make(map[string]int),
		enumTypeIDs:      make(map[string]int),
		nextClassTypeID:  1000,   // Start class IDs at 1000
		nextRecordTypeID: 200000, // Start record IDs at 200000
		nextEnumTypeID:   300000, // Start enum IDs at 300000
	}
}

// ========== Class Registry ==========
// Task 3.4.2: Class methods now delegate to ClassRegistry

// RegisterClass registers a new class in the type system.
// The name is stored case-insensitively (converted to lowercase).
func (ts *TypeSystem) RegisterClass(name string, class *ClassInfo) {
	ts.classRegistry.Register(name, class)
}

// RegisterClassWithParent registers a class with an explicit parent name.
// This allows the ClassRegistry to track inheritance relationships.
func (ts *TypeSystem) RegisterClassWithParent(name string, class *ClassInfo, parentName string) {
	ts.classRegistry.RegisterWithParent(name, class, parentName)
}

// LookupClass returns the ClassInfo for the given name.
// The lookup is case-insensitive. Returns nil if not found.
func (ts *TypeSystem) LookupClass(name string) *ClassInfo {
	info, ok := ts.classRegistry.Lookup(name)
	if !ok {
		return nil
	}
	// Type assertion from any to *ClassInfo
	if classInfo, ok := info.(*ClassInfo); ok {
		return classInfo
	}
	return nil
}

// HasClass checks if a class with the given name exists.
// The check is case-insensitive.
func (ts *TypeSystem) HasClass(name string) bool {
	return ts.classRegistry.Exists(name)
}

// AllClasses returns a map of all registered classes.
// Note: The returned map uses lowercase keys.
func (ts *TypeSystem) AllClasses() map[string]*ClassInfo {
	allClasses := ts.classRegistry.GetAllClasses()
	result := make(map[string]*ClassInfo, len(allClasses))
	for key, info := range allClasses {
		if classInfo, ok := info.(*ClassInfo); ok {
			result[key] = classInfo
		}
	}
	return result
}

// LookupClassHierarchy returns all classes in the inheritance chain.
// The result is ordered from most specific (the class itself) to root.
// Returns nil if the class is not found.
func (ts *TypeSystem) LookupClassHierarchy(name string) []*ClassInfo {
	hierarchy := ts.classRegistry.LookupHierarchy(name)
	if hierarchy == nil {
		return nil
	}

	result := make([]*ClassInfo, 0, len(hierarchy))
	for _, info := range hierarchy {
		if classInfo, ok := info.(*ClassInfo); ok {
			result = append(result, classInfo)
		}
	}
	return result
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
// The name is stored case-insensitively (converted to lowercase).
func (ts *TypeSystem) RegisterRecord(name string, record *RecordTypeValue) {
	if record == nil {
		return
	}
	ts.records[strings.ToLower(name)] = record
}

// LookupRecord returns the RecordTypeValue for the given name.
// The lookup is case-insensitive. Returns nil if not found.
func (ts *TypeSystem) LookupRecord(name string) *RecordTypeValue {
	return ts.records[strings.ToLower(name)]
}

// HasRecord checks if a record type with the given name exists.
// The check is case-insensitive.
func (ts *TypeSystem) HasRecord(name string) bool {
	_, exists := ts.records[strings.ToLower(name)]
	return exists
}

// AllRecords returns a map of all registered record types.
// Note: The returned map uses lowercase keys.
func (ts *TypeSystem) AllRecords() map[string]*RecordTypeValue {
	return ts.records
}

// ========== Interface Registry ==========

// RegisterInterface registers a new interface in the type system.
// The name is stored case-insensitively (converted to lowercase).
func (ts *TypeSystem) RegisterInterface(name string, iface *InterfaceInfo) {
	if iface == nil {
		return
	}
	ts.interfaces[strings.ToLower(name)] = iface
}

// LookupInterface returns the InterfaceInfo for the given name.
// The lookup is case-insensitive. Returns nil if not found.
func (ts *TypeSystem) LookupInterface(name string) *InterfaceInfo {
	return ts.interfaces[strings.ToLower(name)]
}

// HasInterface checks if an interface with the given name exists.
// The check is case-insensitive.
func (ts *TypeSystem) HasInterface(name string) bool {
	_, exists := ts.interfaces[strings.ToLower(name)]
	return exists
}

// AllInterfaces returns a map of all registered interfaces.
// Note: The returned map uses lowercase keys.
func (ts *TypeSystem) AllInterfaces() map[string]*InterfaceInfo {
	return ts.interfaces
}

// ========== Function Registry ==========

// RegisterFunction registers a function overload in the type system.
// Multiple functions with the same name can be registered (overloading).
func (ts *TypeSystem) RegisterFunction(name string, fn *ast.FunctionDecl) {
	if fn == nil {
		return
	}
	ts.functions[name] = append(ts.functions[name], fn)
}

// LookupFunctions returns all overloads for the given function name.
// Returns nil if no function with that name exists.
func (ts *TypeSystem) LookupFunctions(name string) []*ast.FunctionDecl {
	return ts.functions[name]
}

// HasFunction checks if any function with the given name exists.
func (ts *TypeSystem) HasFunction(name string) bool {
	overloads, exists := ts.functions[name]
	return exists && len(overloads) > 0
}

// AllFunctions returns the entire function registry.
// The returned map should not be modified directly.
func (ts *TypeSystem) AllFunctions() map[string][]*ast.FunctionDecl {
	return ts.functions
}

// ========== Helper Registry ==========

// RegisterHelper registers a helper method for a type.
// The type name is stored in uppercase for consistency.
func (ts *TypeSystem) RegisterHelper(typeName string, helper *HelperInfo) {
	if helper == nil {
		return
	}
	key := strings.ToUpper(typeName)
	ts.helpers[key] = append(ts.helpers[key], helper)
}

// LookupHelpers returns all helper methods for the given type name.
// Returns nil if no helpers exist for the type.
func (ts *TypeSystem) LookupHelpers(typeName string) []*HelperInfo {
	return ts.helpers[strings.ToUpper(typeName)]
}

// HasHelpers checks if any helper methods exist for the given type.
func (ts *TypeSystem) HasHelpers(typeName string) bool {
	helpers, exists := ts.helpers[strings.ToUpper(typeName)]
	return exists && len(helpers) > 0
}

// AllHelpers returns the entire helper registry.
// The returned map should not be modified directly.
func (ts *TypeSystem) AllHelpers() map[string][]*HelperInfo {
	return ts.helpers
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
	normalized := strings.ToLower(className)
	if id, exists := ts.classTypeIDs[normalized]; exists {
		return id
	}
	id := ts.nextClassTypeID
	ts.classTypeIDs[normalized] = id
	ts.nextClassTypeID++
	return id
}

// GetClassTypeID returns the RTTI type ID for a class if it exists.
// Returns 0 if the class doesn't have an allocated type ID.
func (ts *TypeSystem) GetClassTypeID(className string) int {
	return ts.classTypeIDs[strings.ToLower(className)]
}

// GetOrAllocateRecordTypeID returns the RTTI type ID for a record.
// If the record doesn't have an ID yet, a new one is allocated.
func (ts *TypeSystem) GetOrAllocateRecordTypeID(recordName string) int {
	normalized := strings.ToLower(recordName)
	if id, exists := ts.recordTypeIDs[normalized]; exists {
		return id
	}
	id := ts.nextRecordTypeID
	ts.recordTypeIDs[normalized] = id
	ts.nextRecordTypeID++
	return id
}

// GetRecordTypeID returns the RTTI type ID for a record if it exists.
// Returns 0 if the record doesn't have an allocated type ID.
func (ts *TypeSystem) GetRecordTypeID(recordName string) int {
	return ts.recordTypeIDs[strings.ToLower(recordName)]
}

// GetOrAllocateEnumTypeID returns the RTTI type ID for an enum.
// If the enum doesn't have an ID yet, a new one is allocated.
func (ts *TypeSystem) GetOrAllocateEnumTypeID(enumName string) int {
	normalized := strings.ToLower(enumName)
	if id, exists := ts.enumTypeIDs[normalized]; exists {
		return id
	}
	id := ts.nextEnumTypeID
	ts.enumTypeIDs[normalized] = id
	ts.nextEnumTypeID++
	return id
}

// GetEnumTypeID returns the RTTI type ID for an enum if it exists.
// Returns 0 if the enum doesn't have an allocated type ID.
func (ts *TypeSystem) GetEnumTypeID(enumName string) int {
	return ts.enumTypeIDs[strings.ToLower(enumName)]
}

// ========== Type Information ==========
// The TypeSystem stores references to types defined in the interp package.
// We use 'any' (interface{}) to avoid circular dependencies between packages.
// The interp package will cast these to the appropriate concrete types.

type ClassInfo = any
type RecordTypeValue = any
type InterfaceInfo = any
type HelperInfo = any

// ========== Operator Registry ==========

// OperatorRegistry manages operator overloads.
type OperatorRegistry struct {
	entries map[string][]*OperatorEntry
}

// OperatorEntry represents a registered operator overload.
type OperatorEntry struct {
	Class         interface{} // *ClassInfo (avoiding import cycle)
	Operator      string
	BindingName   string
	OperandTypes  []string
	SelfIndex     int
	IsClassMethod bool
}

// NewOperatorRegistry creates a new operator registry.
func NewOperatorRegistry() *OperatorRegistry {
	return &OperatorRegistry{
		entries: make(map[string][]*OperatorEntry),
	}
}

// Register registers a new operator overload.
// Returns an error if an operator with the same signature is already registered.
func (r *OperatorRegistry) Register(entry *OperatorEntry) error {
	if entry == nil {
		return fmt.Errorf("operator entry cannot be nil")
	}
	key := strings.ToLower(entry.Operator)

	// Check for duplicate signatures
	for _, existing := range r.entries[key] {
		if operatorSignatureKey(existing.OperandTypes) == operatorSignatureKey(entry.OperandTypes) {
			return fmt.Errorf("operator already registered")
		}
	}

	r.entries[key] = append(r.entries[key], entry)
	return nil
}

// Lookup finds an operator overload matching the given operator and operand types.
// Returns the entry and true if found, nil and false otherwise.
func (r *OperatorRegistry) Lookup(operator string, operandTypes []string) (*OperatorEntry, bool) {
	if r == nil {
		return nil, false
	}
	key := strings.ToLower(operator)

	// First try exact match for performance
	for _, entry := range r.entries[key] {
		if operatorSignatureKey(entry.OperandTypes) == operatorSignatureKey(operandTypes) {
			return entry, true
		}
	}

	// Note: Assignment-compatible matching (for inheritance) is handled
	// in the interpreter layer since it requires class hierarchy information

	return nil, false
}

// Clone creates a deep copy of the operator registry.
func (r *OperatorRegistry) Clone() *OperatorRegistry {
	if r == nil {
		return NewOperatorRegistry()
	}
	clone := NewOperatorRegistry()
	for op, list := range r.entries {
		copied := make([]*OperatorEntry, len(list))
		copy(copied, list)
		clone.entries[op] = copied
	}
	return clone
}

// operatorSignatureKey generates a key for operator signature matching.
func operatorSignatureKey(operandTypes []string) string {
	return strings.Join(operandTypes, "|")
}

// ========== Conversion Registry ==========

// ConversionRegistry manages type conversions (implicit and explicit).
type ConversionRegistry struct {
	implicit map[string]*ConversionEntry
	explicit map[string]*ConversionEntry
}

// ConversionEntry represents a registered type conversion.
type ConversionEntry struct {
	From        string
	To          string
	BindingName string
	Implicit    bool
}

// NewConversionRegistry creates a new conversion registry.
func NewConversionRegistry() *ConversionRegistry {
	return &ConversionRegistry{
		implicit: make(map[string]*ConversionEntry),
		explicit: make(map[string]*ConversionEntry),
	}
}

// Register registers a new type conversion.
// Returns an error if a conversion with the same signature is already registered.
func (r *ConversionRegistry) Register(entry *ConversionEntry) error {
	if entry == nil {
		return fmt.Errorf("conversion entry cannot be nil")
	}
	key := conversionKey(entry.From, entry.To)

	if entry.Implicit {
		if _, exists := r.implicit[key]; exists {
			return fmt.Errorf("implicit conversion already registered")
		}
		r.implicit[key] = entry
	} else {
		if _, exists := r.explicit[key]; exists {
			return fmt.Errorf("explicit conversion already registered")
		}
		r.explicit[key] = entry
	}

	return nil
}

// FindImplicit finds an implicit conversion from one type to another.
// Returns the conversion entry and true if found, nil and false otherwise.
func (r *ConversionRegistry) FindImplicit(from, to string) (*ConversionEntry, bool) {
	if r == nil {
		return nil, false
	}
	entry, ok := r.implicit[conversionKey(from, to)]
	return entry, ok
}

// FindConversionPath uses BFS to find the shortest path of implicit conversions.
// Returns a slice of intermediate type names, or nil if no path exists.
// maxDepth limits the number of conversions in the chain.
func (r *ConversionRegistry) FindConversionPath(from, to string, maxDepth int) []string {
	if r == nil || maxDepth <= 0 {
		return nil
	}

	// Normalize type names
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)

	// Direct conversion check
	if _, ok := r.implicit[conversionKey(from, to)]; ok {
		return []string{from, to}
	}

	// BFS to find shortest conversion path
	type queueItem struct {
		currentType string
		path        []string
	}

	visited := make(map[string]bool)
	queue := []queueItem{{currentType: from, path: []string{from}}}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Check if path is too long
		if len(current.path) > maxDepth {
			continue
		}

		// Try all possible conversions from current type
		for _, entry := range r.implicit {
			// Check if this conversion starts from current type
			if strings.ToUpper(entry.From) == current.currentType {
				nextType := strings.ToUpper(entry.To)

				// Found target!
				if nextType == to {
					return append(current.path, nextType)
				}

				// Add to queue if not visited
				if !visited[nextType] {
					visited[nextType] = true
					newPath := make([]string, len(current.path)+1)
					copy(newPath, current.path)
					newPath[len(current.path)] = nextType
					queue = append(queue, queueItem{
						currentType: nextType,
						path:        newPath,
					})
				}
			}
		}
	}

	// No path found
	return nil
}

// conversionKey generates a key for conversion lookup.
func conversionKey(from, to string) string {
	return strings.ToUpper(from) + "->" + strings.ToUpper(to)
}
