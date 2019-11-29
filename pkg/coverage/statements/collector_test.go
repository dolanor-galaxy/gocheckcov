package statements

import (
	"go/parser"
	"go/token"
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_CollectFunctions(t *testing.T) {
	g := NewGomegaWithT(t)

	file, err := ioutil.TempFile("", "profile.test")
	if err != nil {
		t.Errorf("could not create temp file")
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
	err = ioutil.WriteFile(file.Name(), []byte(profileFileContent), 0644)
	if err != nil {
		t.Errorf("could not write to temp file %v", err)
		t.FailNow()
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file.Name(), []byte(profileFileContent), 0)
	if err != nil {
		t.Errorf("could not create ast for file %v %v", file.Name(), err)
		t.FailNow()
	}

	funcs, err := CollectFunctions(f, fset)
	g.Expect(err).To(BeNil())
	g.Expect(funcs).ToNot(BeNil())
	g.Expect(funcs).To(HaveLen(1))
}
