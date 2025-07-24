package main

import (
	"os"
	"strings"

	akri "github.com/nubificus/go-akri/pkg/discovery-handler"
)

var AttestationServer string

func init() {
	AttestationServer = os.Getenv("DICE_AUTH_SERVICE_PORT")
	AttestationServer = strings.ReplaceAll(AttestationServer, "tcp", "http")
}

func main() {
	log_level := os.Getenv("LOG_LEVEL")
	var logLevel akri.LogLevel
	switch log_level {
	case "trace":
		logLevel = akri.TraceLevel
	case "debug":
		logLevel = akri.DebugLevel
	case "info":
		logLevel = akri.InfoLevel
	case "warn":
		logLevel = akri.WarnLevel
	case "error":
		logLevel = akri.ErrorLevel
	case "fatal":
		logLevel = akri.FatalLevel
	case "panic":
		logLevel = akri.PanicLevel
	default:
		logLevel = akri.InfoLevel
	}
	app := akri.NewApp(Discovery, akri.WithLogLevel(logLevel), akri.WithDiscoverSleep(10))
	app.Run()
}
