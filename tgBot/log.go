package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func getLogger(logFileName string) (*zap.SugaredLogger, error) {
	encoder := getEncoder()
	writer, err := getLogWriter(logFileName)
	if err != nil {
		return nil, err
	}
	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)
	logger := zap.New(core, zap.AddCaller())
	defer logger.Sync()
	return logger.Sugar(), nil
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(logFileName string) (zapcore.WriteSyncer, error) {
	f, err := os.Create(logFileName)
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(f), nil
}
