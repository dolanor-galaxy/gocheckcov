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

package profile

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/cvgw/gocheckcov/pkg/coverage/parser/goparser"
	"github.com/cvgw/gocheckcov/pkg/coverage/parser/goparser/functions"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"golang.org/x/tools/cover"
)

func NodesFromProfiles(goSrcPath string, profiles []*cover.Profile, fset *token.FileSet) (map[string]*ast.File, error) {
	filePaths := make([]string, 0)

	for _, prof := range profiles {
		if prof.FileName == "" {
			return nil, fmt.Errorf("profile has a blank file name %v", prof)
		}

		filePaths = append(filePaths, prof.FileName)
	}

	filePathToNode := make(map[string]*ast.File)

	for _, filePath := range filePaths {
		node, err := goparser.NodeFromFilePath(filePath, goSrcPath, fset)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("could not get node from file path %v", filePath))
		}

		filePathToNode[filePath] = node
	}

	return filePathToNode, nil
}

type FunctionCoverage struct {
	StatementCount int64
	CoveredCount   int64
	Name           string
	Function       functions.Function
}

type Parser struct {
	Fset     *token.FileSet
	FilePath string
	Profile  *cover.Profile
}

func (p Parser) RecordFunctionCoverage(functions []functions.Function) []FunctionCoverage {
	out := make([]FunctionCoverage, 0, len(functions))

	for _, function := range functions {
		fc := FunctionCoverage{
			Name:     function.Name,
			Function: function,
		}

		if p.Profile != nil {
			fc = p.recordCoverageHits(fc, function)
		}

		if int(fc.StatementCount) != len(function.Statements) {
			log.Debugf(
				"function %v statement counts don't match Profile: %v AST: %v",
				function.Name,
				fc.StatementCount,
				len(function.Statements),
			)

			if int(fc.StatementCount) == 0 && len(function.Statements) > 0 {
				fc.StatementCount = int64(len(function.Statements))
			}
		}

		out = append(out, fc)
	}

	return out
}

func (p Parser) recordCoverageHits(fc FunctionCoverage, function functions.Function) FunctionCoverage {
	for _, block := range p.Profile.Blocks {
		startLine := function.StartLine
		startCol := function.StartCol
		endLine := function.EndLine
		endCol := function.EndCol

		if block.StartLine > endLine || (block.StartLine == endLine && block.StartCol >= endCol) {
			// Block starts after the function statement ends
			continue
		}

		if block.EndLine < startLine || (block.EndLine == startLine && block.EndCol <= startCol) {
			// Block ends before the function statement starts
			continue
		}

		fc.StatementCount += int64(block.NumStmt)
		if block.Count > 0 {
			fc.CoveredCount += int64(block.NumStmt)
		}
	}

	return fc
}

//func (p Parser) RecordStatementCoverage(functions []functions.Function) []functions.Function {
//  for fIdx, function := range functions {
//    statements := function.Statements
//    for sIdx, statement := range statements {
//      for _, block := range p.Profile.Blocks {
//        startLine := statement.StartLine
//        startCol := statement.StartCol
//        endLine := statement.EndLine
//        endCol := statement.EndCol

//        if block.StartLine > endLine || (block.StartLine == endLine && block.StartCol >= endCol) {
//          // Block starts after the function ends
//          continue
//        }

//        if block.EndLine < startLine || (block.EndLine == startLine && block.EndCol <= startCol) {
//          // Block ends before the function starts
//          continue
//        }

//        statement.ExecutedCount += block.Count
//        statements[sIdx] = statement

//        break
//      }
//    }

//    function.Statements = statements
//    functions[fIdx] = function
//  }

//  return functions
//}
