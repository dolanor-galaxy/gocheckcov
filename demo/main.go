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

package main

import (
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/cvgw/cov-analyzer/pkg/coverage/analyzer"
	"github.com/cvgw/cov-analyzer/pkg/coverage/config"
	profile "github.com/cvgw/cov-analyzer/pkg/coverage/profile"
	"github.com/cvgw/cov-analyzer/pkg/coverage/statements"
	"golang.org/x/tools/cover"
)

func main() {
	configFile := `
packages:
- name: github.com/GoogleContainerTools/kaniko/pkg/executor
  min_coverage_percentage: 80
`
	filePath := "cp.out"
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

	pkgToCoverage := reportCoverage(packageToFunctions)
	verifyCoverage([]byte(configFile), pkgToCoverage)
}

func verifyCoverage(configFile []byte, pkgToCoverage map[string]float64) {
	cfg := config.ConfigFile{}
	if err := yaml.Unmarshal([]byte(configFile), &cfg); err != nil {
		panic(err)
	}

	for _, pkg := range cfg.Packages {
		cov, ok := pkgToCoverage[pkg.Name]
		if !ok {
			log.Fatalf("could not find coverage for package %v", pkg)
		}
		if pkg.MinCoveragePercentage > cov {
			log.Fatalf("coverage %v%% for package %v did not meet minimum %v%%", cov, pkg.Name, pkg.MinCoveragePercentage)
		}
	}
}

func reportCoverage(packageToFunctions map[string][]statements.Function) map[string]float64 {
	pkgToCoverage := make(map[string]float64)
	pc := analyzer.NewPackageCoverages(packageToFunctions)
	for pkg, functions := range packageToFunctions {
		cov, ok := pc.Coverage(pkg)
		if !ok {
			log.Fatalf("could not get coverage for package %v", pkg)
		}
		profile.PrintReport(functions)
		log.Printf("coverage for pkg %v is %v%% (%v/%v statements)", pkg, cov.CoveragePercent, cov.ExecutedCount, cov.StatementCount)
		pkgToCoverage[pkg] = cov.CoveragePercent
	}
	return pkgToCoverage
}
