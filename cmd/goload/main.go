package main

import (
	"log"

	"github.com/nhtuan0700/GoLoad/internal/configs"
	"github.com/nhtuan0700/GoLoad/internal/wiring"
)

var (
	version    string
	commitHash string
)

const (
	flagConfigFilePath = "config-file-path"
)

func main() {
	app, cleanup, err := wiring.InitializeStandaloneServer(configs.ConfigFilePath("configs/local.yaml"))

	if err != nil {
		log.Panic(err)
	}
	defer cleanup()
	app.Start()
}
