// Copyright 2019 Cole Giovannoni Wippern
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reporter

import (
	"testing"

	"github.com/cvgw/gocheckcov/mocks/coverage/mock_reporter"
	"github.com/cvgw/gocheckcov/pkg/coverage/analyzer"
	"github.com/cvgw/gocheckcov/pkg/coverage/config"
	"github.com/cvgw/gocheckcov/pkg/coverage/parser/profile"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
)

func Test_Verifier_VerifyCoverage(t *testing.T) {
	g := NewGomegaWithT(t)

	type testcase struct {
		description string
		verifier    *Verifier
		pkg         config.ConfigPackage
		result      bool
	}

	type tcFn func(*gomock.Controller) testcase

	testCases := []tcFn{
		func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)
			mockLogger.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

			return testcase{
				description: "empty package",
				verifier:    &Verifier{Out: mockLogger},
				pkg:         config.ConfigPackage{},
				result:      true,
			}
		},
		func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)
			mockLogger.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

			return testcase{
				description: "cov is less than pkg min",
				verifier:    &Verifier{Out: mockLogger},
				pkg: config.ConfigPackage{
					MinCoveragePercentage: 100,
				},
			}
		},
	}

	for i := range testCases {
		ctrl := gomock.NewController(t)

		// Assert that Bar() is invoked.
		defer ctrl.Finish()

		tc := testCases[i](ctrl)
		t.Run(tc.description, func(t *testing.T) {
			v := tc.verifier
			pkg := tc.pkg
			ok := v.VerifyCoverage(pkg, 0)
			g.Expect(ok).To(Equal(tc.result))
		})
	}
}

func Test_Verifier_ReportCoverage(t *testing.T) {
	type testcase struct {
		verifier   *Verifier
		input      map[string][]profile.FunctionCoverage
		configData []byte
		printFuncs bool
		expectErr  bool
	}

	type tcFn func(*gomock.Controller) testcase

	testCases := map[string]tcFn{
		"empty function map and nil coverage data": func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)

			return testcase{
				verifier: &Verifier{Out: mockLogger},
				input:    map[string][]profile.FunctionCoverage{},
			}
		},
		"one empty function": func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)

			mockLogger.EXPECT().Printf(gomock.Any(), gomock.Any()).MinTimes(1)

			return testcase{
				verifier: &Verifier{Out: mockLogger},
				input: map[string][]profile.FunctionCoverage{
					"baz": []profile.FunctionCoverage{
						profile.FunctionCoverage{},
					},
				},
			}
		},
		"one function with one statement not in coverage data": func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)

			mockLogger.EXPECT().Printf(gomock.Any(), gomock.Any()).MinTimes(1)

			return testcase{
				verifier: &Verifier{Out: mockLogger},
				input: map[string][]profile.FunctionCoverage{
					"foo/bar": []profile.FunctionCoverage{
						{CoveredCount: 1, StatementCount: 1},
					},
				},
			}
		},
		"one function with one statement in coverage data": func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)

			mockLogger.EXPECT().Printf(gomock.Any(), gomock.Any()).MinTimes(1)

			return testcase{
				verifier: &Verifier{Out: mockLogger},
				input: map[string][]profile.FunctionCoverage{
					"foo/bar": []profile.FunctionCoverage{
						{CoveredCount: 1, StatementCount: 1},
					},
				},
				configData: []byte(`
packages:
- name: foo/bar
  min_coverage_percentage: 10
`),
			}
		},
		"one function with one statement with a bad config file": func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)

			mockLogger.EXPECT().Printf(gomock.Any(), gomock.Any()).MinTimes(1)

			return testcase{
				verifier: &Verifier{Out: mockLogger},
				input: map[string][]profile.FunctionCoverage{
					"foo/bar": []profile.FunctionCoverage{
						{CoveredCount: 1, StatementCount: 1},
					},
				},
				expectErr:  true,
				configData: []byte("meow"),
			}
		},
	}

	for description := range testCases {
		description := description

		t.Run(description, func(t *testing.T) {
			g := NewGomegaWithT(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tc := testCases[description](ctrl)
			v := tc.verifier
			_, err := v.ReportCoverage(tc.input, tc.printFuncs, tc.configData)
			if tc.expectErr {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func Test_Verifier_ReportPackageCoverages(t *testing.T) {
	type testcase struct {
		verifier   *Verifier
		input      map[string][]profile.FunctionCoverage
		covs       *analyzer.PackageCoverages
		printFuncs bool
		expectErr  bool
	}

	type tcFn func(*gomock.Controller) testcase

	testCases := map[string]tcFn{
		"empty function map and nil coverage data": func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)

			return testcase{
				verifier: &Verifier{Out: mockLogger},
				input:    map[string][]profile.FunctionCoverage{},
			}
		},
		"one empty function": func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)
			mockLogger.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

			return testcase{
				verifier: &Verifier{Out: mockLogger},
				input: map[string][]profile.FunctionCoverage{
					"baz": []profile.FunctionCoverage{
						profile.FunctionCoverage{},
					},
				},
				covs: analyzer.NewPackageCoverages(map[string][]profile.FunctionCoverage{
					"baz": []profile.FunctionCoverage{},
				}),
			}
		},
		"one function with one statement not in coverage data": func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)

			return testcase{
				verifier: &Verifier{Out: mockLogger},
				input: map[string][]profile.FunctionCoverage{
					"foo/bar": []profile.FunctionCoverage{
						{CoveredCount: 1, StatementCount: 1},
					},
				},
				covs:      &analyzer.PackageCoverages{},
				expectErr: true,
			}
		},
	}

	for description := range testCases {
		description := description

		t.Run(description, func(t *testing.T) {
			g := NewGomegaWithT(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tc := testCases[description](ctrl)
			v := tc.verifier
			err := v.ReportPackageCoverages(tc.input, tc.covs, tc.printFuncs)
			if tc.expectErr {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func Test_Verifier_PrintReport(t *testing.T) {
	type testcase struct {
		verifier  *Verifier
		functions []profile.FunctionCoverage
	}

	type tcFn func(*gomock.Controller) testcase

	testCases := map[string]tcFn{
		"empty function list": func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)

			return testcase{
				verifier:  &Verifier{Out: mockLogger},
				functions: []profile.FunctionCoverage{},
			}
		},
		"one empty function": func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)
			mockLogger.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

			return testcase{
				verifier: &Verifier{Out: mockLogger},
				functions: []profile.FunctionCoverage{
					profile.FunctionCoverage{},
				},
			}
		},
		"one function with one statement": func(ctrl *gomock.Controller) testcase {
			mockLogger := mock_reporter.NewMocklogger(ctrl)
			mockLogger.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

			return testcase{
				verifier: &Verifier{Out: mockLogger},
				functions: []profile.FunctionCoverage{
					{CoveredCount: 1, StatementCount: 1},
				},
			}
		},
	}

	for description := range testCases {
		description := description

		t.Run(description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tc := testCases[description](ctrl)
			v := tc.verifier
			v.PrintReport(tc.functions)
		})
	}
}
