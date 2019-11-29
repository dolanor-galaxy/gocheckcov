// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"go/token"
	"io/ioutil"
	"log"
	"os"

	"github.com/cvgw/cov-analyzer/pkg/coverage/analyzer"
	"github.com/cvgw/cov-analyzer/pkg/coverage/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// checkInitCmd represents the checkInit command
var checkInitCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		profilePath := ProfileFile
		fset := token.NewFileSet()
		dir := "/Users/colewippern/Code/src/github.com/GoogleContainerTools/kaniko/pkg"
		projectFiles, err := filesForPath(dir)
		if err != nil {
			log.Printf("could not retrieve files for path %v %v", dir, err)
			os.Exit(1)
		}

		packageToFunctions := mapPackagesToFunctions(profilePath, projectFiles, fset)
		pc := analyzer.NewPackageCoverages(packageToFunctions)

		configFile := config.ConfigFile{}
		for pkg := range packageToFunctions {
			cov, ok := pc.Coverage(pkg)
			if !ok {
				log.Printf("could not get coverage for package %v", pkg)
				os.Exit(1)
			}
			cfgPkg := config.ConfigPackage{MinCoveragePercentage: cov.CoveragePercent, Name: pkg}
			configFile.Packages = append(configFile.Packages, cfgPkg)
		}

		configContent, err := yaml.Marshal(configFile)
		if err != nil {
			log.Printf("couldn't marshal config file %v", err)
			os.Exit(1)
		}
		if err := ioutil.WriteFile("config.yaml", configContent, 0644); err != nil {
			log.Printf("could not read config file %v %v", configFile, err)
			os.Exit(1)
		}
	},
}

func init() {
	checkCmd.AddCommand(checkInitCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkInitCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkInitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
