package internal

import _ "embed"

//go:embed command.go.template
var CommandTmpl string

//go:embed query.go.template
var QuertTmpl string
