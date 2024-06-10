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
	configs.WireSet,
	dataaccess.WireSet,
	handler.WireSet,
	logic.WireSet,
	utils.WireSet,
	app.WireSet,
)

func InitializeStandaloneServer(configFilePath configs.ConfigFilePath) (*app.Server, func(), error) {
	wire.Build(WireSet)

	return nil, nil, nil
}
