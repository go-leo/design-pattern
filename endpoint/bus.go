package endpoint

import "context"

type BusApi interface {
	Subscribe(endpointName string, ep AnyEndpoint)
	Dispatch(ctx context.Context, req any) (resp any, err error)
}
