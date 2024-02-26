package cqrs

import (
	"github.com/go-leo/cqrs"
)

type CommandHandler[Command any] interface {
	cqrs.CommandHandler[Command]
}

type QueryHandler[Query any, Result any] interface {
	cqrs.QueryHandler[Query, Result]
}

type Bus = cqrs.Bus

type Option = cqrs.Option

var Pool = cqrs.Pool

var NewBus = cqrs.NewBus
