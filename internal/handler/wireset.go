package handler

import (
	"github.com/google/wire"
	"github.com/nhtuan0700/GoLoad/internal/handler/grpc"
)

var WireSet = wire.NewSet(
	grpc.WireSet,
)
