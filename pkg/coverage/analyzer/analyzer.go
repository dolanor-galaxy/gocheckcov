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
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"
	"strings"

	profile "github.com/cvgw/cov-analyzer/pkg/coverage/profile"
	"github.com/cvgw/cov-analyzer/pkg/coverage/statements"
	"golang.org/x/tools/cover"
)

func MapPackagesToFunctions(filePath string) map[string][]statements.Function {
	profiles, err := cover.ParseProfiles(filePath)
	if err != nil {
		log.Fatal(fmt.Errorf("could not parse profiles from %v %v", filePath, err))
	}

	goPath := build.Default.GOPATH
	packageToFunctions := make(map[string][]statements.Function)

	for _, prof := range profiles {
		pFilePath := filepath.Join(goPath, "src", prof.FileName)

		src, err := ioutil.ReadFile(pFilePath)
		if err != nil {
			log.Fatal(fmt.Errorf("could not read file from profile %v %v", pFilePath, err))
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, pFilePath, src, 0)
		if err != nil {
			panic(err)
		}

		p := profile.Parser{FilePath: pFilePath, Fset: fset, Profile: prof}
		functions, err := statements.CollectFunctions(f, fset)
		if err != nil {
			panic(err)
		}
		functions = p.RecordStatementCoverage(functions)

		pkg := strings.TrimPrefix(filepath.Dir(pFilePath), filepath.Join(goPath, "src"))
		pkg = strings.TrimPrefix(pkg, "/")
		packageToFunctions[pkg] = append(packageToFunctions[pkg], functions...)
	}
	return packageToFunctions
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
		covPer := float64(math.Floor((float64(executedCount)/float64(statementCount))*10000) / 100)
		c := coverage{StatementCount: statementCount, ExecutedCount: executedCount, CoveragePercent: covPer}
		pkgToCoverage[pkg] = c
	}
	return &PackageCoverages{
		coverages: pkgToCoverage,
	}
}
