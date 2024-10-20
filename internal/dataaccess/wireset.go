package dataaccess

import (
	"github.com/google/wire"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/cache"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/database"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/file"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/mq"
)

var WireSet = wire.NewSet(
	database.WireSet,
	cache.WireSet,
	mq.WireSet,
	file.WireSet,
)
