package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	tools "github.com/jlvihv/tools/utils"
	"os"
	"path/filepath"
)

type TgBot struct {
}

var (
	configFile  = ""
	tgBotConfig = TgBot{}
)

func start() {
	readConfig()
	initLogger()
}

func readConfig() {
	if !filepath.IsAbs(configFile) {
		exePath, err := tools.GetExePath()
		if err != nil {
			fmt.Printf("can not get exePath: %s\n", err)
			os.Exit(1)
		}
		configFile, err = filepath.Abs(exePath + "/" + configFile)
		if err != nil {
			fmt.Printf("can not get config file abs path: %s\n", err)
			os.Exit(1)
		}
	}
	_, err := toml.DecodeFile(configFile, &tgBotConfig)
	if err != nil {
		fmt.Printf("can not decode config file: %s\n", err)
		os.Exit(1)
	}
}

func initLogger() {}
