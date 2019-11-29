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

package analyzer

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"

	"github.com/cvgw/cov-analyzer/pkg/coverage/statements"
	"golang.org/x/tools/cover"
)

func NodesFromProfiles(profiles []*cover.Profile, fset *token.FileSet) (map[string]*ast.File, error) {
	filePaths := make([]string, 0)
	for _, prof := range profiles {
		filePaths = append(filePaths, prof.FileName)
	}

	filePathToNode := make(map[string]*ast.File)
	for _, filePath := range filePaths {
		node, err := NodeFromFilePath(filePath, fset)
		if err != nil {
			return nil, err
		}
		filePathToNode[filePath] = node
	}

	return filePathToNode, nil
}

func NodeFromFilePath(filePath string, fset *token.FileSet) (*ast.File, error) {
	goPath := build.Default.GOPATH
	pFilePath := filepath.Join(goPath, "src", filePath)

	src, err := ioutil.ReadFile(pFilePath)
	if err != nil {
		log.Fatal(fmt.Errorf("could not read file from profile %v %v", pFilePath, err))
	}

	f, err := parser.ParseFile(fset, pFilePath, src, 0)
	if err != nil {
		panic(err)
	}
	return f, nil
}

type PackageCoverages struct {
	coverages           map[string]coverage
	packagesToFunctions map[string][]statements.Function
}

type coverage struct {
	StatementCount  int
	ExecutedCount   int
	CoveragePercent float64
}

func (p *PackageCoverages) Coverage(pkg string) (coverage, bool) {
	cov, ok := p.coverages[pkg]
	return cov, ok
}

func NewPackageCoverages(packagesToFunctions map[string][]statements.Function) *PackageCoverages {
	pkgToCoverage := make(map[string]coverage)
	for pkg, functions := range packagesToFunctions {
		statementCount := 0
		executedCount := 0
		for _, function := range functions {
			for _, stmt := range function.Statements {
				statementCount++
				if stmt.ExecutedCount > 0 {
					executedCount++
				}
			}
		}
		var covPer float64
		if executedCount == 0 && statementCount == 0 {
			covPer = 100
		} else {
			covPer = float64(math.Floor((float64(executedCount)/float64(statementCount))*10000) / 100)
		}
		c := coverage{StatementCount: statementCount, ExecutedCount: executedCount, CoveragePercent: covPer}
		pkgToCoverage[pkg] = c
	}
	return &PackageCoverages{
		coverages: pkgToCoverage,
	}
}
