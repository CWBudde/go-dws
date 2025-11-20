package interp

import "testing"

// ============================================================================ //
// Set initialization and membership basics (including range-like scenarios)
// ============================================================================ //

func TestSetUninitializedVariable(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "uninitialized inline set type",
			input: `
				type TColor = (Red, Green, Blue);
				var s: set of TColor;
				PrintLn('ok');
			`,
			expect: "ok\n",
		},
		{
			name: "multi-identifier set declaration",
			input: `
				type TColor = (Red, Green, Blue);
				var s1, s2, s3: set of TColor;
				PrintLn('ok');
			`,
			expect: "ok\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestSetInOperatorEmpty(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "in operator with empty set - should be false",
			input: `
				type TColor = (Red, Green, Blue);
				var s: set of TColor;
				if Red in s then
					PrintLn('found')
				else
					PrintLn('not found');
			`,
			expect: "not found\n",
		},
		{
			name: "multiple checks on empty set",
			input: `
				type TColor = (Red, Green, Blue);
				var s: set of TColor;
				var count := 0;
				if Red in s then count := count + 1;
				if Green in s then count := count + 1;
				if Blue in s then count := count + 1;
				PrintLn(count);
			`,
			expect: "0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestSetInOperatorAfterInclude(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "Include then check membership",
			input: `
				type TColor = (Red, Green, Blue);
				var s: set of TColor;
				s.Include(Red);
				if Red in s then
					PrintLn('found')
				else
					PrintLn('not found');
			`,
			expect: "found\n",
		},
		{
			name: "Include multiple then check all",
			input: `
				type TColor = (Red, Green, Blue);
				var s: set of TColor;
				s.Include(Red);
				s.Include(Blue);

				var count := 0;
				if Red in s then count := count + 1;
				if Green in s then count := count + 1;
				if Blue in s then count := count + 1;
				PrintLn(count);
			`,
			expect: "2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestSetMultiIdentifierSeparateInstances(t *testing.T) {
	input := `
		type TColor = (Red, Green, Blue);
		var s1, s2: set of TColor;

		s1.Include(Red);
		s2.Include(Blue);

		if Red in s1 then PrintLn('s1 has Red');
		if Blue in s1 then PrintLn('s1 has Blue');

		if Red in s2 then PrintLn('s2 has Red');
		if Blue in s2 then PrintLn('s2 has Blue');
	`
	_, output := testEvalWithOutput(input)
	expect := "s1 has Red\ns2 has Blue\n"
	if output != expect {
		t.Errorf("expected %q, got %q", expect, output)
	}
}

func TestSetForInEmpty(t *testing.T) {
	input := `
		type TColor = (Red, Green, Blue);
		var s: set of TColor;
		var count := 0;
		for var e in s do
			count := count + 1;
		PrintLn(count);
	`
	_, output := testEvalWithOutput(input)
	expect := "0\n"
	if output != expect {
		t.Errorf("expected %q, got %q", expect, output)
	}
}

// Covers the "set of range" pattern taken from the eratosthene fixture.
func TestSetInitializationEratosthenePattern(t *testing.T) {
	input := `
		type TRange = enum (Low = 2, High = 20);
		var sieve: set of TRange;

		var count := 0;
		for var e in TRange do begin
			if e in sieve then
				count := count + 1;
		end;
		PrintLn(count);

		sieve.Include(TRange.Low);
		count := 0;
		for var e in TRange do begin
			if e in sieve then
				count := count + 1;
		end;
		PrintLn(count);
	`
	_, output := testEvalWithOutput(input)
	expect := "0\n1\n"
	if output != expect {
		t.Errorf("expected %q, got %q", expect, output)
	}
}
