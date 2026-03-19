package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/magicdrive/goreg/internal/commandline"
	"github.com/magicdrive/goreg/internal/core"
	"github.com/magicdrive/goreg/internal/initcmd"
)

func Execute(version string) {
	args := os.Args[1:]

	// Check for init subcommand
	if len(args) > 0 && args[0] == "init" {
		InitCommand()
		return
	}

	_, opt, err := commandline.OptParse(args)
	if err != nil {
		log.Fatalf("Faital Error: %v\n", err)
	}

	if opt.VersionFlag {
		fmt.Printf("goreg version %s\n", version)
		os.Exit(0)
	}

	if opt.HelpFlag {
		opt.FlagSet.Usage()
		os.Exit(0)
	}

	if opt.FileName == "" {
		fmt.Println("Error: a file name is required")
		os.Exit(1)
	}

	if opt.ModulePath == "" {
		if _modulePath, err := core.GetModulePath(); err != nil {
			fmt.Println("Error: local modulepath not found. specify your local modulepath with --local option")
			os.Exit(1)
		} else {
			opt.ModulePath = _modulePath
		}
	}

	if err := core.Apply(opt); err != nil {
		log.Fatal(err)
	}
}

func InitCommand() {
	if err := initcmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
