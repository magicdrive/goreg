package core

import "go/ast"

type ImportGroup int

const (
	stdLib ImportGroup = iota
	thirdParty
	local
)

type ImportPack struct {
	Entity      *ast.ImportSpec
	LineComment *ast.Comment
	Doc         []string
	End         string
	Alias       string
}
