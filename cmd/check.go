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
	"bufio"
	"go/build"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"

	"github.com/cvgw/gocheckcov/pkg/coverage/analyzer"
	"github.com/cvgw/gocheckcov/pkg/coverage/config"
	"github.com/cvgw/gocheckcov/pkg/coverage/files"
	"github.com/cvgw/gocheckcov/pkg/coverage/reporter"
	"github.com/spf13/cobra"
)

var (
	noConfig       bool
	configFile     string
	ProfileFile    string
	printFunctions bool
	minCov         float64
	skipDirs       string
	checkCmd       = &cobra.Command{
		Use:   "check",
		Short: "Check whether pkg coverage meets specified minimum",
		Run: func(cmd *cobra.Command, args []string) {
			err := runCheckCommand(args)
			if err != nil {
				os.Exit(1)
			}
		},
	}
)

func runCheckCommand(args []string) error {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	srcPath := files.SetSrcPath(args)

	ignoreDirs := strings.Split(skipDirs, ",")
	dir := srcPath
	projectFiles, err := files.FilesForPath(dir, ignoreDirs)

	if err != nil {
		log.Printf("could not retrieve project files from path %v %v", dir, err)
		return err
	}

	profilePath := ProfileFile
	fset := token.NewFileSet()
	goSrc := filepath.Join(build.Default.GOPATH, "src")

	packageToFunctions, err := analyzer.MapPackagesToFunctions(profilePath, projectFiles, fset, goSrc)
	if err != nil {
		log.Print(err)
		return err
	}

	var cfContent []byte
	if !noConfig {
		cfContent, err = config.GetConfigFile(configFile)
		if err != nil {
			log.Debug(err)
			return err
		}
	}

	out := bufio.NewWriter(os.Stdout)

	defer func() {
		if err := out.Flush(); err != nil {
			log.Debug(err)
		}
	}()

	tabber := tabwriter.NewWriter(out, 1, 8, 1, '\t', 0)

	defer func() {
		if err := tabber.Flush(); err != nil {
			log.Debug(err)
		}
	}()

	cliL := &reporter.CliLogger{
		Out: tabber,
	}

	v := reporter.Verifier{
		Out:            cliL,
		PrintFunctions: printFunctions,
		MinCov:         minCov,
	}
	if _, err := v.ReportCoverage(packageToFunctions, printFunctions, cfContent); err != nil {
		cliL.Printf("%v", err)
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().BoolVar(&printFunctions, "print-functions", false, "print coverage for individual functions")

	checkCmd.Flags().BoolVar(&noConfig, "no-config", false, "do not read configuration from file")

	checkCmd.Flags().Float64VarP(
		&minCov,
		"minimum-coverage",
		"m",
		0,
		"minimum coverage percentage to enforce for all packages (defaults to 0)",
	)

	checkCmd.Flags().StringVarP(
		&configFile,
		"config-file",
		"c",
		"",
		"path to configuration file",
	)

	checkCmd.PersistentFlags().StringVarP(
		&skipDirs,
		"skip-dirs",
		"s",
		"vendor",
		"command separted list of directories to skip when reporting coverage",
	)

	checkCmd.PersistentFlags().StringVarP(&ProfileFile, "profile-file", "p", "", "path to coverage profile file")

	if err := checkCmd.MarkPersistentFlagRequired("profile-file"); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
