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
