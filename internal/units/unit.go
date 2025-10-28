// Package units manages DWScript units (modules) and their dependencies.
// Units enable multi-file projects with proper encapsulation and namespace management.
package units

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// Unit represents a DWScript unit (module).
// A unit consists of an interface section (public declarations) and an
// implementation section (private implementation), along with optional
// initialization and finalization code.
//
// Example DWScript unit:
//
//	unit MyUnit;
//	interface
//	  uses OtherUnit;
//	  function Add(x, y: Integer): Integer;
//	implementation
//	  function Add(x, y: Integer): Integer;
//	  begin
//	    Result := x + y;
//	  end;
//	initialization
//	  // Setup code
//	finalization
//	  // Cleanup code
//	end.
type Unit struct {
	// InterfaceSection contains the public declarations (types, functions, etc.)
	// Symbols in the interface section are visible to units that use this unit.
	InterfaceSection *ast.BlockStatement
	// ImplementationSection contains the private implementation code.
	// Symbols here are only visible within this unit.
	ImplementationSection *ast.BlockStatement
	// InitializationSection contains code that runs when the unit is loaded.
	// Runs before the main program begins execution.
	InitializationSection *ast.BlockStatement
	// FinalizationSection contains cleanup code that runs when the program exits.
	// Runs in reverse order of initialization (last unit initialized is finalized first).
	FinalizationSection *ast.BlockStatement

	// Symbols is the symbol table containing exported symbols from the interface section.
	// Only symbols defined in the interface section are added to this table.
	Symbols *semantic.SymbolTable

	// Name is the unit's name (case-insensitive in DWScript)
	Name string

	// FilePath is the absolute path to the unit's source file.
	// Used for error reporting and relative path resolution.
	FilePath string

	// Uses lists the names of units imported by this unit.
	// These dependencies must be loaded before this unit can be used.
	Uses []string
}

// NewUnit creates a new Unit with the given name and file path.
// The name is normalized to lowercase for case-insensitive comparison.
func NewUnit(name, filePath string) *Unit {
	return &Unit{
		Name:     name,
		FilePath: filePath,
		Uses:     []string{},
		Symbols:  semantic.NewSymbolTable(),
	}
}

// NormalizedName returns the unit name in lowercase for case-insensitive comparison.
// DWScript is case-insensitive, so "MyUnit", "MYUNIT", and "myunit" are the same.
func (u *Unit) NormalizedName() string {
	return strings.ToLower(u.Name)
}

// HasDependency checks if this unit depends on another unit (directly).
// Returns true if unitName appears in the Uses list.
func (u *Unit) HasDependency(unitName string) bool {
	normalized := strings.ToLower(unitName)
	for _, dep := range u.Uses {
		if strings.ToLower(dep) == normalized {
			return true
		}
	}
	return false
}

// String returns a string representation of the unit for debugging.
func (u *Unit) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("unit %s;\n", u.Name))

	if len(u.Uses) > 0 {
		sb.WriteString("uses ")
		for i, dep := range u.Uses {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(dep)
		}
		sb.WriteString(";\n")
	}

	if u.InterfaceSection != nil {
		sb.WriteString("\ninterface\n")
		sb.WriteString("  // interface declarations\n")
	}

	if u.ImplementationSection != nil {
		sb.WriteString("\nimplementation\n")
		sb.WriteString("  // implementation code\n")
	}

	if u.InitializationSection != nil {
		sb.WriteString("\ninitialization\n")
		sb.WriteString("  // initialization code\n")
	}

	if u.FinalizationSection != nil {
		sb.WriteString("\nfinalization\n")
		sb.WriteString("  // finalization code\n")
	}

	sb.WriteString("end.")

	return sb.String()
}
