package interp

import "testing"

func TestJSONAssociativeSerialization_Values(t *testing.T) {
	tests := []struct {
		name, source, want string
	}{
		{
			name:   "empty",
			source: `var m: array[String] of Integer; PrintLn(JSON.Stringify(m));`,
			want:   "{}\n",
		},
		{
			name: "integer lookup preserves value",
			source: `var v: JSONVariant = 1; var m: array[String] of Integer;
				m['test'] := v; PrintLn(m['test']); PrintLn(m['test'] + 1);
				PrintLn(JSON.Stringify(m['test'])); PrintLn(JSON.Stringify(m));`,
			want: "1\n2\n1\n{\"test\":1}\n",
		},
		{
			name: "JSON numeric and null values",
			source: `var m: array[String] of JSONVariant;
				m['v'] := 0; PrintLn(JSON.Stringify(m));
				m['v'] := -7; PrintLn(JSON.Stringify(m));
				m['v'] := 1.25; PrintLn(JSON.Stringify(m));
				m['v'] := 9007199254740993; PrintLn(JSON.Stringify(m));
				m['v'] := Null; PrintLn(JSON.Stringify(m));`,
			want: "{\"v\":0}\n{\"v\":-7}\n{\"v\":1.25}\n{\"v\":9007199254740993}\n{\"v\":null}\n",
		},
		{
			name: "nested map and array",
			source: `var inner: array[String] of array of Integer;
				inner['items'] := [1, 2];
				var outer: array[String] of array[String] of array of Integer;
				outer['nested'] := inner; PrintLn(JSON.Stringify(outer));`,
			want: "{\"nested\":{\"items\":[1,2]}}\n",
		},
		{
			name: "record containing map",
			source: `type TData = record Items: array[String] of Integer; end;
				var data: TData; data.Items['v'] := 3; PrintLn(JSON.Stringify(data));`,
			want: "{\"Items\":{\"v\":3}}\n",
		},
		{
			name: "array containing map",
			source: `var m: array[String] of Integer; m['v'] := 4;
				var a: array of array[String] of Integer; a.Add(m); PrintLn(JSON.Stringify(a));`,
			want: "[{\"v\":4}]\n",
		},
		{
			name: "object property getter",
			source: `type TData = class
				function GetValue: Integer; begin Result := 5; end;
				property Value: Integer read GetValue;
			end;
			var m: array[String] of TData; m['obj'] := new TData;
			PrintLn(JSON.Stringify(m));`,
			want: "{\"obj\":{\"Value\":5}}\n",
		},
		{
			name: "custom stringify and nil object",
			source: `type TData = class
				function Stringify: String; begin Result := '[7]'; end;
			end;
			var m: array[String] of TData; m['obj'] := new TData;
			PrintLn(JSON.Stringify(m)); m['obj'] := nil; PrintLn(JSON.Stringify(m));`,
			want: "{\"obj\":[7]}\n{\"obj\":null}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, tt.source), tt.want)
		})
	}
}

func TestJSONAssociativeSerialization_Keys(t *testing.T) {
	tests := []struct {
		name, source, want string
	}{
		{"string alias", `type TKey = String; var m: array[TKey] of Integer; m['MiXeD'] := 1;`, "{\"MiXeD\":1}\n"},
		{"integer alias", `type TKey = Integer; var m: array[TKey] of Integer; m[-2] := 1;`, "{\"-2\":1}\n"},
		{"float", `var m: array[Float] of Integer; m[1.25] := 1;`, "{\"1.25\":1}\n"},
		{"variant", `var m: array[Variant] of Integer; m[42] := 1;`, "{\"42\":1}\n"},
		{"record excluded", `type TKey = record X: Integer; end; var key: TKey; var m: array[TKey] of Integer; m[key] := 1;`, "{}\n"},
		{"object excluded", `var m: array[TObject] of Integer; m[new TObject] := 1;`, "{}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, tt.source+` PrintLn(JSON.Stringify(m));`), tt.want)
		})
	}
}

func TestJSONAssociativeSerialization_NamespaceMethods(t *testing.T) {
	source := `var m: array[String] of Integer; m['v'] := 9;
		PrintLn(JSON.Stringify(m));
		PrintLn(JSON.StringifyUTF8(m));
		PrintLn(JSON.Stringify(JSON.Serialize(m)));
		PrintLn(JSON.PrettyStringify(m, '  '));`
	assertOutput(t, runQuickwinScript(t, source), "{\"v\":9}\n{\"v\":9}\n{\"v\":9}\n{\r\n  \"v\" : 9\r\n}\n")
}

func TestJSONAssociativeSerialization_PreservesJSONOwnership(t *testing.T) {
	source := `var source := JSON.Parse('{"child":{"v":1}}');
		var m: array[String] of JSONVariant;
		m['first'] := source.child; m['second'] := source.child;
		var serialized := JSON.Serialize(m);
		PrintLn(JSON.Stringify(source));
		PrintLn(JSON.Stringify(serialized.first));
		PrintLn(JSON.Stringify(serialized.second));
		serialized.first.v := 2;
		PrintLn(JSON.Stringify(source));
		PrintLn(JSON.Stringify(m['first']));
		PrintLn(JSON.Stringify(serialized.second));
		var nested: array[String] of array of JSONVariant;
		nested['items'] := [source.child];
		PrintLn(JSON.Stringify(nested));
		PrintLn(JSON.Stringify(source));`
	assertOutput(t, runQuickwinScript(t, source), "{\"child\":{\"v\":1}}\n{\"v\":1}\n{\"v\":1}\n{\"child\":{\"v\":1}}\n{\"v\":1}\n{\"v\":1}\n{\"items\":[{\"v\":1}]}\n{\"child\":{\"v\":1}}\n")
}

func TestJSONAssociativeSerialization_GetterException(t *testing.T) {
	source := `var calls := 0;
		type TData = class
			function GetValue: Integer;
			begin
				Inc(calls);
				raise Exception.Create('getter failed');
			end;
			property Value: Integer read GetValue;
		end;
		var m: array[String] of TData;
		m['first'] := new TData; m['second'] := new TData;
		try
			PrintLn(JSON.Stringify(m));
		except
			on E: Exception do PrintLn(E.Message);
		end;
		PrintLn(calls);`
	assertOutput(t, runQuickwinScript(t, source), "getter failed\n1\n")
}
