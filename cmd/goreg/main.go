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

	if err := core.Apply(opt.FileName, core.GetModulePath(), opt.WriteFlag); err != nil {
		log.Fatal(err)
	}
}
