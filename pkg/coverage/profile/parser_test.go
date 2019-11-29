package profile

import (
	"go/token"
	"io/ioutil"
	"testing"

	"github.com/cvgw/gocheckcov/pkg/coverage/statements"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/cover"
)

func Test_NodesFromProfiles(t *testing.T) {
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

	profiles := []*cover.Profile{
		&cover.Profile{FileName: file.Name()},
	}

	fset := token.NewFileSet()

	res, err := NodesFromProfiles("", profiles, fset)
	g.Expect(err).To(BeNil())
	g.Expect(res).To(HaveLen(1))
	g.Expect(res).To(HaveKey(file.Name()))
}

func Test_NodeFromProfile(t *testing.T) {
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

	res, err := NodeFromFilePath("", file.Name(), fset)
	g.Expect(err).To(BeNil())
	g.Expect(res).ToNot(BeNil())
}

func Test_Parser_RecordStatementCoverage(t *testing.T) {
	g := NewGomegaWithT(t)

	functions := []statements.Function{
		{StartCol: 2, StartLine: 4, EndCol: 4, EndLine: 7},
	}
	profile := &cover.Profile{
		Blocks: []cover.ProfileBlock{
			{StartCol: 2, StartLine: 4, EndCol: 4, EndLine: 7, Count: 2},
		},
	}

	filePath := "foo.go"
	fset := token.NewFileSet()

	p := Parser{
		Fset:     fset,
		FilePath: filePath,
		Profile:  profile,
	}

	expected := []statements.Function{}
	for _, f := range functions {
		statements := []statements.Statement{}
		copy(statements, f.Statements)
		for i, stmt := range statements {
			stmt.ExecutedCount = 2
			statements[i] = stmt
		}
		expected = append(expected, f)
	}

	res := p.RecordStatementCoverage(functions)
	g.Expect(res).To(HaveLen(1))
	g.Expect(res).To(ConsistOf(expected))
}
