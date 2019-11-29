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
