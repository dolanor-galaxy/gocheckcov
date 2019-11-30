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

package reporter

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/cvgw/gocheckcov/pkg/coverage/analyzer"
	"github.com/cvgw/gocheckcov/pkg/coverage/config"
	"github.com/cvgw/gocheckcov/pkg/coverage/parser/profile"
	"gopkg.in/yaml.v2"
)

type CliLogger struct{}

func (l CliLogger) Printf(fmtString string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(fmtString, args...))
}

type logger interface {
	Printf(string, ...interface{})
}

type Verifier struct {
	Out            logger
	MinCov         float64
	PrintFunctions bool
}

func (v Verifier) VerifyCoverage(pkg config.ConfigPackage, cov float64) bool {
	if pkg.MinCoveragePercentage > cov {
		v.Out.Printf(
			"coverage %v%% for package %v did not meet minimum %v%%",
			cov,
			pkg.Name,
			pkg.MinCoveragePercentage,
		)

		return false
	}

	v.Out.Printf(
		"coverage %v%% for package %v meets minimum %v%%",
		cov,
		pkg.Name,
		pkg.MinCoveragePercentage,
	)

	return true
}

func (v Verifier) PrintReport(functions []profile.FunctionCoverage) {
	for _, function := range functions {
		var executedStatementsCount int64

		executedStatementsCount += function.CoveredCount

		val := (float64(executedStatementsCount) / float64(function.StatementCount)) * 10000
		percent := (math.Floor(val) / 10000) * 100
		v.Out.Printf(
			"function %v has %v statements of which %v were executed for a percent of %v",
			function.Name,
			function.StatementCount,
			executedStatementsCount,
			percent,
		)
	}
}

func (v Verifier) ReportPackageCoverages(
	packageToFunctions map[string][]profile.FunctionCoverage,
	pc *analyzer.PackageCoverages,
	printFunctions bool,
) {
	for pkg := range packageToFunctions {
		functions := packageToFunctions[pkg]
		cov, ok := pc.Coverage(pkg)

		if !ok {
			log.Printf("could not get coverage for package %v", pkg)
			os.Exit(1)
		}

		if v.PrintFunctions {
			v.PrintReport(functions)
		}

		v.Out.Printf(
			"pkg %v coverage is %v%% (%v/%v statements)\n",
			pkg,
			cov.CoveragePercent,
			cov.ExecutedCount,
			cov.StatementCount,
		)
	}
}

func (v Verifier) ReportCoverage(
	packageToFunctions map[string][]profile.FunctionCoverage,
	printFunctions bool,
	configFile []byte,
) map[string]float64 {
	pkgToCoverage := make(map[string]float64)
	pc := analyzer.NewPackageCoverages(packageToFunctions)

	v.ReportPackageCoverages(packageToFunctions, pc, printFunctions)

	fail := false

	for pkg := range packageToFunctions {
		cov, ok := pc.Coverage(pkg)
		if !ok {
			log.Printf("could not get coverage for package %v", pkg)
			os.Exit(1)
		}

		var cfgPkg config.ConfigPackage

		if len(configFile) != 0 {
			cfg := config.ConfigFile{}
			if err := yaml.Unmarshal(configFile, &cfg); err != nil {
				log.Printf("could not unmarshal yaml for config file %v", err)
				os.Exit(1)
			}

			var ok bool
			cfgPkg, ok = cfg.GetPackage(pkg)

			if !ok {
				continue
			}
		} else {
			cfgPkg = config.ConfigPackage{
				Name:                  pkg,
				MinCoveragePercentage: v.MinCov,
			}
		}

		if ok := v.VerifyCoverage(cfgPkg, cov.CoveragePercent); !ok {
			fail = true
		}
	}

	if fail {
		os.Exit(1)
	}

	return pkgToCoverage
}
