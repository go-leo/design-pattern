package gostruct

import "context"

//go:generate cqrs -service=Demo

// Demo
// @CQRS @QueryPath(../../query) @CommandPath(../../command)
type Demo interface {
	// DemoCommand
	// @CQRS @Command
	DemoCommand(ctx context.Context, req *DemoCommandReq) (*DemoCommandResp, error)

	// DemoQuery
	// @CQRS @Query
	DemoQuery(ctx context.Context, req *DemoQueryReq) (*DemoQueryResp, error)

	DemoDefault(ctx context.Context, req *DemoDefaultReq) (*DemoDefaultResp, error)
}

type DemoCommandReq struct {
}

type DemoCommandResp struct {
}

type DemoQueryReq struct {
}

type DemoQueryResp struct {
}

type DemoDefaultReq struct {
}

type DemoDefaultResp struct {
}
