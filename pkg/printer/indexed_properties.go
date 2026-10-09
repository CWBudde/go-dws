package printer

import (
	"strings"

	"github.com/cwbudde/go-dws/pkg/ast"
)

func (p *Printer) printIndexExpression(ie *ast.IndexExpression) {
	indices := []ast.Expression{ie.Index}
	first := ie
	for first.CommaPos.Line != 0 {
		inner, ok := first.Left.(*ast.IndexExpression)
		if !ok {
			break
		}
		indices = append(indices, inner.Index)
		first = inner
	}
	p.printDWScript(first.Left)
	p.write("[")
	for i := len(indices) - 1; i >= 0; i-- {
		if i < len(indices)-1 {
			p.write(",")
			p.space()
		}
		p.printDWScript(indices[i])
	}
	p.write("]")
}

func (p *Printer) printPropertyDecl(pd *ast.PropertyDecl) {
	if pd.IsClassProperty {
		p.write("class")
		p.space()
	}
	p.write("property")
	p.space()
	p.printDWScript(pd.Name)
	if len(pd.IndexParams) > 0 {
		p.write("[")
		for i, param := range pd.IndexParams {
			if i > 0 {
				p.write(";")
				p.space()
			}
			p.printParameter(param)
		}
		p.write("]")
	}
	p.write(":")
	p.space()
	p.printDWScript(pd.Type)

	if pd.ReadSpec != nil {
		p.space()
		p.write("read")
		p.space()
		p.printDWScript(pd.ReadSpec)
	}

	if pd.WriteSpec != nil {
		p.space()
		p.write("write")
		p.space()
		p.printDWScript(pd.WriteSpec)
	}

	if pd.HasDescription {
		p.space()
		p.write("description \"")
		p.write(strings.ReplaceAll(pd.Description, "\"", "\"\""))
		p.write("\"")
	}

	if pd.IsReintroduce {
		p.space()
		p.write("reintroduce")
	}

	if pd.IsDefault {
		p.write(";")
		p.space()
		p.write("default")
	}
}
