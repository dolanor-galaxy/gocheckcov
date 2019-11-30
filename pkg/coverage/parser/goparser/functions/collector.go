// Copyright 2019 Cole Giovannoni Wippern
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package functions

import (
	"go/ast"
	"go/token"

	"github.com/cvgw/gocheckcov/pkg/coverage/parser/goparser/statements"
	log "github.com/sirupsen/logrus"
)

type visitor struct {
	err       error
	fset      *token.FileSet
	functions []Function
}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	switch x := n.(type) {
	case *ast.FuncDecl:
		body := x.Body
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

		sc := &statements.StmtCollector{}
		if err := sc.Collect(body, v.fset); err != nil {
			v.err = err
			return nil
		}

		stmts := sc.Statements

		log.Debugf("%v statements %v", f.Name, stmts)

		convertedStmts := make([]statements.Statement, 0, len(stmts))

		for _, stmnt := range stmts {
			start := v.fset.Position(stmnt.Pos())
			end := v.fset.Position(stmnt.End())
			startLine := start.Line
			startCol := start.Column
			endLine := end.Line
			endCol := end.Column
			s := statements.Statement{
				StartLine: startLine,
				StartCol:  startCol,
				EndLine:   endLine,
				EndCol:    endCol,
			}
			convertedStmts = append(convertedStmts, s)
		}

		f.Statements = convertedStmts
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
