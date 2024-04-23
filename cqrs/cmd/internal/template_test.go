package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommand(t *testing.T) {
	v := &File{
		AbsFilename:   "/Users/songyancheng/Workspace/github/go-leo/cqrs/cmd/example/command/login.go",
		Package:       "command",
		Endpoint:      "Login",
		LowerEndpoint: "login",
	}
	assert.NoError(t, v.GenCommand())
}

func TestQuery(t *testing.T) {
	v := &File{
		AbsFilename:   "/Users/songyancheng/Workspace/github/go-leo/cqrs/cmd/example/query/find.go",
		Package:       "query",
		Endpoint:      "Find",
		LowerEndpoint: "find",
	}
	assert.NoError(t, v.GenQuery())
}
