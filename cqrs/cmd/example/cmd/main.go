package main

import (
	"context"
	"errors"
	"github.com/go-leo/design-pattern/cqrs"
	"github.com/go-leo/design-pattern/cqrs/cmd/example/command"
	"github.com/go-leo/design-pattern/cqrs/cmd/example/query"
	"log"
)

func main() {
	bus := cqrs.NewBus()
	err := errors.Join(
		bus.RegisterCommand(command.NewDemoCommand()),
		bus.RegisterQuery(query.NewDemoQuery()),
	)
	if err != nil {
		panic(err)
	}

	if err := bus.Exec(context.Background(), &command.DemoCommandCmd{}); err != nil {
		panic(err)
	}

	q, err := bus.Query(context.Background(), &query.DemoQueryQuery{})
	if err != nil {
		panic(err)
	}
	log.Println(q)

}
