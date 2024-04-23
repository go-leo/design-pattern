package command

import (
	"context"
	"github.com/go-leo/design-pattern/cqrs"
)

type DemoCommandAssembler[Req any] interface {
    FromDemoCommandRequest(ctx context.Context, req Req) (*DemoCommandCmd, error)
}

type DemoCommandCmd struct {
}

type DemoCommand cqrs.CommandHandler[*DemoCommandCmd]

func NewDemoCommand() DemoCommand {
	return &demoCommand{}
}

type demoCommand struct {
}

func (h *demoCommand) Handle(ctx context.Context, cmd *DemoCommandCmd) error {
	//TODO implement me
	panic("implement me")
}
