package functions

type Function struct {
	Name string
	//Statements  []statements.Statement
	StartOffset int
	StartLine   int
	StartCol    int
	EndOffset   int
	EndLine     int
	EndCol      int
}
