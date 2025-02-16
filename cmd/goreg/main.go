package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/magicdrive/goreg/internal/core"
)

func main() {
	flag.Parse()
	writeToFile := flag.Lookup("w") != nil

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: goreg [-w] <file.go>")
		os.Exit(1)
	}
	filename := flag.Arg(0)

	if err := core.Apply(filename, core.GetModulePath(), writeToFile); err != nil {
		log.Fatal(err)
	}
}
