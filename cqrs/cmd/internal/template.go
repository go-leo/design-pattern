package internal

import (
	_ "embed"
	"errors"
	"os"
	"text/template"
)

//go:embed command.go.template
var commandContent string

//go:embed query.go.template
var queryContent string

type File struct {
	Type          string
	AbsFilename   string
	Package       string
	Endpoint      string
	LowerEndpoint string
}

func (v File) Gen() error {
	if v.Type == "command" {
		return v.GenCommand()
	} else if v.Type == "query" {
		return v.GenQuery()
	}
	return errors.New("unknown endpoint type")
}

func (v File) GenQuery() error {
	tmpl, err := template.New("query").Parse(queryContent)
	if err != nil {
		return err
	}
	_, err = os.Stat(v.AbsFilename)
	if os.IsNotExist(err) {
		file, err := os.Create(v.AbsFilename)
		if err != nil {
			return err
		}
		return tmpl.Execute(file, &v)
	}
	if err != nil {
		return err
	}
	return nil
}

func (v File) GenCommand() error {
	tmpl, err := template.New("command").Parse(commandContent)
	if err != nil {
		return err
	}
	_, err = os.Stat(v.AbsFilename)
	if os.IsNotExist(err) {
		file, err := os.Create(v.AbsFilename)
		if err != nil {
			return err
		}
		return tmpl.Execute(file, &v)
	}
	if err != nil {
		return err
	}
	return nil
}
