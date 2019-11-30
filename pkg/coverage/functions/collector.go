package functions

import (
	log "github.com/sirupsen/logrus"
	"go/ast"
	"go/token"
)

type visitor struct {
	err       error
	fset      *token.FileSet
	functions []Function
}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	switch x := n.(type) {
	case *ast.FuncDecl:
		//body := x.Body
		name := x.Name.Name

		start := v.fset.Position(n.Pos())
		end := v.fset.Position(n.End())
		startLine := start.Line
		startCol := start.Column
		endLine := end.Line
		endCol := end.Column
		f := Function{
			Name:      name,
			StartLine: startLine,
			StartCol:  startCol,
			EndLine:   endLine,
			EndCol:    endCol,
		}

		//sc := &stmtCollector{}
		//if err := sc.collect(body, v.fset); err != nil {
		//  v.err = err
		//  return nil
		//}

		//stmts := sc.statements

		//log.Debugf("%v statements %v", f.Name, stmts)

		//convertedStmts := make([]Statement, 0, len(stmts))

		//for _, stmnt := range stmts {
		//  start := v.fset.Position(stmnt.Pos())
		//  end := v.fset.Position(stmnt.End())
		//  startLine := start.Line
		//  startCol := start.Column
		//  endLine := end.Line
		//  endCol := end.Column
		//  s := Statement{
		//    StartLine: startLine,
		//    StartCol:  startCol,
		//    EndLine:   endLine,
		//    EndCol:    endCol,
		//  }
		//  convertedStmts = append(convertedStmts, s)
		//}

		//f.Statements = convertedStmts
		v.functions = append(v.functions, f)
	default:
	}

	return v
}

func CollectFunctions(f *ast.File, fset *token.FileSet) ([]Function, error) {
	v := &visitor{fset: fset}
	ast.Walk(v, f)

	if v.err != nil {
		return nil, v.err
	}

	log.Debugf("visitor functions %v", v.functions)

	return v.functions, nil
}
