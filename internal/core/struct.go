package core

type ImportGroup int

const (
	stdLib ImportGroup = iota
	thirdParty
	local
)

type ImportBlock struct {
	Doc []string
	End []string
}
