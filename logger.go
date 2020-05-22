package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/ttys3/rotatefilehook"
)

var logger = logrus.New()

// LoggerInitError is a custom error representation
func LoggerInitError(err error) {
	logger.Error(err)
	os.Exit(3)
}

func initLogger(logLevel string) {
	logLevels := map[string]logrus.Level{
		"trace": logrus.TraceLevel,
		"debug": logrus.DebugLevel,
		"info":  logrus.InfoLevel,
		"warn":  logrus.WarnLevel,
		"error": logrus.ErrorLevel,
		"fatal": logrus.FatalLevel,
		"panic": logrus.PanicLevel,
	}

	// set default loglevel
	if logLevel == "" {
		logLevel = "info"
	}

	if LogLevel, ok := logLevels[logLevel]; !ok {
		LoggerInitError(fmt.Errorf("log level definition not found for '%s'", logLevel))
	} else {
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, ForceColors: true})
		logger.SetOutput(os.Stdout)
		logger.SetLevel(LogLevel)

		if logLevel == "trace" || logLevel == "debug" {
			logger.SetReportCaller(true)
		}

		rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
			Filename:   "s3-bucket-remover.log",
			MaxSize:    5, // the maximum size in megabytes
			MaxBackups: 7, // the maximum number of old log files to retain
			MaxAge:     7, // the maximum number of days to retain old log files
			LocalTime:  true,
			Level:      logrus.DebugLevel,
			Formatter:  &logrus.TextFormatter{FullTimestamp: true},
		})
		if err != nil {
			panic(err)
		}
		logger.AddHook(rotateFileHook)
	}
}
