package handler

import (
	"github.com/google/wire"
	"github.com/nhtuan0700/GoLoad/internal/handler/consumers"
	"github.com/nhtuan0700/GoLoad/internal/handler/grpc"
	"github.com/nhtuan0700/GoLoad/internal/handler/http"
)

var WireSet = wire.NewSet(
	grpc.WireSet,
	http.WireSet,
	consumers.WireSet,
)
