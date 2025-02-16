package main

import (
	/*fmtdayo*/
	"fmt"
	"log"

	//hogehogeho
	"github.com/magicdrive/goreg/internal/process"

	/*
		osssssssssssssssssss
		osssssssssssssssssss
		osssssssssssssssssss
		osssssssssssssssssss
		osssssssssssssssssss
	*/
	"os"
)

func main2() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: goreg <file.go>")
		os.Exit(1)
	}
	filename := os.Args[1]

	if err := process.ProcessFile(filename); err != nil {
		log.Fatal(err)
	}
}
