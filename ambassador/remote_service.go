package main

import (
	"github.com/go-leo/gox/mathx/randx"
	"time"
)

// RemoteService shared by RemoteServiceImpl and ServiceAmbassador.
type RemoteService interface {
	DoRemoteFunction(value int) int64
}

// RemoteServiceImpl a remote legacy application represented by a Singleton implementation.
type RemoteServiceImpl struct{}

func (receiver RemoteServiceImpl) DoRemoteFunction(value int) int64 {
	result := randx.Int63n(1000)
	time.Sleep(time.Duration(result) * time.Millisecond)
	return int64(-value)
}
