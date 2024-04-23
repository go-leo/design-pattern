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
		bus.RegisterQuery(query.NewDemoDefault()),
	)
	if err != nil {
		panic(err)
	}

	if err := bus.Exec(context.Background(), &command.DemoCommandCmd{}); err != nil {
		panic(err)
	}

	err, ok := <-bus.AsyncExec(context.Background(), &command.DemoCommandCmd{})
	if ok {
		panic(err)
	}

	q, err := bus.Query(context.Background(), &query.DemoQueryQuery{})
	if err != nil {
		panic(err)
	}
	log.Println(q)

	q, err = bus.Query(context.Background(), &query.DemoDefaultQuery{})
	if err == nil {
		panic("err should not nil")
	}
	log.Println(err)

	resC, errC := bus.AsyncQuery(context.Background(), &query.DemoQueryQuery{})
	if err, ok := <-errC; ok {
		panic(err)
	}
	log.Println(<-resC)

	genericBus := cqrs.NewGenericBus[*query.DemoQueryResult](bus)
	result, err := genericBus.Query(context.Background(), &query.DemoQueryQuery{})
	if err != nil {
		panic(err)
	}
	log.Println(result)

	resultC, errC := genericBus.AsyncQuery(context.Background(), &query.DemoQueryQuery{})
	if err, ok := <-errC; ok {
		panic(err)
	}
	log.Println(<-resultC)

}
