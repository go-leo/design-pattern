package query

import (
	"context"
	"github.com/go-leo/design-pattern/cqrs"
)

type DemoDefaultAssembler[Req any, Resp any] interface {
	FromDemoDefaultReq(ctx context.Context, req Req) (*DemoDefaultQuery, error)
	ToDemoDefaultResp(ctx context.Context, res *DemoDefaultResult) (Resp, error)
}

type DemoDefaultQuery struct {
}

type DemoDefaultResult struct {
}

type DemoDefault cqrs.QueryHandler[*DemoDefaultQuery, *DemoDefaultResult]

func NewDemoDefault() DemoDefault {
	return &demoDefault{}
}

type demoDefault struct {
}

func (h *demoDefault) Handle(ctx context.Context, q *DemoDefaultQuery) (*DemoDefaultResult, error) {
	//TODO implement me
	panic("implement me")
}
