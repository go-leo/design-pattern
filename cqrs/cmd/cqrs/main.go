package main

import (
	"flag"
	"fmt"
	"github.com/go-leo/design-pattern/cqrs/cmd/internal"
	"github.com/go-leo/gox/slicex"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
	"path/filepath"
)

var (
	serviceName = flag.String("service", "", "service interface Name. (required)")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of gors:\n")
	fmt.Fprintf(os.Stderr, "\tcqrs -service Demo\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("cqrs: ")
}

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Usage = Usage
	flag.Parse()
	if *showVersion {
		fmt.Printf("cqrs %v\n", internal.Version)
		return
	}

	// must set service names
	if len(*serviceName) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	// load package information
	pkg := loadPkg(args)

	// inspect package
	serviceFile, serviceDecl, serviceSpec, serviceType, serviceMethods := inspect(pkg)
	if serviceFile == nil || serviceDecl == nil || serviceSpec == nil || serviceType == nil {
		log.Fatal("error: not found service")
	}

	var files []*internal.File
	if serviceDecl != nil && serviceSpec != nil && serviceType != nil && len(serviceMethods) > 0 {
		if serviceDecl == nil || serviceDecl.Doc == nil {
			log.Println("not found", serviceSpec.Name.String(), "annotation:", `"@CQRS @QueryPath() @CommandPath()"`)
			os.Exit(2)
		}
		var comments []string
		for _, comment := range serviceDecl.Doc.List {
			comments = append(comments, comment.Text)
		}
		path := internal.NewPath(comments)
		if len(path.Command) == 0 || len(path.Query) == 0 {
			log.Println(`QueryPath or CommandPath is empty`)
			os.Exit(2)
		}

		queryAbs, err := filepath.Abs(path.Query)
		if err != nil {
			fmt.Printf("query path error: %s\n", err)
			os.Exit(2)
		}

		commandAbs, err := filepath.Abs(path.Command)
		if err != nil {
			fmt.Printf("command path error: %s\n", err)
			os.Exit(2)
		}

		for _, method := range serviceMethods {
			if slicex.IsEmpty(method.Names) {
				continue
			}
			methodName := method.Names[0]
			_, ok := method.Type.(*ast.FuncType)
			if !ok {
				log.Fatalf("error: func %s not convert to *ast.FuncType", methodName)
			}
			endpoint := methodName.String()

			if method.Doc == nil {
				files = append(files, internal.NewQueryFile(endpoint, queryAbs))
				continue
			}
			comments := slicex.Map[[]*ast.Comment, []string](
				method.Doc.List,
				func(i int, e1 *ast.Comment) string { return e1.Text },
			)
			file := internal.NewFileFromComment(endpoint, queryAbs, commandAbs, comments)
			if file == nil {
				continue
			}
			files = append(files, file)
		}
	}

	for _, f := range files {
		if err := f.Gen(); err != nil {
			log.Printf("%s.%s.%s error: %s\n", pkg.PkgPath, *serviceName, f.Endpoint, err)
			continue
		}
		log.Printf("%s.%s.%s wrote %s\n", pkg.PkgPath, *serviceName, f.Endpoint, f.AbsFilename)
	}
}

func loadPkg(args []string) *packages.Package {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps | packages.NeedExportFile | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypesSizes,
	}
	pkgs, err := packages.Load(cfg, args...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	pkg := pkgs[0]
	return pkg
}

func inspect(pkg *packages.Package) (*ast.File, *ast.GenDecl, *ast.TypeSpec, *ast.InterfaceType, []*ast.Field) {
	var serviceFile *ast.File
	var serviceDecl *ast.GenDecl
	var serviceSpec *ast.TypeSpec
	var serviceType *ast.InterfaceType
	var serviceMethods []*ast.Field
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			if node == nil {
				return true
			}
			denDecl, ok := node.(*ast.GenDecl)
			if !ok {
				return true
			}
			if denDecl.Tok != token.TYPE {
				// We only care about type declarations.
				return true
			}
			for _, spec := range denDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}
				if typeSpec.Name.Name != *serviceName {
					// This is not the interface type we're looking for.
					continue
				}
				serviceFile = file
				serviceDecl = denDecl
				serviceSpec = typeSpec
				serviceType = interfaceType
				serviceMethods = interfaceType.Methods.List
				return false
			}
			return true
		})
	}
	return serviceFile, serviceDecl, serviceSpec, serviceType, serviceMethods
}

type fileInfo struct {
	Type          string
	Name          string
	Template      string
	Endpoint      string
	LowerEndpoint string
}
