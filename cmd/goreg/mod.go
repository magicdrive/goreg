package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/magicdrive/goreg/internal/commandline"
	"github.com/magicdrive/goreg/internal/core"
)

func Execute(version string) {
	_, opt, err := commandline.OptParse(os.Args[1:])
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
