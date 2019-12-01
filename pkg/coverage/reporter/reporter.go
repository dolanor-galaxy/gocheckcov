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
	"io"
	"math"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/cvgw/gocheckcov/pkg/coverage/analyzer"
	"github.com/cvgw/gocheckcov/pkg/coverage/config"
	"github.com/cvgw/gocheckcov/pkg/coverage/parser/profile"
	"gopkg.in/yaml.v2"
)

type CliLogger struct {
	Out io.Writer
}

func (l CliLogger) Printf(fmtString string, args ...interface{}) {
	if l.Out == nil {
		fmt.Println(fmt.Sprintf(fmtString, args...))
		return
	}

	if _, err := fmt.Fprintf(l.Out, fmtString, args...); err != nil {
		panic(fmt.Errorf("could not write to output %v", err))
	}
}

type logger interface {
	Printf(string, ...interface{})
}

type Verifier struct {
	Out            logger
	MinCov         float64
	PrintFunctions bool
}

func (v Verifier) VerifyCoverage(pkg config.ConfigPackage, pc *analyzer.PackageCoverages) (bool, error) {
	if pc == nil {
		err := fmt.Errorf("can't report coverages because coverage data is nil")
		log.Debug(err)

		return false, err
	}

	cov, ok := pc.Coverage(pkg.Name)

	if !ok {
		err := fmt.Errorf("could not get coverage for package %v", pkg)
		log.Debug(err)

		return false, err
	}
	v.Out.Printf(
		"pkg %v\tcoverage %v%% \tminimum %v%% \tstatements\t%v/%v\n",
		pkg.Name,
		cov.CoveragePercent,
		pkg.MinCoveragePercentage,
		cov.ExecutedCount,
		cov.StatementCount,
	)

	if pkg.MinCoveragePercentage > cov.CoveragePercent {
		return false, nil
	}

	return true, nil
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
) error {
	for pkg := range packageToFunctions {
		functions := packageToFunctions[pkg]

		if pc == nil {
			err := fmt.Errorf("can't report coverages because coverage data is nil")
			log.Debug(err)

			return err
		}

		cov, ok := pc.Coverage(pkg)

		if !ok {
			err := fmt.Errorf("could not get coverage for package %v", pkg)
			log.Debug(err)

			return err
		}

		if v.PrintFunctions {
			v.PrintReport(functions)
		}

		v.Out.Printf(
			"pkg %v\tcoverage %v%%\tstatements\t%v/%v\n",
			pkg,
			cov.CoveragePercent,
			cov.ExecutedCount,
			cov.StatementCount,
		)
	}

	return nil
}

func (v Verifier) ReportCoverage(
	packageToFunctions map[string][]profile.FunctionCoverage,
	printFunctions bool,
	configFile []byte,
) (map[string]float64, error) {
	pkgToCoverage := make(map[string]float64)
	pc := analyzer.NewPackageCoverages(packageToFunctions)

	fail := false

	for pkg := range packageToFunctions {

		var cfgPkg config.ConfigPackage

		if len(configFile) != 0 {
			cfg := config.ConfigFile{}
			if err := yaml.Unmarshal(configFile, &cfg); err != nil {
				err = errors.Wrap(err, "could not unmarshal yaml for config file %v")
				log.Debug(err)

				return nil, err
			}

			var ok bool
			cfgPkg, ok = cfg.GetPackage(pkg)

			if !ok {
				cfgPkg = config.ConfigPackage{
					Name:                  pkg,
					MinCoveragePercentage: cfg.MinCoveragePercentage,
				}
			}
		} else {
			cfgPkg = config.ConfigPackage{
				Name:                  pkg,
				MinCoveragePercentage: v.MinCov,
			}
		}

		ok, err := v.VerifyCoverage(cfgPkg, pc)
		if err != nil {
			return nil, err
		}

		if !ok {
			fail = true
		}
	}

	if fail {
		return nil, fmt.Errorf("packages failed to meet minimum coverage")
	}

	return pkgToCoverage, nil
}
