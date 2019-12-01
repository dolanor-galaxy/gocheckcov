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
	"go/ast"
	"go/token"

	log "github.com/sirupsen/logrus"
)

func CollectFunctions(f *ast.File, fset *token.FileSet, filePath string) ([]Function, error) {
	functions := []Function{}

	for i := range f.Decls {
		switch x := f.Decls[i].(type) {
		case *ast.FuncDecl:
			name := x.Name.Name

			start := fset.Position(x.Pos())
			end := fset.Position(x.End())
			startLine := start.Line
			startCol := start.Column
			endLine := end.Line
			endCol := end.Column
			f := Function{
				Name:        name,
				StartLine:   startLine,
				StartCol:    startCol,
				EndLine:     endLine,
				EndCol:      endCol,
				SrcPath:     filePath,
				StartOffset: start.Offset,
				EndOffset:   end.Offset,
			}
			functions = append(functions, f)
		}
	}

	log.Debugf("found functions %v", functions)

	return functions, nil
}
