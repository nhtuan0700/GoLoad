package dataaccess

import (
	"github.com/google/wire"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/database"
)

var WireSet = wire.NewSet(
	database.WireSet,
)
