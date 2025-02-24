package model

import "go/ast"

type ImportGroup int

const (
	StdLib ImportGroup = iota
	ThirdParty
	Organization
	Local
)

type ImportPack struct {
	Entity      *ast.ImportSpec
	LineComment *ast.Comment
	Doc         []string
	End         string
	Alias       string
}

const DefaultOrderString = "std,thirdparty,organization,local"

var DefaultOrder = []ImportGroup{StdLib, ThirdParty, Organization, Local}
