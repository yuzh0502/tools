package main

import (
	"github.com/jlvihv/tools/tools"
	"os"
	"path/filepath"
)

var (
	configfile = ""
)

func main() {
	bot, err := getNewBot()
	if err != nil {
		os.Exit(1)
	}
}

func start() {

}

func readConfig() {
	if !filepath.IsAbs(configfile) {
		exePath, err := tools.GetExePath()
	}
}
func initLogger() {}
