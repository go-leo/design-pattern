package cqrs

import (
	"github.com/go-leo/gox/syncx/gopher"
	"github.com/go-leo/gox/syncx/gopher/sample"
	"sync"
	"sync/atomic"
)

type option struct {
	Pool gopher.Gopher
}

func newOption(opts ...Option) *option {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}
	if o.Pool == nil {
		o.Pool = sample.Gopher{}
	}
	return o
}

type Option func(*option)

func Pool(pool gopher.Gopher) Option {
	return func(o *option) {
		o.Pool = pool
	}
}

func NewBus(opts ...Option) Bus {
	return &bus{
		commands:   sync.Map{},
		queries:    sync.Map{},
		wg:         sync.WaitGroup{},
		inShutdown: atomic.Bool{},
		options:    newOption(opts...),
	}
}
