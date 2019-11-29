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
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/cvgw/gocheckcov/pkg/coverage/statements"
	"github.com/pkg/errors"
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
		node, err := NodeFromFilePath(filePath, goSrcPath, fset)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("could not get node from file path %v", filePath))
		}
		filePathToNode[filePath] = node
	}

	return filePathToNode, nil
}

func NodeFromFilePath(filePath, goSrcPath string, fset *token.FileSet) (*ast.File, error) {
	pFilePath := filepath.Join(goSrcPath, filePath)

	src, err := ioutil.ReadFile(pFilePath)
	if err != nil {
		log.Printf("could not read file from profile %v %v", pFilePath, err)
		return nil, err
	}

	f, err := parser.ParseFile(fset, pFilePath, src, 0)
	if err != nil {
		log.Printf("could not parse file %v %v", pFilePath, err)
		return nil, err
	}
	return f, nil
}

type Parser struct {
	Fset     *token.FileSet
	FilePath string
	Profile  *cover.Profile
}

func (p Parser) RecordStatementCoverage(functions []statements.Function) []statements.Function {
	for fIdx, function := range functions {
		statements := function.Statements
		for sIdx, statement := range statements {
			for _, block := range p.Profile.Blocks {
				//name := function.name
				startLine := statement.StartLine
				startCol := statement.StartCol
				endLine := statement.EndLine
				endCol := statement.EndCol

				if block.StartLine > endLine || (block.StartLine == endLine && block.StartCol >= endCol) {
					// Block starts after the function statement ends
					continue
				}
				if block.EndLine < startLine || (block.EndLine == startLine && block.EndCol <= startCol) {
					// Block ends before the function statement starts
					continue
				}
				statement.ExecutedCount += block.Count
				statements[sIdx] = statement
				break
			}
		}
		function.Statements = statements
		functions[fIdx] = function
	}
	return functions
}
