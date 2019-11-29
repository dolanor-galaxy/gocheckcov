// Copyright © 2019 Cole Giovannoni Wippern
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package statements

import (
	"fmt"
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

		sc := &stmtCollector{}
		if err := sc.collect(body, v.fset); err != nil {
			v.err = err
			return nil
		}
		stmts := sc.statements
		convertedStmts := make([]Statement, 0, len(stmts))
		for _, stmnt := range stmts {
			start := v.fset.Position(stmnt.Pos())
			end := v.fset.Position(stmnt.End())
			startLine := start.Line
			startCol := start.Column
			endLine := end.Line
			endCol := end.Column
			s := Statement{
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
	return v.functions, nil
}

type stmtCollector struct {
	statements []ast.Stmt
}

func (sc *stmtCollector) collect(s ast.Stmt, fset *token.FileSet) error {
	statements := []ast.Stmt{}
	switch s := s.(type) {
	case *ast.BlockStmt:
		statements = s.List
	case *ast.CaseClause:
		statements = s.Body
	case *ast.CommClause:
		statements = s.Body
	case *ast.ForStmt:
		if s.Init != nil {
			if err := sc.collect(s.Init, fset); err != nil {
				return err
			}
		}
		if s.Post != nil {
			if err := sc.collect(s.Post, fset); err != nil {
				return err
			}
		}
		if err := sc.collect(s.Body, fset); err != nil {
			return err
		}
	case *ast.IfStmt:
		if s.Init != nil {
			if err := sc.collect(s.Init, fset); err != nil {
				return err
			}
		}
		if err := sc.collect(s.Body, fset); err != nil {
			return err
		}
		if s.Else != nil {
			if err := sc.handleIfStmtElse(s, fset); err != nil {
				return err
			}
		}
	case *ast.LabeledStmt:
		if err := sc.collect(s.Stmt, fset); err != nil {
			return err
		}
	case *ast.RangeStmt:
		if err := sc.collect(s.Body, fset); err != nil {
			return err
		}
	case *ast.SelectStmt:
		if err := sc.collect(s.Body, fset); err != nil {
			return err
		}
	case *ast.SwitchStmt:
		if s.Init != nil {
			if err := sc.collect(s.Init, fset); err != nil {
				return err
			}
		}
		if err := sc.collect(s.Body, fset); err != nil {
			return err
		}
	case *ast.TypeSwitchStmt:
		if s.Init != nil {
			if err := sc.collect(s.Init, fset); err != nil {
				return err
			}
		}
		if err := sc.collect(s.Assign, fset); err != nil {
			return err
		}
		if err := sc.collect(s.Body, fset); err != nil {
			return err
		}
	}
	for i := 0; i < len(statements); i++ {
		s := (statements)[i]
		switch s.(type) {
		case *ast.CaseClause, *ast.CommClause, *ast.BlockStmt:
			// don't descend any deeper into the tree
			break
		default:
			sc.statements = append(sc.statements, s)
		}
		if err := sc.collect(s, fset); err != nil {
			return err
		}
	}
	return nil
}

func (sc *stmtCollector) handleIfStmtElse(s *ast.IfStmt, fset *token.FileSet) error {
	// Code copied from go.tools/cmd/cover, to deal with "if x {} else if y {}"
	// Copied from go.tools/cmd/cover
	// Handle "if x {} else if y {}"
	// AST doesn't record the location of else statements. Make
	// a reasonable guess
	const backupToElse = token.Pos(len("else "))
	switch stmt := s.Else.(type) {
	case *ast.IfStmt:
		block := &ast.BlockStmt{
			// Covered part probably starts at the "else"
			Lbrace: stmt.If - backupToElse,
			List:   []ast.Stmt{stmt},
			Rbrace: stmt.End(),
		}
		s.Else = block
	case *ast.BlockStmt:
		// Block probably starts at the "else"
		stmt.Lbrace -= backupToElse
	default:
		return fmt.Errorf("unexpected node type for if statement")
	}
	sc.collect(s.Else, fset)
	return nil
}
