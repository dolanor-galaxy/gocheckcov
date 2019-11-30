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

package analyzer

import (
	"fmt"
	"go/token"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cvgw/gocheckcov/pkg/coverage/parser/profile"
	. "github.com/onsi/gomega"
)

func Test_NewPackageCoverages(t *testing.T) {
	g := NewGomegaWithT(t)

	pkgToFuncs := map[string][]profile.FunctionCoverage{
		"github.com/foo/bar/pkg/baz": []profile.FunctionCoverage{},
	}

	p := NewPackageCoverages(pkgToFuncs)
	g.Expect(p).ToNot(BeNil())
}

func Test_PackageCoverages_Coverage(t *testing.T) {
	g := NewGomegaWithT(t)

	pkgToFuncs := map[string][]profile.FunctionCoverage{
		"github.com/foo/bar/pkg/baz": []profile.FunctionCoverage{},
	}

	p := NewPackageCoverages(pkgToFuncs)
	cov, ok := p.Coverage("github.com/foo/bar/pkg/baz")
	g.Expect(ok).To(BeTrue())
	g.Expect(cov.CoveragePercent).To(Equal(float64(100)))
}

func Test_MapPackagesToFunctions(t *testing.T) {
	g := NewGomegaWithT(t)

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Errorf("could not create temp dir")
		t.FailNow()
	}

	profileFileContent := `
package foo

func Meow(x, y int) bool {
  if x > y {
	  return true
  }
	return false
}
`
	goSrcPath := filepath.Join(dir, "src.go")

	err = ioutil.WriteFile(goSrcPath, []byte(profileFileContent), 0644)
	if err != nil {
		t.Errorf("could not write to temp file %v", err)
		t.FailNow()
	}

	profilePath := filepath.Join(dir, "profile.out")

	coverageContent := fmt.Sprintf("mode: set\ngithub.com/cvgw/cov-analyzer/pkg/coverage/config/config.go:21.66,22.31 1 1")
	log.Print(coverageContent)
	err = ioutil.WriteFile(profilePath, []byte(coverageContent), 0644)

	if err != nil {
		t.Errorf("could not write to temp file %v", err)
		t.FailNow()
	}

	fset := token.NewFileSet()

	res := MapPackagesToFunctions(profilePath, []string{goSrcPath}, fset, "")
	g.Expect(res).ToNot(BeNil())
	g.Expect(res).To(HaveLen(1))
	g.Expect(res).To(HaveKey(strings.TrimPrefix(dir, "/")))
}
