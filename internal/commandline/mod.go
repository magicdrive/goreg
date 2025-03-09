package commandline

import (
	"flag"
	"fmt"
	"os"

	_ "embed"

	"github.com/magicdrive/goreg/internal/common"
	"github.com/magicdrive/goreg/internal/model"
)

//go:embed help.txt
var helpMessage string

func OptParse(args []string) (int, *Option, error) {

	optLength := len(args)

	cfg, _ := common.LoadConfig()

	fs := flag.NewFlagSet("goreg", flag.ExitOnError)

	/* ------------------ */
	/* cfg Import section */
	/* ------------------ */

	// --order
	orderOpt := fs.String("order", cfg.Import.Order, "Specify module group order.")
	fs.StringVar(orderOpt, "o", cfg.Import.Order, "Specify module group order.")

	// --organization
	organizationOpt := fs.String("organization", cfg.Import.OrganizationModule, "Specify organization modulepath.")
	fs.StringVar(organizationOpt, "n", cfg.Import.OrganizationModule, "Specify organization modulepath.")

	// --local
	modulePathOpt := fs.String("local", cfg.Import.LocalModule, "Specify local modulepath.")
	fs.StringVar(modulePathOpt, "l", cfg.Import.LocalModule, "Specify local modulepath.")

	/* ------------------ */
	/* cfg Format section */
	/* ------------------ */

	// --minimize-group
	minimizeGroupOpt := fs.Bool("minimize-group", cfg.Format.MinimizeGroup, "Not separate module group by alias.")
	fs.BoolVar(minimizeGroupOpt, "m", cfg.Format.MinimizeGroup, "Not separate module group by alias.")

	// --sort-include-aliases
	sortIncludeAliasOpt := fs.Bool("sort-include-alias",
		cfg.Format.SortIncludeAlias, "Imports with aliases will also be sorted within the group.")
	fs.BoolVar(sortIncludeAliasOpt, "a", cfg.Format.SortIncludeAlias,
		"Imports with aliases will also be sorted within the group.")

	// --remove-import-comment
	removeImportCommentOpt := fs.Bool("remove-import-comment",
		cfg.Format.RemoveImportComment, "Remove the comments in the import.")
	fs.BoolVar(removeImportCommentOpt, "r", cfg.Format.RemoveImportComment, "Remove the comments in the import.")

	// --write
	writeFlagOpt := fs.Bool("write", false, "Show help message.")
	fs.BoolVar(writeFlagOpt, "w", false, "Show help message.")

	// --help
	helpFlagOpt := fs.Bool("help", false, "Show help message.")
	fs.BoolVar(helpFlagOpt, "h", false, "Show help message.")

	// --version
	versionFlagOpt := fs.Bool("version", false, "Show version.")
	fs.BoolVar(versionFlagOpt, "v", false, "Show version.")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "\nHelpOption:")
		fmt.Fprintln(os.Stderr, "    goreg --help")
	}
	err := fs.Parse(args)
	if err != nil {
		return optLength, nil, err
	}

	var filename = ""
	_args := fs.Args()
	if len(_args) > 0 {
		filename = _args[0]
	}

	var _importOrder []model.ImportGroup
	if *orderOpt == "" {
		_importOrder = model.DefaultOrder
	} else {
		_importOrder, err = GenerateOrderStrings(*orderOpt)
		if err != nil {
			return optLength, nil, err
		}
	}

	result := &Option{
		ImportOrder:          _importOrder,
		OrganizationName:     *organizationOpt,
		MinimizeGroupFlag:    *minimizeGroupOpt,
		SortIncludeAliasFlag: *sortIncludeAliasOpt,
		WriteFlag:            *writeFlagOpt,
		HelpFlag:             *helpFlagOpt,
		VersionFlag:          *versionFlagOpt,
		ModulePath:           *modulePathOpt,
		FileName:             filename,
		FlagSet:              fs,
	}

	OverRideHelp(fs)

	return optLength, result, nil
}

func OverRideHelp(fs *flag.FlagSet) *flag.FlagSet {
	fs.Usage = func() {
		fmt.Print(helpMessage)
	}
	return fs
}
