package interp

import "testing"

// TestRecordCopyOnAssign covers DWScript record value semantics: every assignment
// of a record-typed value stores an independent copy, so a later mutation of the
// source is never observable through the destination (PLAN.md §3.3).
func TestRecordCopyOnAssign(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name: "simple variable assignment",
			source: `
type TPoint = record x, y : Integer; end;
var p : TPoint = (x: 1; y: 2);
var q : TPoint;
q := p;
p.y := 3;
PrintLn(IntToStr(q.x) + ',' + IntToStr(q.y));
`,
			want: "1,2\n",
		},
		{
			name: "record field member assignment",
			source: `
type TPoint = record x, y : Integer; end;
type TRec = record Inner : TPoint; end;
var p : TPoint = (x: 1; y: 2);
var r : TRec;
r.Inner := p;
p.y := 3;
PrintLn(IntToStr(r.Inner.x) + ',' + IntToStr(r.Inner.y));
`,
			want: "1,2\n",
		},
		{
			name: "record property member assignment",
			source: `
type TPoint = record x, y : Integer; end;
type TRec = record
      private FInner : TPoint;
      public property Inner : TPoint read FInner write FInner;
   end;
var p : TPoint = (x: 1; y: 2);
var r : TRec;
r.Inner := p;
p.y := 3;
PrintLn(IntToStr(r.Inner.x) + ',' + IntToStr(r.Inner.y));
`,
			want: "1,2\n",
		},
		{
			name: "object field member assignment",
			source: `
type TPoint = record x, y : Integer; end;
type THolder = class
      Inner : TPoint;
   end;
var p : TPoint = (x: 1; y: 2);
var h := THolder.Create;
h.Inner := p;
p.y := 3;
PrintLn(IntToStr(h.Inner.x) + ',' + IntToStr(h.Inner.y));
`,
			want: "1,2\n",
		},
		{
			name: "dynamic array element assignment",
			source: `
type TPoint = record x, y : Integer; end;
var p : TPoint = (x: 1; y: 2);
var a : array of TPoint;
a.SetLength(1);
a[0] := p;
p.y := 3;
PrintLn(IntToStr(a[0].x) + ',' + IntToStr(a[0].y));
`,
			want: "1,2\n",
		},
		{
			name: "nested record member assignment",
			source: `
type TPoint = record x, y : Integer; end;
type TInner = record P : TPoint; end;
type TOuter = record I : TInner; end;
var p : TPoint = (x: 1; y: 2);
var o : TOuter;
o.I.P := p;
p.y := 3;
PrintLn(IntToStr(o.I.P.x) + ',' + IntToStr(o.I.P.y));
`,
			want: "1,2\n",
		},
		{
			name: "reading a record field yields a copy",
			source: `
type TPoint = record x, y : Integer; end;
type TRec = record Inner : TPoint; end;
var r : TRec;
r.Inner.x := 1;
r.Inner.y := 2;
var q : TPoint;
q := r.Inner;
q.y := 9;
PrintLn(IntToStr(r.Inner.x) + ',' + IntToStr(r.Inner.y));
`,
			want: "1,2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRecordRegressionScript(t, tt.source)
			if got != tt.want {
				t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", tt.want, got)
			}
		})
	}
}
