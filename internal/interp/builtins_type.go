package interp

import (
	"math"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Type Introspection Built-in Functions
// TypeOf, TypeOfClass, High, Low, and related helpers

// builtinLow implements the Low() built-in function.
// It returns the lower bound of an array or the lowest value of an enum type.
func (i *Interpreter) builtinLow(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Low() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle type meta-values (type names as values)
	if typeMetaVal, ok := arg.(*TypeMetaValue); ok {
		// Handle built-in types
		switch typeMetaVal.TypeInfo {
		case types.INTEGER:
			return &IntegerValue{Value: math.MinInt64}
		case types.FLOAT:
			return &FloatValue{Value: -math.MaxFloat64}
		case types.BOOLEAN:
			return &BooleanValue{Value: false}
		}

		// Handle enum types
		if enumType, ok := typeMetaVal.TypeInfo.(*types.EnumType); ok {
			if len(enumType.OrderedNames) == 0 {
				return i.newErrorWithLocation(i.currentNode, "enum type '%s' has no values", typeMetaVal.TypeName)
			}

			// Return the first enum value
			firstValueName := enumType.OrderedNames[0]
			firstOrdinal := enumType.Values[firstValueName]

			return &EnumValue{
				TypeName:     typeMetaVal.TypeName,
				ValueName:    firstValueName,
				OrdinalValue: firstOrdinal,
			}
		}

		return i.newErrorWithLocation(i.currentNode, "Low() not supported for type %s", typeMetaVal.TypeName)
	}

	// Handle array values
	if arrayVal, ok := arg.(*ArrayValue); ok {
		if arrayVal.ArrayType == nil {
			return i.newErrorWithLocation(i.currentNode, "array has no type information")
		}

		// For static arrays, return LowBound
		// For dynamic arrays, return 0
		if arrayVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrayVal.ArrayType.LowBound)}
		}
		return &IntegerValue{Value: 0}
	}

	// Handle enum values
	if enumVal, ok := arg.(*EnumValue); ok {
		// Look up the enum type metadata
		enumTypeKey := "__enum_type_" + ident.Normalize(enumVal.TypeName)
		typeVal, ok := i.env.Get(enumTypeKey)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "enum type '%s' not found", enumVal.TypeName)
		}

		enumTypeVal, ok := typeVal.(*EnumTypeValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "invalid enum type metadata for '%s'", enumVal.TypeName)
		}

		enumType := enumTypeVal.EnumType
		if len(enumType.OrderedNames) == 0 {
			return i.newErrorWithLocation(i.currentNode, "enum type '%s' has no values", enumVal.TypeName)
		}

		// Return the first enum value
		firstValueName := enumType.OrderedNames[0]
		firstOrdinal := enumType.Values[firstValueName]

		return &EnumValue{
			TypeName:     enumVal.TypeName,
			ValueName:    firstValueName,
			OrdinalValue: firstOrdinal,
		}
	}

	return i.newErrorWithLocation(i.currentNode, "Low() expects array, enum, or type name, got %s", arg.Type())
}

// builtinHigh implements the High() built-in function.
// It returns the upper bound of an array or the highest value of an enum type.
func (i *Interpreter) builtinHigh(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "High() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle type meta-values (type names as values)
	if typeMetaVal, ok := arg.(*TypeMetaValue); ok {
		// Handle built-in types
		switch typeMetaVal.TypeInfo {
		case types.INTEGER:
			return &IntegerValue{Value: math.MaxInt64}
		case types.FLOAT:
			return &FloatValue{Value: math.MaxFloat64}
		case types.BOOLEAN:
			return &BooleanValue{Value: true}
		}

		// Handle enum types
		if enumType, ok := typeMetaVal.TypeInfo.(*types.EnumType); ok {
			if len(enumType.OrderedNames) == 0 {
				return i.newErrorWithLocation(i.currentNode, "enum type '%s' has no values", typeMetaVal.TypeName)
			}

			// Return the last enum value
			lastValueName := enumType.OrderedNames[len(enumType.OrderedNames)-1]
			lastOrdinal := enumType.Values[lastValueName]

			return &EnumValue{
				TypeName:     typeMetaVal.TypeName,
				ValueName:    lastValueName,
				OrdinalValue: lastOrdinal,
			}
		}

		return i.newErrorWithLocation(i.currentNode, "High() not supported for type %s", typeMetaVal.TypeName)
	}

	// Handle array values
	if arrayVal, ok := arg.(*ArrayValue); ok {
		if arrayVal.ArrayType == nil {
			return i.newErrorWithLocation(i.currentNode, "array has no type information")
		}

		// For static arrays, return HighBound
		// For dynamic arrays, return Length - 1
		if arrayVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrayVal.ArrayType.HighBound)}
		}
		// Dynamic array: High = Length - 1
		return &IntegerValue{Value: int64(len(arrayVal.Elements) - 1)}
	}

	// Handle enum values
	if enumVal, ok := arg.(*EnumValue); ok {
		// Look up the enum type metadata
		enumTypeKey := "__enum_type_" + ident.Normalize(enumVal.TypeName)
		typeVal, ok := i.env.Get(enumTypeKey)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "enum type '%s' not found", enumVal.TypeName)
		}

		enumTypeVal, ok := typeVal.(*EnumTypeValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "invalid enum type metadata for '%s'", enumVal.TypeName)
		}

		enumType := enumTypeVal.EnumType
		if len(enumType.OrderedNames) == 0 {
			return i.newErrorWithLocation(i.currentNode, "enum type '%s' has no values", enumVal.TypeName)
		}

		// Return the last enum value
		lastValueName := enumType.OrderedNames[len(enumType.OrderedNames)-1]
		lastOrdinal := enumType.Values[lastValueName]

		return &EnumValue{
			TypeName:     enumVal.TypeName,
			ValueName:    lastValueName,
			OrdinalValue: lastOrdinal,
		}
	}

	return i.newErrorWithLocation(i.currentNode, "High() expects array, enum, or type name, got %s", arg.Type())
}

// builtinTypeOf implements the TypeOf() built-in function.
//
// Returns runtime type information for the given value.
// Can accept any value (object, class, primitive, etc.)
func (i *Interpreter) builtinTypeOf(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "TypeOf() expects exactly 1 argument, got %d", len(args))
	}

	val := args[0]

	// Get type ID and type name based on value type
	typeID, typeName := i.getTypeIDAndName(val)

	return &RTTITypeInfoValue{
		TypeID:   typeID,
		TypeName: typeName,
		TypeInfo: i.getValueType(val),
	}
}

// builtinTypeOfClass implements the TypeOfClass() built-in function.
//
// Returns type information for a class reference (metaclass).
func (i *Interpreter) builtinTypeOfClass(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "TypeOfClass() expects exactly 1 argument, got %d", len(args))
	}

	val := args[0]

	// TypeOfClass expects a class reference (ClassValue or ClassInfoValue)
	var classInfo *ClassInfo
	switch v := val.(type) {
	case *ClassValue:
		classInfo = v.ClassInfo
	case *ClassInfoValue:
		classInfo = v.ClassInfo
	default:
		return i.newErrorWithLocation(i.currentNode, "TypeOfClass() expects a class reference, got %s", val.Type())
	}

	if classInfo == nil {
		return i.newErrorWithLocation(i.currentNode, "TypeOfClass() received nil class reference")
	}

	// Generate type ID for the class
	typeID := i.getClassTypeID(classInfo.Name)

	return &RTTITypeInfoValue{
		TypeID:   typeID,
		TypeName: classInfo.Name,
		TypeInfo: nil, // Could be enhanced with class type metadata
	}
}

// getTypeIDAndName returns a unique type ID and human-readable type name for a value.
func (i *Interpreter) getTypeIDAndName(val Value) (int, string) {
	switch v := val.(type) {
	case *IntegerValue:
		return 1, "Integer"
	case *FloatValue:
		return 2, "Float"
	case *StringValue:
		return 3, "String"
	case *BooleanValue:
		return 4, "Boolean"
	case *InterfaceInstance:
		// For interface instances, return the underlying object's type
		if v.Object != nil && v.Object.Class != nil {
			return i.getClassTypeID(v.Object.Class.Name), v.Object.Class.Name
		}
		// If no underlying object, return the interface type name
		if v.Interface != nil {
			return i.getClassTypeID(v.Interface.Name), v.Interface.Name
		}
		return 100, "IInterface"
	case *ObjectInstance:
		if v.Class != nil {
			return i.getClassTypeID(v.Class.Name), v.Class.Name
		}
		return 100, "TObject"
	case *ClassValue:
		if v.ClassInfo != nil {
			return i.getClassTypeID(v.ClassInfo.Name), v.ClassInfo.Name
		}
		return 100, "TObject"
	case *ClassInfoValue:
		if v.ClassInfo != nil {
			return i.getClassTypeID(v.ClassInfo.Name), v.ClassInfo.Name
		}
		return 100, "TObject"
	case *ArrayValue:
		return 10, "Array"
	case *RecordValue:
		if v.RecordType != nil {
			return i.getRecordTypeID(v.RecordType.Name), v.RecordType.Name
		}
		return 20, "Record"
	case *EnumValue:
		return i.getEnumTypeID(v.TypeName), v.TypeName
	case *SetValue:
		return 30, "Set"
	case *VariantValue:
		// For Variant, return the type of the contained value
		if v.Value != nil {
			return i.getTypeIDAndName(v.Value)
		}
		return 40, "Variant"
	case *NilValue:
		return 0, "Nil"
	case *TypeMetaValue:
		return 50, v.TypeName
	default:
		return 999, "Unknown"
	}
}

// getClassTypeID returns a unique type ID for a class name.
// Type IDs start at 1000 for classes.
// Uses a registry to ensure unique IDs and handles case-insensitivity.
func (i *Interpreter) getClassTypeID(className string) int {
	// Normalize for case-insensitive comparison
	normalizedName := ident.Normalize(className)

	// Check if we already have an ID for this class
	if id, exists := i.classTypeIDRegistry[normalizedName]; exists {
		return id
	}

	// Assign new ID and store in registry
	id := i.nextClassTypeID
	i.classTypeIDRegistry[normalizedName] = id
	i.nextClassTypeID++
	return id
}

// getRecordTypeID returns a unique type ID for a record name.
// Type IDs start at 200000 for records.
// Uses a registry to ensure unique IDs and handles case-insensitivity.
func (i *Interpreter) getRecordTypeID(recordName string) int {
	// Normalize for case-insensitive comparison
	normalizedName := ident.Normalize(recordName)

	// Check if we already have an ID for this record
	if id, exists := i.recordTypeIDRegistry[normalizedName]; exists {
		return id
	}

	// Assign new ID and store in registry
	id := i.nextRecordTypeID
	i.recordTypeIDRegistry[normalizedName] = id
	i.nextRecordTypeID++
	return id
}

// getEnumTypeID returns a unique type ID for an enum name.
// Type IDs start at 300000 for enums.
// Uses a registry to ensure unique IDs and handles case-insensitivity.
func (i *Interpreter) getEnumTypeID(enumName string) int {
	// Normalize for case-insensitive comparison
	normalizedName := ident.Normalize(enumName)

	// Check if we already have an ID for this enum
	if id, exists := i.enumTypeIDRegistry[normalizedName]; exists {
		return id
	}

	// Assign new ID and store in registry
	id := i.nextEnumTypeID
	i.enumTypeIDRegistry[normalizedName] = id
	i.nextEnumTypeID++
	return id
}
