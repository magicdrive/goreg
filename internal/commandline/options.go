package commandline

import (
	"flag"

	"github.com/magicdrive/goreg/internal/model"
)

type Option struct {
	ImportOrder             []model.ImportGroup
	OrganizationName        string
	RemoveImportCommentFlag bool
	MinimizeGroupFlag       bool
	SortIncludeAliasFlag    bool
	WriteFlag               bool
	HelpFlag                bool
	VersionFlag             bool
	FileName                string
	ModulePath              string
	FlagSet                 *flag.FlagSet
}
