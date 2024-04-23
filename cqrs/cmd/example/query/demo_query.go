package query

import (
	"context"
	"github.com/go-leo/design-pattern/cqrs"
)

type DemoQueryAssembler[Req any, Resp any] interface {
    FromDemoQueryReq(ctx context.Context, req Req) (*DemoQueryQuery, error)
    ToDemoQueryResp(ctx context.Context, res *DemoQueryResult) (Resp, error)
}

type DemoQueryQuery struct {
}

type DemoQueryResult struct {
}

type DemoQuery cqrs.QueryHandler[*DemoQueryQuery, *DemoQueryResult]

func NewDemoQuery() DemoQuery {
	return &demoQuery{}
}

type demoQuery struct {
}

func (h *demoQuery) Handle(ctx context.Context, q *DemoQueryQuery) (*DemoQueryResult, error) {
	//TODO implement me
	panic("implement me")
}
