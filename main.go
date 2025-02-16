package main

import (
	"runtime/debug"

	cmd "github.com/magicdrive/goreg/cmd/goreg"
)

var version string

func main() {
	cmd.Execute(Version())
}

func Version() string {
	if version != "" {
		return version
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		return buildInfo.Main.Version
	}
	return "unknown"

}
