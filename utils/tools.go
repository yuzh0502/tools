package tools

import (
	"os"
	"path/filepath"
	"strings"
)

// GetExePath 获取程序所在路径
func GetExePath() (string, error) {
	exePath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "", err
	}
	exePath = exePath[:strings.LastIndex(exePath, string(os.PathSeparator))]
	return exePath, nil
}
