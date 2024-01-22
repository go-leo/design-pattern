package example

import "context"

type Source struct {
}

type Target struct {
}

type SourceToTarget interface {
	Clone(ctx context.Context, source *Source) (*Target, error)
}

func init() {
	//arg := struct{ Name string }{Name: "jax"}
	//Call(arg)
}
