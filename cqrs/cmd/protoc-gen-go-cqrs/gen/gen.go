package gen

import (
	//"strings"

	"bufio"
	"flag"
	"fmt"
	design_pattern "github.com/go-leo/design-pattern"
	"github.com/go-leo/design-pattern/cqrs/cmd/internal"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
	"os"
	"path/filepath"
	"strings"
)

func Gen(mainPkg string) {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-gors %v\n", design_pattern.Version)
		return
	}

	var flags flag.FlagSet
	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(mainPkg, gen, f)
		}
		return nil
	})
}

func generateFile(mainPkg string, gen *protogen.Plugin, file *protogen.File) {
	if len(file.Services) == 0 {
		return
	}
	generateFileContent(mainPkg, gen, file)
}

func generateFileContent(mainPkg string, gen *protogen.Plugin, file *protogen.File) {
	if len(file.Services) == 0 {
		return
	}
	for _, service := range file.Services {
		files, err := getFileInfo(mainPkg, gen, file, service)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warn: %s\n", err.Error())
			return
		}
		for _, f := range files {
			if err := f.Gen(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s.%s error: %s \n", service.Desc.FullName(), f.Endpoint, err)
				continue
			}
		}
	}
}

func getFileInfo(mainPkg string, gen *protogen.Plugin, file *protogen.File, service *protogen.Service) ([]*internal.File, error) {
	path := internal.NewPath(splitComment(service.Comments.Leading.String()))
	if len(path.Command) == 0 || len(path.Query) == 0 {
		return nil, fmt.Errorf(`%s QueryPath or CommandPath is empty`, service.Desc.FullName())
	}
	cwd, _ := os.Getwd()
	queryAbs := filepath.Join(filepath.Dir(filepath.Join(cwd, file.Desc.Path())), path.Query)
	commandAbs := filepath.Join(filepath.Dir(filepath.Join(cwd, file.Desc.Path())), path.Command)
	var files []*internal.File
	for _, method := range service.Methods {
		if !method.Desc.IsStreamingServer() && !method.Desc.IsStreamingClient() {
			// Unary RPC method
			endpoint := method.GoName
			file := internal.NewFileFromComment(mainPkg, endpoint, queryAbs, commandAbs, splitComment(method.Comments.Leading.String()))
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

func fullMethodName(service *protogen.Service, method *protogen.Method) string {
	return fmt.Sprintf("/%s/%s", service.Desc.FullName(), method.Desc.Name())
}
