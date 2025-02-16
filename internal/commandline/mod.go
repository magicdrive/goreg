package commandline

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
)

//go:embed help.txt
var helpMessage string

func OptParse(args []string) (int, *Option, error) {

	optLength := len(args)

	fs := flag.NewFlagSet("goreg", flag.ExitOnError)

	// --help
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

	_args := fs.Args()
	if len(args) == 0 {
		fmt.Println("Error: a file name is required")
		os.Exit(1)
	}

	filename := _args[0]

	result := &Option{
		WriteFlag:   *writeFlagOpt,
		HelpFlag:    *helpFlagOpt,
		VersionFlag: *versionFlagOpt,
		FileName:    filename,
		FlagSet:     fs,
	}
	OverRideHelp(fs, false)

	return optLength, result, nil
}

func OverRideHelp(fs *flag.FlagSet, noPagerFlag bool) *flag.FlagSet {
	fs.Usage = func() {
		fmt.Print(helpMessage)
	}
	return fs
}
