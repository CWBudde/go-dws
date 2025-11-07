package dwscript

import "testing"

const benchmarkLoopScript = `
var i: Integer := 0;
while i < 1000 do
begin
  i := i + 1;
end;
`

func BenchmarkCompileModes(b *testing.B) {
	b.Run("ast", func(b *testing.B) {
		engine, err := New(WithCompileMode(CompileModeAST))
		if err != nil {
			b.Fatalf("failed to create engine: %v", err)
		}

		program, err := engine.Compile(benchmarkLoopScript)
		if err != nil {
			b.Fatalf("Compile failed: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := engine.Run(program); err != nil {
				b.Fatalf("Run failed: %v", err)
			}
		}
	})

	b.Run("bytecode", func(b *testing.B) {
		engine, err := New(WithCompileMode(CompileModeBytecode))
		if err != nil {
			b.Fatalf("failed to create engine: %v", err)
		}

		program, err := engine.Compile(benchmarkLoopScript)
		if err != nil {
			b.Fatalf("Compile failed: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := engine.Run(program); err != nil {
				b.Fatalf("Run failed: %v", err)
			}
		}
	})
}
