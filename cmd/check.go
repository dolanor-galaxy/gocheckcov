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
	"io/ioutil"
	"log"
	"os"

	"github.com/cvgw/cov-analyzer/pkg/coverage/analyzer"
	"github.com/cvgw/cov-analyzer/pkg/coverage/config"
	profile "github.com/cvgw/cov-analyzer/pkg/coverage/profile"
	"github.com/cvgw/cov-analyzer/pkg/coverage/statements"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// checkCmd represents the check command
var (
	configFile  string
	ProfileFile string
	checkCmd    = &cobra.Command{
		Use:   "check",
		Short: "Check whether pkg coverage meets specified minimum",
		//  Long: `A longer description that spans multiple lines and likely contains examples
		//and usage of using your command. For example:

		//Cobra is a CLI library for Go that empowers applications.
		//This application is a tool to generate the needed files
		//to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			filePath := ProfileFile
			packageToFunctions := analyzer.MapPackagesToFunctions(filePath)

			pkgToCoverage := reportCoverage(packageToFunctions)
			cfContent, err := ioutil.ReadFile(configFile)
			if err != nil {
				log.Printf("could not read config file %v %v", configFile, err)
				os.Exit(1)
			}

			verifyCoverage(cfContent, pkgToCoverage)
		},
	}
)

func init() {
	checkCmd.AddCommand(checkInitCmd)
	rootCmd.AddCommand(checkCmd)
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

func verifyCoverage(configFile []byte, pkgToCoverage map[string]float64) {
	cfg := config.ConfigFile{}
	if err := yaml.Unmarshal([]byte(configFile), &cfg); err != nil {
		log.Printf("could not unmarshal yaml for config file %v", err)
		os.Exit(1)
	}

	for _, pkg := range cfg.Packages {
		cov, ok := pkgToCoverage[pkg.Name]
		if !ok {
			log.Printf("could not find coverage for package %v", pkg)
			os.Exit(1)
		}
		if pkg.MinCoveragePercentage > cov {
			log.Printf("coverage %v%% for package %v did not meet minimum %v%%", cov, pkg.Name, pkg.MinCoveragePercentage)
			os.Exit(1)
		}
		log.Printf("coverage %v%% for package %v meets minimum %v%%", cov, pkg.Name, pkg.MinCoveragePercentage)
	}
}

func reportCoverage(packageToFunctions map[string][]statements.Function) map[string]float64 {
	pkgToCoverage := make(map[string]float64)
	pc := analyzer.NewPackageCoverages(packageToFunctions)
	for pkg, functions := range packageToFunctions {
		cov, ok := pc.Coverage(pkg)
		if !ok {
			log.Printf("could not get coverage for package %v", pkg)
			os.Exit(1)
		}
		profile.PrintReport(functions)
		log.Printf("coverage for pkg %v is %v%% (%v/%v statements)", pkg, cov.CoveragePercent, cov.ExecutedCount, cov.StatementCount)
		pkgToCoverage[pkg] = cov.CoveragePercent
	}
	return pkgToCoverage
}
