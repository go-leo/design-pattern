package endpoint

import "github.com/go-leo/design-pattern/decorator"

type Decorator[Req any, Resp any] decorator.Decorator[Endpoint[Req, Resp]]

type DecoratorFunc[Req any, Resp any] decorator.Func[Endpoint[Req, Resp]]

func Chain[Req any, Resp any](ep Endpoint[Req, Resp], middlewares ...Decorator[Req, Resp]) Endpoint[Req, Resp] {
	decorators := make([]decorator.Decorator[Endpoint[Req, Resp]], 0, len(middlewares))
	for _, middleware := range middlewares {
		decorators = append(decorators, middleware)
	}
	return decorator.Chain[Endpoint[Req, Resp]](ep, decorators...)
}
