//go:build wireinject
// +build wireinject

//
//go:generate go run github.com/google/wire/cmd/wire
package wiring

import (
	"github.com/google/wire"
	"github.com/nhtuan0700/GoLoad/internal/app"
	"github.com/nhtuan0700/GoLoad/internal/configs"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess"
	"github.com/nhtuan0700/GoLoad/internal/handler"
	"github.com/nhtuan0700/GoLoad/internal/logic"
	"github.com/nhtuan0700/GoLoad/internal/utils"
)

var WireSet = wire.NewSet(
	app.WireSet,
	configs.WireSet,
	dataaccess.WireSet,
	handler.WireSet,
	logic.WireSet,
	utils.WireSet,
)

func InitializeStandaloneServer(configPath configs.ConfigFilePath) (*app.StandaloneServer, func(), error) {
	wire.Build(WireSet)

	return nil, nil, nil
}

