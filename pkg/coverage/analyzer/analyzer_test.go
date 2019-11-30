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
	"testing"

	"github.com/cvgw/gocheckcov/pkg/coverage/statements"
	. "github.com/onsi/gomega"
)

func Test_NewPackageCoverages(t *testing.T) {
	g := NewGomegaWithT(t)

	pkgToFuncs := map[string][]statements.Function{
		"github.com/foo/bar/pkg/baz": []statements.Function{},
	}

	p := NewPackageCoverages(pkgToFuncs)
	g.Expect(p).ToNot(BeNil())
}

func Test_PackageCoverages_Coverage(t *testing.T) {
	g := NewGomegaWithT(t)

	pkgToFuncs := map[string][]statements.Function{
		"github.com/foo/bar/pkg/baz": []statements.Function{},
	}

	p := NewPackageCoverages(pkgToFuncs)
	cov, ok := p.Coverage("github.com/foo/bar/pkg/baz")
	g.Expect(ok).To(BeTrue())
	g.Expect(cov.CoveragePercent).To(Equal(float64(100)))
}
