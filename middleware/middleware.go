package middleware

import (
	"context"
)

type Invoker[Req any, Resp any] func(ctx context.Context, req Req) (Resp, error)

type Middleware[Req any, Resp any] func(ctx context.Context, req Req, invoker Invoker[Req, Resp]) (Resp, error)

func Chain[Req any, Resp any](middlewares ...Middleware[Req, Resp]) Middleware[Req, Resp] {
	var mdw Middleware[Req, Resp]
	if len(middlewares) == 0 {
		mdw = nil
	} else if len(middlewares) == 1 {
		mdw = middlewares[0]
	} else {
		mdw = func(ctx context.Context, req Req, invoker Invoker[Req, Resp]) (Resp, error) {
			return middlewares[0](ctx, req, getInvoker(middlewares, 0, invoker))
		}
	}
	return mdw
}

func getInvoker[Req any, Resp any](interceptors []Middleware[Req, Resp], curr int, finalInvoker Invoker[Req, Resp]) Invoker[Req, Resp] {
	if curr == len(interceptors)-1 {
		return finalInvoker
	}
	return func(ctx context.Context, req Req) (Resp, error) {
		return interceptors[curr+1](ctx, req, getInvoker(interceptors, curr+1, finalInvoker))
	}
}

func Invoke[Req any, Resp any](_ context.Context, req Req) (Resp, error) {
	var resp Resp
	return resp, nil
}
