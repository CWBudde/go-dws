package main

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/interp/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

func main() {
	ctx := interp.New(nil)
	args := []builtins.Value{
		&runtime.StringValue{Value: "%.2f"},
		&runtime.ArrayValue{Elements: []builtins.Value{&runtime.FloatValue{Value: 13.0}}},
	}
	res := builtins.Format(ctx, args)
	fmt.Printf("%T %s\n", res, res.String())
}
