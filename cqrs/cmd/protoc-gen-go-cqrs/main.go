package main

import (
	"github.com/go-leo/cqrs/cmd/protoc-gen-go-cqrs/gen"
)

func main() {
	gen.Gen("github.com/go-leo/design-pattern/cqrs")
}
