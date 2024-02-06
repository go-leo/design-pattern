package main

import (
	"fmt"
	"time"
)

// ServiceAmbassador provides an interface to access RemoteService.
// The interface adds logging, latency testing and usage of the service in a safe way that will not
// add stress to the remote service when connectivity issues occur.
type ServiceAmbassador struct {
	RemoteService RemoteService
}

func (receiver ServiceAmbassador) DoRemoteFunction(value int) int64 {
	startTime := time.Now()
	result := receiver.RemoteService.DoRemoteFunction(value)
	endTime := time.Now()
	timeTaken := endTime.Sub(startTime)
	fmt.Println("Time taken (ms): ", timeTaken)
	return result
}
