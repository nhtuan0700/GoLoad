// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package wiring

import (
	"github.com/google/wire"
	"github.com/nhtuan0700/GoLoad/internal/app"
	"github.com/nhtuan0700/GoLoad/internal/configs"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/cache"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/database"
	"github.com/nhtuan0700/GoLoad/internal/handler"
	"github.com/nhtuan0700/GoLoad/internal/handler/grpc"
	"github.com/nhtuan0700/GoLoad/internal/handler/http"
	"github.com/nhtuan0700/GoLoad/internal/logic"
	"github.com/nhtuan0700/GoLoad/internal/utils"
)

// Injectors from wire.go:

func InitializeStandaloneServer(configFilePath configs.ConfigFilePath) (*app.Server, func(), error) {
	config, err := configs.NewConfig(configFilePath)
	if err != nil {
		return nil, nil, err
	}
	configsDatabase := config.Database
	log := config.Log
	logger, cleanup, err := utils.InitializeLogger(log)
	if err != nil {
		return nil, nil, err
	}
	db, cleanup2, err := database.InitializeAndMigrateUpDB(configsDatabase, logger)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	goquDatabase := database.InitializeGoquDB(db)
	accountDataAccessor := database.NewAccountDataAccessor(goquDatabase, logger)
	accountPasswordDataAccessor := database.NewAccountPasswordDataAccessor(goquDatabase, logger)
	configsCache := config.Cache
	client, err := cache.NewClient(configsCache, logger)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	takeAccountName := cache.NewTakenAccountName(client, logger)
	auth := config.Auth
	hash := logic.NewHash(auth)
	tokenPublicKeyDataAccessor := database.NewTokenPublicKeyAccessor(goquDatabase, logger)
	tokenPublicKeyCache := cache.NewTokenPublicKeyCache(client, logger)
	token, err := logic.NewToken(accountDataAccessor, tokenPublicKeyDataAccessor, tokenPublicKeyCache, auth, logger)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	account := logic.NewAccount(goquDatabase, accountDataAccessor, accountPasswordDataAccessor, takeAccountName, hash, token, logger)
	downloadTaskDataAccessor := database.NewDownloadTaskDataAccessor(goquDatabase, logger)
	downloadTask := logic.NewDownloadTask(goquDatabase, downloadTaskDataAccessor, accountDataAccessor, token, logger)
	goLoadServiceServer := grpc.NewHandler(account, downloadTask)
	server := grpc.NewServer(goLoadServiceServer, config, logger)
	configsGRPC := config.GRPC
	configsHTTP := config.HTTP
	httpServer := http.NewServer(configsGRPC, configsHTTP, auth, logger)
	appServer := app.NewServer(server, httpServer, logger)
	return appServer, func() {
		cleanup2()
		cleanup()
	}, nil
}

// wire.go:

var WireSet = wire.NewSet(configs.WireSet, dataaccess.WireSet, handler.WireSet, logic.WireSet, utils.WireSet, app.WireSet)
