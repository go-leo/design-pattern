package builder

import "context"

type Builder[T any] interface {
	Build(ctx context.Context) (T, error)
}
