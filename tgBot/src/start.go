package src

import (
	"fmt"
	"github.com/BurntSushi/toml"
	tools "github.com/jlvihv/tools/utils"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

type TgBot struct {
	Log         Log
	TgBotConfig TgBotConfig
}

type Log struct {
	LogFileName string
}

type TgBotConfig struct {
	AdminID string
	Proxy   string
	Token   string
}

var (
	configFile  = "../tgBot.toml"
	tgBotConfig = TgBot{}
	logger      *zap.SugaredLogger
)

func Start() {
	readConfig()
	initLogger()
}

func readConfig() {
	fmt.Println("read config...")
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

func initLogger() {
	fmt.Println("init logger...")
	var err error
	logger, err = getLogger(tgBotConfig.Log.LogFileName)
	if err != nil {
		fmt.Printf("get logger failed: %s\n", err)
		os.Exit(2)
	}
	logger.Info("logger init success")
}
