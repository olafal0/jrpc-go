package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

//go:embed handler.tpl
var configSrc string
var configTemplate = template.Must(template.New("handlers").Funcs(
	map[string]interface{}{
		"export":   export,
		"unexport": unexport,
	},
).Parse(configSrc))

type methodParam struct {
	Name string
	Type string
}

type method struct {
	Name         string
	ExportedName string
	TakesCtx     bool
	Params       []methodParam
}

type handlerSet struct {
	GenFile     string
	TestFile    string
	PkgName     string
	PkgPath     string
	SpecPkgName string
	SpecPkgPath string
	SpecModule  string
	Imports     []string
	Receiver    string
	Methods     []method
}

func export(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}

func unexport(s string) string {
	uppers := 0
	for i := 0; i < len(s); i++ {
		if strings.ToUpper(s[0:i]) != s[0:i] {
			break
		}
		uppers++
	}
	return strings.ToLower(s[0:uppers]) + s[uppers:]
}

func importPathToPkgName(s string) string {
	pathMatcher := regexp.MustCompile(`[\w.]+\/`)
	replaced := pathMatcher.ReplaceAll([]byte(s), []byte{})
	return string(replaced)
}

func main() {
	genPath := flag.String("genpath", "server", "the file path to generate files in")
	genPkg := flag.String("genpkg", "server", "the package name to generate")
	receiver := flag.String("receiver", "Service", "the handler receiver type to use")
	flag.Parse()

	err := os.MkdirAll(*genPath, os.ModeDir|0755)
	if err != nil {
		log.Fatal(err)
	}

	h := &handlerSet{
		PkgName:  *genPkg,
		Receiver: *receiver,
	}

	h.GenFile = fmt.Sprintf("%s%chandler.go", *genPath, os.PathSeparator)

	h.parsePackage()

	buf := new(bytes.Buffer)
	err = configTemplate.Execute(buf, h)
	if err != nil {
		log.Fatal(err)
	}
	saveFormatted(h.GenFile, buf)
}

func saveFormatted(filename string, content *bytes.Buffer) {
	formatted, err := format.Source(content.Bytes())
	if err != nil {
		fmt.Println(content.String())
		log.Fatal(err)
	}

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.Write(formatted)
	if err != nil {
		log.Fatal(err)
	}
}

// parsePackage returns a list of configuration items and a key prefix
func (h *handlerSet) parsePackage() {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedImports |
			packages.NeedModule,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, flag.Args()...)
	if err != nil {
		log.Fatal(err)
	}
	pkg := pkgs[0]

	h.SpecPkgName = pkg.Name
	h.SpecPkgPath = pkg.PkgPath
	h.SpecModule = pkg.Module.Path
	h.Methods = make([]method, 0)
	h.Imports = make([]string, 0, len(pkg.Imports))

	for importName := range pkg.Imports {
		if importName == "context" {
			continue
		}
		h.Imports = append(h.Imports, importName)
	}

	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				if !funcDecl.Name.IsExported() {
					return false
				}
				// Exported func declaration - turn it into a handler
				meth := method{}
				meth.ExportedName = funcDecl.Name.Name
				meth.Name = unexport(meth.ExportedName)
				meth.Params = make([]methodParam, 0)

				for _, field := range funcDecl.Type.Params.List {
					if len(field.Names) != 1 {
						panic(fmt.Sprintf("length of field names was %d!", len(field.Names)))
					}
					fieldName := field.Names[0].Name
					fieldType := pkg.TypesInfo.TypeOf(field.Type).String()
					// Strip module name from type descriptor
					// e.g. if the type is *sourcepkg.SourceType, the complete type will be
					// *module/sourcepkg.SourceType
					fieldType = importPathToPkgName(fieldType)
					if fieldType == "context.Context" {
						meth.TakesCtx = true
						continue
					}
					meth.Params = append(meth.Params, methodParam{
						Name: fieldName,
						Type: fieldType,
					})
				}

				h.Methods = append(h.Methods, meth)

				return false
			}
			return true
		})
	}
}
