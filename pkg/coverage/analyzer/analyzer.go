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
	"go/token"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/cvgw/gocheckcov/pkg/coverage/profile"
	"github.com/cvgw/gocheckcov/pkg/coverage/statements"
	log "github.com/sirupsen/logrus"
	"golang.org/x/tools/cover"
)

type PackageCoverages struct {
	coverages map[string]coverage
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
			covPer = math.Floor((float64(executedCount)/float64(statementCount))*10000) / 100
		}

		c := coverage{StatementCount: statementCount, ExecutedCount: executedCount, CoveragePercent: covPer}
		pkgToCoverage[pkg] = c
	}

	return &PackageCoverages{
		coverages: pkgToCoverage,
	}
}

func MapPackagesToFunctions(
	filePath string,
	projectFiles []string,
	fset *token.FileSet,
	goSrc string,
) map[string][]statements.Function {
	profiles, err := cover.ParseProfiles(filePath)
	if err != nil {
		log.Printf("could not parse profiles from %v %v", filePath, err)
		os.Exit(1)
	}

	filePathToProfileMap := make(map[string]*cover.Profile)
	for _, prof := range profiles {
		filePathToProfileMap[prof.FileName] = prof
	}

	packageToFunctions := make(map[string][]statements.Function)

	for _, filePath := range projectFiles {
		node, err := profile.NodeFromFilePath(filePath, goSrc, fset)
		if err != nil {
			log.Printf("could not retrieve node from filepath %v", err)
			os.Exit(1)
		}

		functions, err := statements.CollectFunctions(node, fset)
		if err != nil {
			log.Printf("could not collect functions for filepath %v %v", filePath, err)
			os.Exit(1)
		}

		log.Debugf("functions for file %v %v", filePath, functions)
		pkg := strings.TrimPrefix(filePath, fmt.Sprintf("%s/", goSrc))
		pkg = filepath.Dir(pkg)

		if prof, ok := filePathToProfileMap[filePath]; ok {
			p := profile.Parser{FilePath: filePath, Fset: fset, Profile: prof}
			functions = p.RecordStatementCoverage(functions)
		}

		packageToFunctions[pkg] = append(packageToFunctions[pkg], functions...)
	}

	log.Debugf("map of packages to functions %v", packageToFunctions)

	return packageToFunctions
}
