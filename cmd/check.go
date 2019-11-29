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

package cmd

import (
	"fmt"
	"go/build"
	"go/token"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cvgw/cov-analyzer/pkg/coverage/analyzer"
	"github.com/cvgw/cov-analyzer/pkg/coverage/config"
	profile "github.com/cvgw/cov-analyzer/pkg/coverage/profile"
	"github.com/cvgw/cov-analyzer/pkg/coverage/statements"
	"github.com/spf13/cobra"
	"golang.org/x/tools/cover"
	"gopkg.in/yaml.v2"
)

var (
	configFile     string
	ProfileFile    string
	printFunctions bool
	minCov         float64
	cliOutput      cliLogger
	skipDirs       dirsToIgnore
	checkCmd       = &cobra.Command{
		Use:   "check",
		Short: "Check whether pkg coverage meets specified minimum",
		Run: func(cmd *cobra.Command, args []string) {
			var srcPath string
			if len(args) > 0 {
				srcPath = args[0]
				log.Printf("srcPath %v", srcPath)
				absSrcPath, err := filepath.Abs(srcPath)
				if err != nil {
					log.Printf("could not get absolute path from %v %v", srcPath, err)
				} else {
					log.Printf("absSrcPath %v", absSrcPath)
					srcPath = absSrcPath
				}
			}

			if srcPath == "" {
				var err error
				srcPath, err = os.Getwd()
				if err != nil {
					log.Printf("could not get working directory %v", err)
				}
			}

			dir := srcPath
			projectFiles, err := filesForPath(dir)
			if err != nil {
				log.Printf("could not retrieve project files from path %v %v", dir, err)
				os.Exit(1)
			}

			profilePath := ProfileFile
			fset := token.NewFileSet()
			packageToFunctions := mapPackagesToFunctions(profilePath, projectFiles, fset)

			var cfContent []byte
			if configFile != "" {
				cfContent, err = ioutil.ReadFile(configFile)
				if err != nil {
					log.Printf("could not read config file %v %v", configFile, err)
					os.Exit(1)
				}
			}
			reportCoverage(packageToFunctions, printFunctions, cfContent)
		},
	}
)

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().BoolVar(&printFunctions, "print-functions", false, "print coverage for individual functions")

	checkCmd.Flags().Float64VarP(&minCov, "minimum-coverage", "m", 0, "minimum coverage percentage to enforce for all packages (defaults to 0)")

	checkCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "path to configuration file")

	checkCmd.PersistentFlags().StringVarP(&ProfileFile, "profile-file", "p", "", "path to coverage profile file")
	if err := checkCmd.MarkPersistentFlagRequired("profile-file"); err != nil {
		log.Print(err)
		os.Exit(1)
	}

	cliOutput = cliLogger{}
	skipDirs = []string{
		"vendor",
	}
}

func mapPackagesToFunctions(filePath string, projectFiles []string, fset *token.FileSet) map[string][]statements.Function {
	goPath := build.Default.GOPATH

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
		node, err := analyzer.NodeFromFilePath(filePath, fset)
		if err != nil {
			log.Printf("could not retrieve node from filepath %v", err)
			os.Exit(1)
		}
		functions, err := statements.CollectFunctions(node, fset)
		if err != nil {
			log.Printf("could not collect functions for filepath %v %v", filePath, err)
			os.Exit(1)
		}

		pkg := strings.TrimPrefix(filePath, fmt.Sprintf("%s/", filepath.Join(goPath, "src")))
		pkg = filepath.Dir(pkg)

		if prof, ok := filePathToProfileMap[filePath]; ok {
			p := profile.Parser{FilePath: filePath, Fset: fset, Profile: prof}
			functions = p.RecordStatementCoverage(functions)
		}

		packageToFunctions[pkg] = functions
	}
	return packageToFunctions
}

type dirsToIgnore []string

func (d dirsToIgnore) Includes(dir string) bool {
	for _, ignore := range d {
		if ignore == dir {
			return true
		}
	}

	return false
}

func filesForPath(dir string) ([]string, error) {
	goPath := build.Default.GOPATH
	files := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("could not access path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && skipDirs.Includes(info.Name()) {
			return filepath.SkipDir
		}

		if info.Mode().IsRegular() {
			if regexp.MustCompile(".go$").Match([]byte(path)) {
				if regexp.MustCompile("_test.go$").Match([]byte(path)) {
					return nil
				}
				path = strings.TrimPrefix(path, fmt.Sprintf("%v/", filepath.Join(goPath, "src")))
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, err
}

func verifyCoverage(pkg config.ConfigPackage, cov float64) bool {
	if pkg.MinCoveragePercentage > cov {
		cliOutput.Printf("coverage %v%% for package %v did not meet minimum %v%%", cov, pkg.Name, pkg.MinCoveragePercentage)
		return false
	}
	cliOutput.Printf("coverage %v%% for package %v meets minimum %v%%", cov, pkg.Name, pkg.MinCoveragePercentage)
	return true
}

func printReport(functions []statements.Function) {
	for _, function := range functions {
		executedStatementsCount := 0
		for _, s := range function.Statements {
			if s.ExecutedCount > 0 {
				executedStatementsCount++
			}
		}
		v := (float64(executedStatementsCount) / float64(len(function.Statements))) * 10000
		percent := (math.Floor(v) / 10000) * 100
		cliOutput.Printf("function %v has %v statements of which %v were executed for a percent of %v", function.Name, len(function.Statements), executedStatementsCount, percent)
	}
}

func reportCoverage(packageToFunctions map[string][]statements.Function, printFunctions bool, configFile []byte) map[string]float64 {
	pkgToCoverage := make(map[string]float64)
	pc := analyzer.NewPackageCoverages(packageToFunctions)
	for pkg := range packageToFunctions {
		functions := packageToFunctions[pkg]
		cov, ok := pc.Coverage(pkg)
		if !ok {
			log.Printf("could not get coverage for package %v", pkg)
			os.Exit(1)
		}
		if printFunctions {
			printReport(functions)
		}

		cliOutput.Printf("pkg %v coverage is %v%% (%v/%v statements)\n", pkg, cov.CoveragePercent, cov.ExecutedCount, cov.StatementCount)
	}
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
			if err := yaml.Unmarshal([]byte(configFile), &cfg); err != nil {
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
				MinCoveragePercentage: minCov,
			}
		}

		if ok := verifyCoverage(cfgPkg, cov.CoveragePercent); !ok {
			fail = true
		}
	}
	if fail {
		os.Exit(1)
	}
	return pkgToCoverage
}

type cliLogger struct{}

func (l cliLogger) Printf(fmtString string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(fmtString, args...))
}
