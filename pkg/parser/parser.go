package parser

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

type MapImport map[string]string

func (m MapImport) Add(name string) {
	i := strings.LastIndex(name, "/")
	m[name[i+1:]] = name
}

func (m MapImport) Get(pkg string) (val string, ok bool) {
	val, ok = m[pkg]
	return
}

type Result struct {
	Pkg         string
	Root        string
	ServiceName string
	MapImport   MapImport
	Methods     []Method
}

type Method struct {
	Name    string
	Params  []Field
	Results []Field
}

type Field struct {
	Name string
	Type string
	Pkg  string
}

func (f Field) TypeWithPkg() string {
	if f.Pkg != "" {
		return f.Pkg + "." + f.Type
	}
	return f.Type
}

type Parser struct {
}

func (p *Parser) getIdentNameWithPrefix(idents []*ast.Ident, def string) string {
	if len(idents) > 0 {
		return idents[0].Name
	}
	return def
}

func (p *Parser) getParamInfo(expr ast.Expr) (pkg string, t string) {
	fmt.Println(expr)

	switch t := expr.(type) {
	case *ast.Ident:
		if t.Obj != nil {
			return "", t.Obj.Name
		}
		return "", t.Name
	case *ast.SelectorExpr:
		return t.X.(*ast.Ident).Name, t.Sel.Name
	case *ast.ArrayType:
		pkg, tp := p.getParamInfo(t.Elt)
		return pkg, "[]" + tp
	}

	panic("panic type")
}

func (p *Parser) extractFieldList(pkg string, fieldList *ast.FieldList, defPrefix string) []Field {
	var result []Field

	if fieldList != nil {
		for i, param := range fieldList.List {

			f := Field{}

			f.Name = p.getIdentNameWithPrefix(param.Names, defPrefix+strconv.Itoa(i+1))

			pkg, t := p.getParamInfo(param.Type)
			f.Type = t
			f.Pkg = pkg

			result = append(result, f)
		}
	}
	return result
}

func (p *Parser) getRoot(pkgDir string) (string, error) {
	goSrc := build.Default.GOPATH + "/src/"
	return pkgDir[len(goSrc):], nil
}

func (p *Parser) Parse(basePath, serviceIface string) (Result, error) {
	pkg, err := build.Default.ImportDir(basePath, 0)
	if err != nil {
		return Result{}, err
	}

	root, err := p.getRoot(pkg.Dir)
	if err != nil {
		return Result{}, err
	}
	result := Result{
		MapImport: MapImport(map[string]string{}),
		Root:      root,
	}
	result.MapImport.Add(root)

	for _, imp := range pkg.Imports {
		result.MapImport.Add(imp)
	}
	fs := token.NewFileSet()
	for _, name := range pkg.GoFiles {

		name = filepath.Join(basePath, name)

		if !strings.HasSuffix(name, ".go") {
			continue
		}
		parsedFile, err := parser.ParseFile(fs, name, nil, 0)
		if err != nil {
			return Result{}, nil
		}

		for _, d := range parsedFile.Decls {
			if g, ok := d.(*ast.GenDecl); ok {
				for _, s := range g.Specs {
					typeSpec, ok := s.(*ast.TypeSpec)
					if !ok {
						continue
					}
					name := pkg.Name + "." + typeSpec.Name.Name

					if name == serviceIface {
						result.Pkg = pkg.Name
						result.ServiceName = typeSpec.Name.Name

						ifaceType, ok := typeSpec.Type.(*ast.InterfaceType)
						if !ok {
							continue
						}

						for _, f := range ifaceType.Methods.List {
							funcType, ok := f.Type.(*ast.FuncType)
							if !ok {
								continue
							}

							params := p.extractFieldList(pkg.Name, funcType.Params, "param")
							results := p.extractFieldList(pkg.Name, funcType.Results, "result")

							result.Methods = append(result.Methods, Method{
								Name:    f.Names[0].Name,
								Params:  params,
								Results: results,
							})
						}
					}
				}
			}
		}
	}
	return result, nil
}
