package mq

import (
	"github.com/google/wire"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/mq/consumer"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/mq/producer"
)

var WireSet = wire.NewSet(
	producer.WireSet,
	consumer.WireSet,
)
