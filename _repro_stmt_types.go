package main

import (
"fmt"

"github.com/cwbudde/go-dws/internal/lexer"
"github.com/cwbudde/go-dws/internal/parser"
)

func main() {
src := `
type TCounter = class
class var Count: Integer;
end;

begin
TCounter.Count := 5;
end;
`
l := lexer.New(src)
p := parser.New(l)
program := p.ParseProgram()
if len(p.Errors()) > 0 {
fmt.Println("parser errors:")
for _, e := range p.Errors() {
fmt.Println(e)
}
return
}
fmt.Printf("statements: %d\n", len(program.Statements))
for i, s := range program.Statements {
fmt.Printf("%d: %T\n", i, s)
}
}
