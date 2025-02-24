package commandline

import (
	"flag"
	"fmt"
	"os"

	_ "embed"

	"github.com/magicdrive/goreg/internal/model"
)

//go:embed help.txt
var helpMessage string

func OptParse(args []string) (int, *Option, error) {

	optLength := len(args)

	fs := flag.NewFlagSet("goreg", flag.ExitOnError)

	// --order
	orderOpt := fs.String("order", "", "Specify module group order.")
	fs.StringVar(orderOpt, "o", "", "Specify module group order.")

	// --organization
	organizationOpt := fs.String("organization", "", "Specify organization modulepath.")
	fs.StringVar(organizationOpt, "n", "", "Specify organization modulepath.")

	// --minimize-group
	minimizeGroupOpt := fs.Bool("minimize-group", false, "Not separate module group by alias.")
	fs.BoolVar(minimizeGroupOpt, "m", false, "Not separate module group by alias.")

	// --sort-include-aliases
	sortIncludeAliasOpt := fs.Bool("sort-include-alias", false, "foobar")
	fs.BoolVar(sortIncludeAliasOpt, "a", false, "foobar")

	// --local
	modulePathOpt := fs.String("local", "", "Specify local modulepath.")
	fs.StringVar(modulePathOpt, "l", "", "Specify local modulepath.")

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
