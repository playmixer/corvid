package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/playmixer/secret-keeper/pkg/tools"
)

type loggerConfigurator struct {
	level      string
	logPath    string
	isTerminal bool
	isFile     bool
}

type option func(*loggerConfigurator)

func SetLevel(level string) option {
	return func(l *loggerConfigurator) {
		l.level = level
	}
}

func SetLogPath(path string) option {
	return func(l *loggerConfigurator) {
		if path != "" {
			l.logPath = path
		}
	}
}

func SetEnableFileOutput(t bool) option {
	return func(lc *loggerConfigurator) {
		lc.isFile = t
	}
}

func SetEnableTerminalOutput(t bool) option {
	return func(lc *loggerConfigurator) {
		lc.isTerminal = t
	}
}

func New(options ...option) (*zap.Logger, error) {
	cfg := loggerConfigurator{
		level:      "info",
		logPath:    "./logs/log.log",
		isTerminal: true,
		isFile:     true,
	}

	for _, opt := range options {
		opt(&cfg)
	}

	if cfg.logPath != "" {
		err := os.MkdirAll(filepath.Dir(cfg.logPath), tools.Mode0750)
		if err != nil {
			log.Println("failed create directory for logs")
		}
	}

	stdout := zapcore.AddSync(os.Stdout)

	f, err := os.OpenFile(cfg.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, tools.Mode0600)
	if err != nil {
		return nil, fmt.Errorf("failed create log file: %w", err)
	}
	file := zapcore.AddSync(f)

	level, err := zap.ParseAtomicLevel(cfg.level)
	if err != nil {
		return nil, fmt.Errorf("failed parse level: %w", err)
	}

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	ouputs := []zapcore.Core{}
	if cfg.isFile {
		ouputs = append(ouputs, zapcore.NewCore(fileEncoder, file, level))
	}
	if cfg.isTerminal {
		ouputs = append(ouputs, zapcore.NewCore(consoleEncoder, stdout, level))
	}

	core := zapcore.NewTee(ouputs...)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}
