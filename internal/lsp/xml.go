package lsp

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/lexer"
)

func xmlFunctionTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	elem := xmlElementAt(prog.File, pos)
	if elem == nil {
		return nil
	}
	return functionInfoTarget(uri, prog.Info.XMLResolvedFunctions[elem])
}

func xmlElementAt(file *ast.File, pos position) *ast.XMLElement {
	var found *ast.XMLElement
	walkFileExprs(file, func(expr ast.Expr) {
		if found != nil {
			return
		}
		if elem, ok := expr.(*ast.XMLElement); ok && containsXMLElementTag(pos, elem) {
			found = elem
		}
	})
	return found
}

func containsXMLElementTag(pos position, elem *ast.XMLElement) bool {
	namePos := elem.NamePos
	if namePos.Line == 0 {
		namePos = xmlOpeningNamePosition(elem)
	}
	return containsSymbol(pos, namePos, elem.Tag) || containsSymbol(pos, elem.ClosePos, elem.Tag)
}

func xmlOpeningNamePosition(elem *ast.XMLElement) lexer.Position {
	return lexer.Position{
		Offset: elem.Pos.Offset + 1,
		Line:   elem.Pos.Line,
		Column: elem.Pos.Column + 1,
	}
}
