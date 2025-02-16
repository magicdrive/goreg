package commandline

import "flag"

type Option struct {
	WriteFlag   bool
	HelpFlag    bool
	VersionFlag bool
	FileName    string
	FlagSet     *flag.FlagSet
}
