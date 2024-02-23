package event

import (
	"github.com/go-leo/gox/syncx/gopher"
	"github.com/go-leo/gox/syncx/gopher/sample"
	"sync"
	"sync/atomic"
)

type option struct {
	Pool       gopher.Gopher
	MaxBackoff int
}

func newOption(opts ...Option) *option {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}
	if o.Pool == nil {
		o.Pool = sample.Gopher{}
	}
	if o.MaxBackoff == 0 {
		o.MaxBackoff = 16
	}
	return o
}

type Option func(*option)

func Pool(pool gopher.Gopher) Option {
	return func(o *option) {
		o.Pool = pool
	}
}

func MaxBackoff(backoff int) Option {
	return func(o *option) {
		o.MaxBackoff = backoff
	}
}

func NewBus(opts ...Option) Bus {
	return &bus{
		listenerMap:     sync.Map{},
		onceListenerMap: sync.Map{},
		wg:              sync.WaitGroup{},
		inShutdown:      atomic.Bool{},
		options:         newOption(opts...),
	}
}
