package main

import (
	//"strings"

	"bufio"
	"errors"
	"fmt"
	"github.com/go-leo/design-pattern/cqrs/cmd/internal"
	"google.golang.org/protobuf/compiler/protogen"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func generateFile(gen *protogen.Plugin, file *protogen.File) {
	if len(file.Services) == 0 {
		return
	}
	generateFileContent(gen, file)
}

func generateFileContent(gen *protogen.Plugin, file *protogen.File) {
	if len(file.Services) == 0 {
		return
	}
	for _, service := range file.Services {
		files, err := getFileInfo(gen, file, service)
		if err != nil {
			gen.Error(fmt.Errorf("error: %w", err))
			return
		}
		for _, f := range files {
			if err := f.Gen(); err != nil {
				log.Printf("%s.%s error: %s\n", service.Desc.FullName(), f.Endpoint, err)
				continue
			}
			log.Printf("%s.%s wrote %s\n", service.Desc.FullName(), f.Endpoint, f.AbsFilename)
		}
	}
}

func getFileInfo(gen *protogen.Plugin, file *protogen.File, service *protogen.Service) ([]*internal.File, error) {
	path := internal.NewPath(splitComment(service.Comments.Leading.String()))
	if len(path.Command) == 0 || len(path.Query) == 0 {
		return nil, errors.New(`QueryPath or CommandPath is empty`)
	}
	cwd, _ := os.Getwd()
	queryAbs := filepath.Join(filepath.Dir(filepath.Join(cwd, file.Desc.Path())), path.Query)
	commandAbs := filepath.Join(filepath.Dir(filepath.Join(cwd, file.Desc.Path())), path.Command)
	var files []*internal.File
	for _, method := range service.Methods {
		if !method.Desc.IsStreamingServer() && !method.Desc.IsStreamingClient() {
			// Unary RPC method
			endpoint := method.GoName
			file := internal.NewFileFromComment(endpoint, queryAbs, commandAbs, splitComment(method.Comments.Leading.String()))
			if file == nil {
				continue
			}
			files = append(files, file)
		} else {
			// Streaming RPC method
			continue
		}
	}
	return files, nil
}

func splitComment(leadingComment string) []string {
	var comments []string
	scanner := bufio.NewScanner(strings.NewReader(leadingComment))
	for scanner.Scan() {
		line := scanner.Text()
		comments = append(comments, line)
	}
	return comments
}

func clientName(service *protogen.Service) string {
	return service.GoName + "Client"
}

func serverName(service *protogen.Service) string {
	return service.GoName + "Server"
}

func fullMethodName(service *protogen.Service, method *protogen.Method) string {
	return fmt.Sprintf("/%s/%s", service.Desc.FullName(), method.Desc.Name())
}
