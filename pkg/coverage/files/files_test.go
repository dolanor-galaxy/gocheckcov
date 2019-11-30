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

package files

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_SetSrcPath(t *testing.T) {
	type testcase struct {
		args        []string
		expected    string
		description string
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("could not get working directory %v", err)
		t.FailNow()
	}

	testCases := []testcase{
		testcase{
			description: "relative path with ...",
			args:        []string{"./pkg"},
			expected:    filepath.Join(cwd, "pkg"),
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.description, func(t *testing.T) {
			g := NewGomegaWithT(t)

			actual := SetSrcPath(tc.args)
			g.Expect(actual).To(Equal(tc.expected))
		})
	}
}
