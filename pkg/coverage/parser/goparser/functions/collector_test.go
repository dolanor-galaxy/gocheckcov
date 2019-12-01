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

package functions

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

	funcs, err := CollectFunctions(f, fset, file.Name())
	g.Expect(err).To(BeNil())
	g.Expect(funcs).ToNot(BeNil())
	g.Expect(funcs).To(HaveLen(1))
}
