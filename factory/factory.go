package factory

import "context"

type Factory[T any, P any] interface {
	Create(ctx context.Context, param P) (T, error)
}
