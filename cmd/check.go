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

// checkCmd represents the check command
var (
	configFile     string
	ProfileFile    string
	printFunctions bool
	checkCmd       = &cobra.Command{
		Use:   "check",
		Short: "Check whether pkg coverage meets specified minimum",
		//  Long: `A longer description that spans multiple lines and likely contains examples
		//and usage of using your command. For example:

		Run: func(cmd *cobra.Command, args []string) {
			profilePath := ProfileFile
			fset := token.NewFileSet()
			dir := "/Users/colewippern/Code/src/github.com/GoogleContainerTools/kaniko/pkg"
			projectFiles := filesForPath(dir)
			packageToFunctions := mapPackagesToFunctions(profilePath, projectFiles, fset)
			cfContent, err := ioutil.ReadFile(configFile)
			if err != nil {
				log.Printf("could not read config file %v %v", configFile, err)
				os.Exit(1)
			}

			reportCoverage(packageToFunctions, printFunctions, cfContent)
		},
	}
)

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

	packageToFunctions := make(map[string][]statements.Function, 0)
	for _, filePath := range projectFiles {
		node, err := analyzer.NodeFromFilePath(filePath, fset)
		if err != nil {
			log.Printf("could not retrieve node from filepath %v", err)
			os.Exit(1)
		}
		functions, err := statements.CollectFunctions(node, fset)
		if err != nil {
			panic(err)
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

func filesForPath(dir string) []string {
	goPath := build.Default.GOPATH
	files := make([]string, 0)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
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
	return files
}

func init() {
	checkCmd.AddCommand(checkInitCmd)
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().BoolVar(&printFunctions, "print-functions", false, "print coverage for individual functions")
	checkCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "path to configuration file")
	checkCmd.PersistentFlags().StringVarP(&ProfileFile, "profile-file", "p", "", "path to coverage profile file")
	checkCmd.MarkFlagRequired("config-file")
	checkCmd.MarkFlagRequired("profile-file")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func verifyCoverage(pkg config.ConfigPackage, cov float64) {
	if pkg.MinCoveragePercentage > cov {
		log.Printf("coverage %v%% for package %v did not meet minimum %v%%", cov, pkg.Name, pkg.MinCoveragePercentage)
		os.Exit(1)
	}
	log.Printf("coverage %v%% for package %v meets minimum %v%%", cov, pkg.Name, pkg.MinCoveragePercentage)
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
			profile.PrintReport(functions)
		}

		log.Printf("coverage for pkg %v is %v%% (%v/%v statements)", pkg, cov.CoveragePercent, cov.ExecutedCount, cov.StatementCount)
	}
	for pkg := range packageToFunctions {
		cov, ok := pc.Coverage(pkg)
		if !ok {
			log.Printf("could not get coverage for package %v", pkg)
			os.Exit(1)
		}

		cfg := config.ConfigFile{}
		if err := yaml.Unmarshal([]byte(configFile), &cfg); err != nil {
			log.Printf("could not unmarshal yaml for config file %v", err)
			os.Exit(1)
		}
		cfgPkg, ok := cfg.GetPackage(pkg)
		if !ok {
			continue
		}
		verifyCoverage(cfgPkg, cov.CoveragePercent)
	}
	return pkgToCoverage
}
