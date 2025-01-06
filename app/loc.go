package app

import "fmt"

type Loc struct {
	FilePath string
	Row int
	Column int
}

func NewLoc(filePath string, row, column int) *Loc {
	return &Loc{
		FilePath: filePath,
		Row: row,
		Column: column,
	}
}

func (loc *Loc) Display() string {
	return fmt.Sprintf("%s:%d:%d", loc.FilePath, loc.Row + 1, loc.Column + 1)
}