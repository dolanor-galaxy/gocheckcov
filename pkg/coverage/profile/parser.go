// Copyright © 2019 Cole Giovannoni Wippern
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package Profile

import (
	"go/token"

	"github.com/cvgw/cov-analyzer/pkg/coverage/statements"
	"golang.org/x/tools/cover"
)

type Parser struct {
	Fset     *token.FileSet
	FilePath string
	Profile  *cover.Profile
}

func (p Parser) RecordStatementCoverage(functions []statements.Function) []statements.Function {
	for fIdx, function := range functions {
		statements := function.Statements
		for sIdx, statement := range statements {
			for _, block := range p.Profile.Blocks {
				//name := function.name
				startLine := statement.StartLine
				startCol := statement.StartCol
				endLine := statement.EndLine
				endCol := statement.EndCol

				if block.StartLine > endLine || (block.StartLine == endLine && block.StartCol >= endCol) {
					// Block starts after the function statement ends
					continue
				}
				if block.EndLine < startLine || (block.EndLine == startLine && block.EndCol <= startCol) {
					// Block ends before the function statement starts
					continue
				}
				statement.ExecutedCount += block.Count
				statements[sIdx] = statement
				break
			}
		}
		function.Statements = statements
		functions[fIdx] = function
	}
	return functions
}
