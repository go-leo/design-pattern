package main

import "fmt"

// The ambassador pattern creates a helper service that sends network requests on behalf of a
// client. It is often used in cloud-based applications to offload features of a remote service.
// An ambassador service can be thought of as an out-of-process proxy that is co-located with
// the client. Similar to the proxy design pattern, the ambassador service provides an interface for
// another remote service. In addition to the interface, the ambassador provides extra functionality
// and features, specifically offloaded common connectivity tasks. This usually consists of
// monitoring, logging, routing, security etc. This is extremely useful in legacy applications where
// the codebase is difficult to modify and allows for improvements in the application's networking
// capabilities.
// In this example, we will the ServiceAmbassador class represents the ambassador while
// the RemoteService represents a remote application.
func main() {
	ambassador := ServiceAmbassador{
		RemoteService: RemoteServiceImpl{},
	}
	fmt.Println(ambassador.DoRemoteFunction(10))
	fmt.Println(ambassador.DoRemoteFunction(-100))
}
