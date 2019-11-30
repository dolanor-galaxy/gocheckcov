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

		log.Debugf("%v statements %v", f.Name, stmts)

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

	log.Debugf("visitor functions %v", v.functions)

	return v.functions, nil
}

type stmtCollector struct {
	statements []ast.Stmt
}

func (sc *stmtCollector) collect(s ast.Stmt, fset *token.FileSet) error {
	statements := []ast.Stmt{}

	switch s := s.(type) {
	case *ast.BlockStmt:
		if s == nil {
			return fmt.Errorf("something went wrong, block statement was nil")
		}

		statements = s.List
	case *ast.CaseClause:
		statements = s.Body
	case *ast.CommClause:
		statements = s.Body
	default:
		if err := sc.descend(s, fset); err != nil {
			return err
		}
	}

	if err := sc.filterStatements(statements, fset); err != nil {
		return err
	}

	return nil
}

func (sc *stmtCollector) filterStatements(statements []ast.Stmt, fset *token.FileSet) error {
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

func (sc *stmtCollector) descend(n ast.Node, fset *token.FileSet) error {
	var err error
	switch s := n.(type) {
	case *ast.ForStmt:
		err = sc.handleForStmt(s, fset)
	case *ast.IfStmt:
		err = sc.handleIfStmt(s, fset)
	case *ast.LabeledStmt:
		err = sc.collect(s.Stmt, fset)
	case *ast.RangeStmt:
		err = sc.collect(s.Body, fset)
	case *ast.SelectStmt:
		err = sc.collect(s.Body, fset)
	case *ast.SwitchStmt:
		if s.Init != nil {
			if e := sc.collect(s.Init, fset); e != nil {
				return e
			}
		}

		err = sc.collect(s.Body, fset)
	case *ast.TypeSwitchStmt:
		err = sc.handleTypeSwitchStmt(s, fset)
	}

	return err
}

func (sc *stmtCollector) handleTypeSwitchStmt(s *ast.TypeSwitchStmt, fset *token.FileSet) error {
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

	return nil
}

func (sc *stmtCollector) handleForStmt(s *ast.ForStmt, fset *token.FileSet) error {
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

	return nil
}

func (sc *stmtCollector) handleIfStmt(s *ast.IfStmt, fset *token.FileSet) error {
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

	if err := sc.collect(s.Else, fset); err != nil {
		return err
	}

	return nil
}
