package interp

import "testing"

func TestOperatorOverloadNearestAncestor(t *testing.T) {
	for _, test := range []struct{ name, source, want string }{
		{"global", `
 type TBase = class end;
 type TMiddle = class(TBase) end;
 type TLeaf = class(TMiddle) end;
 function BaseOperator(value:TBase; n:Integer):String; begin Result:='base'; end;
 function MiddleOperator(value:TMiddle; n:Integer):String; begin Result:='middle'; end;
 operator + (TBase,Integer):String uses BaseOperator;
 operator + (TMiddle,Integer):String uses MiddleOperator;
 var value:=TLeaf.Create;
 PrintLn(value+1);
 `, "middle\n"},
		{"class", `
 type TBase = class
  function BaseOperator(n:Integer):String;
  class operator + (TBase,Integer):String uses BaseOperator;
 end;
 type TMiddle = class(TBase)
  function MiddleOperator(n:Integer):String;
  class operator + (TMiddle,Integer):String uses MiddleOperator;
 end;
 type TLeaf = class(TMiddle) end;
 function TBase.BaseOperator(n:Integer):String; begin Result:='base'; end;
 function TMiddle.MiddleOperator(n:Integer):String; begin Result:='middle'; end;
 var value:=TLeaf.Create;
 PrintLn(value+1);
 `, "middle\n"},
		{"left operand before right", `
 type TBase = class end;
 type TMiddle = class(TBase) end;
 type TLeaf = class(TMiddle) end;
 function RightNearest(left:TBase; right:TLeaf):String; begin Result:='right'; end;
 function LeftNearest(left:TMiddle; right:TBase):String; begin Result:='left'; end;
 operator + (TBase,TLeaf):String uses RightNearest;
 operator + (TMiddle,TBase):String uses LeftNearest;
 var left:=TLeaf.Create;
 var right:=TLeaf.Create;
 PrintLn(left+right);
 `, "left\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, output := testEvalWithOutput(test.source)
			if output != test.want {
				t.Fatalf("output=%q, want %q (result=%v)", output, test.want, result)
			}
		})
	}
}
