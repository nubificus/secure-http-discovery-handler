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
	app := akri.NewApp(Discovery, akri.WithLogLevel(akri.TraceLevel))
	app.Run()
}
