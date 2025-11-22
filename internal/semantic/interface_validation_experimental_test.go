package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestInterfaceValidationWithExperimentalPasses tests interface type validation
// using the experimental multi-pass system (task 6.1.2.4)
func TestInterfaceValidationWithExperimentalPasses(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid class to interface assignment",
			input: `
				type IPrintable = interface
					function ToString(): String;
				end;

				type TDocument = class(TObject, IPrintable)
				public
					function ToString(): String;
					begin
						Result := 'Document';
					end;
				end;

				var printer: IPrintable;
				var doc: TDocument;
				doc := TDocument.Create();
				printer := doc;
			`,
			expectError: false,
		},
		{
			name: "invalid class to interface assignment",
			input: `
				type IPrintable = interface
					function ToString(): String;
				end;

				type TDocument = class(TObject)
				public
					function GetName(): String;
					begin
						Result := 'Document';
					end;
				end;

				var printer: IPrintable;
				var doc: TDocument;
				doc := TDocument.Create();
				printer := doc;
			`,
			expectError: true,
			errorMsg:    "cannot assign",
		},
		{
			name: "valid interface to interface assignment - same type",
			input: `
				type IBase = interface
					function GetValue(): Integer;
				end;

				type TImpl = class(TObject, IBase)
				public
					function GetValue(): Integer;
					begin
						Result := 42;
					end;
				end;

				var base1: IBase;
				var base2: IBase;
				var instance: TImpl;

				instance := TImpl.Create();
				base1 := instance;
				base2 := base1;
			`,
			expectError: false,
		},
		{
			name: "valid interface to interface assignment - inheritance",
			input: `
				type IBase = interface
					function GetValue(): Integer;
				end;

				type IDerived = interface(IBase)
					function GetExtra(): String;
				end;

				type TImpl = class(TObject, IDerived)
				public
					function GetValue(): Integer;
					begin
						Result := 42;
					end;

					function GetExtra(): String;
					begin
						Result := 'extra';
					end;
				end;

				var base: IBase;
				var derived: IDerived;
				var instance: TImpl;

				instance := TImpl.Create();
				derived := instance;
				base := derived;
			`,
			expectError: false,
		},
		{
			name: "invalid interface to interface assignment - unrelated",
			input: `
				type IReader = interface
					function ReadData(): String;
				end;

				type IWriter = interface
					procedure WriteData(data: String);
				end;

				type TReaderImpl = class(TObject, IReader)
				public
					function ReadData(): String;
					begin
						Result := 'data';
					end;
				end;

				var reader: IReader;
				var writer: IWriter;
				var instance: TReaderImpl;

				instance := TReaderImpl.Create();
				reader := instance;
				writer := reader;
			`,
			expectError: true,
			errorMsg:    "cannot assign",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			// Use experimental passes
			analyzer := NewAnalyzerWithExperimentalPasses()
			err := analyzer.Analyze(program)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got no error", tt.errorMsg)
				}
				if tt.errorMsg != "" && !ErrorMatches(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}
