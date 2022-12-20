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
	Name  string
	Type  string
	IsPtr bool
}

type method struct {
	Name    string
	Takes   methodParam
	Returns methodParam
}

type handlerSet struct {
	GenFile         string
	TestFile        string
	PkgName         string
	PkgPath         string
	SpecPkgName     string
	SpecPkgPath     string
	SpecModule      string
	Imports         map[string]bool
	Receiver        string
	Methods         []method
	CustomMarshal   bool
	CustomUnmarshal bool
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

// moduleNameAndType splits a fully qualified type name into two parts: the import
// path (what you actually import to use the package), and the type name
// (what you use to declare a variable of that type).
//
// For example, a pointer to an http Request (*net/http.Request) becomes the
// import path "net/http", and the type "*http.Request".
func moduleNameAndType(s string) (importPath string, typeName string) {
	isPtr := s[0] == '*'
	// import path is "" if no ".", or s up to last "."
	if !strings.Contains(s, ".") {
		return "", s
	}
	importPath = s[:strings.LastIndex(s, ".")]
	if isPtr {
		importPath = importPath[1:]
	}
	// typeName is s if no "."
	// else, everything after last "/" with "*" prepended (if pointer)
	if !strings.Contains(s, "/") {
		return importPath, s
	}
	typeName = s[strings.LastIndex(s, "/")+1:]
	if isPtr {
		typeName = "*" + typeName
	}
	return importPath, typeName
}

func main() {
	genPkg := flag.String("genpkg", "server", "the package name to generate")
	receiver := flag.String("receiver", "Service", "the handler receiver type to use")
	flag.Parse()

	err := os.MkdirAll(*genPkg, os.ModeDir|0755)
	if err != nil {
		log.Fatal(err)
	}

	h := &handlerSet{
		PkgName:  *genPkg,
		Receiver: *receiver,
	}

	h.GenFile = fmt.Sprintf("%s%chandler.go", *genPkg, os.PathSeparator)

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
	h.Imports = make(map[string]bool, len(pkg.Imports))

	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				if !funcDecl.Name.IsExported() {
					return false
				}
				if funcDecl.Recv.NumFields() < 1 {
					return false
				}

				belongsToReceiver := false
				for _, rec := range funcDecl.Recv.List {
					starExp, ok := rec.Type.(*ast.StarExpr)
					if ok {
						if starExp.X.(*ast.Ident).Name == h.Receiver {
							belongsToReceiver = true
							break
						}
					}
				}
				if !belongsToReceiver {
					return false
				}

				// Exported func declaration - turn it into a handler
				meth := method{}
				meth.Name = funcDecl.Name.Name

				// If the receiver provides marshal or unmarshal methods, use them for encoding
				if meth.Name == "Marshal" {
					h.CustomMarshal = true
					return false
				}
				if meth.Name == "Unmarshal" {
					h.CustomUnmarshal = true
					return false
				}

				for _, field := range funcDecl.Type.Params.List {
					if len(field.Names) != 1 {
						panic(fmt.Sprintf("length of field names was %d!", len(field.Names)))
					}
					fieldName := field.Names[0].Name
					fieldType := pkg.TypesInfo.TypeOf(field.Type).String()

					importPath, typeName := moduleNameAndType(fieldType)
					if importPath != "" {
						h.Imports[importPath] = true
					}

					if typeName != "context.Context" {
						meth.Takes = methodParam{
							Name:  fieldName,
							Type:  strings.TrimPrefix(typeName, "*"),
							IsPtr: typeName[0] == '*',
						}
					}
				}

				h.Methods = append(h.Methods, meth)

				return false
			}
			return true
		})
	}
}
