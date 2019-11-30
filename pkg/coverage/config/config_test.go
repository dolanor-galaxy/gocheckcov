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

package config

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_ConfigFile_GetPackage(t *testing.T) {
	g := NewGomegaWithT(t)

	pkgs := []ConfigPackage{
		{Name: "github.com/foo/bar/pkg/baz",
			MinCoveragePercentage: 22,
		},
	}
	c := ConfigFile{
		Packages: pkgs,
	}

	pkg, ok := c.GetPackage("github.com/foo/bar/pkg/baz")
	g.Expect(ok).To(BeTrue())
	g.Expect(pkg).To(Equal(pkgs[0]))
}
