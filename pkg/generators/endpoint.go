package generators

import (
	"bytes"
	"fmt"
	"go/format"

	"github.com/l-vitaly/gokitgen/pkg/parser"
	"github.com/l-vitaly/gokitgen/pkg/utils"
)

const endpointStructName = "set"

type EndpointTransportDataField struct {
	Field parser.Field
	Name  string
}

type EndpointTransportData struct {
	Name   string
	Feilds []EndpointTransportDataField
}

type Endpoint struct {
	Name     string
	Method   parser.Method
	Request  EndpointTransportData
	Response EndpointTransportData
}

type Endpoints struct {
	Pkg         string
	ServiceName string
	List        []Endpoint
}

type EndpointGenerator struct {
	buf bytes.Buffer
}

func (g *EndpointGenerator) printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *EndpointGenerator) format() ([]byte, error) {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		fmt.Println(g.buf.String())
		return nil, err
	}
	return src, nil
}

func (g *EndpointGenerator) declareEndpoint(endpoints Endpoints) {
	g.printf("// Set collects all of the endpoints that compose an %s service.\n", endpoints.ServiceName)
	g.printf("type %s struct {\n", endpointStructName)
	for _, e := range endpoints.List {
		g.printf("\t%s endpoint.Endpoint\n", e.Name)
	}
	g.printf("}\n\n")
}

func (g *EndpointGenerator) declareRequestResponse(endpoints Endpoints) {
	for _, e := range endpoints.List {
		if len(e.Request.Feilds) > 0 {
			g.printf("type %s struct {\n", e.Request.Name)
			for _, f := range e.Request.Feilds {
				if f.Field.TypeWithPkg() == "context.Context" {
					continue
				}
				g.printf("\t%s %s\n", f.Name, f.Field.Type)
			}
			g.printf("}\n\n")
		}

		if len(e.Response.Feilds) > 0 {
			g.printf("type %s struct {\n", e.Response.Name)

			errField := ""
			for _, f := range e.Response.Feilds {
				g.printf("\t%s %s\n", f.Name, f.Field.Type)

				if errField == "" && f.Field.Type == "error" {
					errField = f.Name
				}

			}
			g.printf("}\n\n")

			if errField != "" {
				g.printf("func (r %s) Error() error { return r.%s }\n\n", e.Response.Name, errField)
			}
		}
	}
}

func (g *EndpointGenerator) declareMakeFuncs(endpoints Endpoints) {
	for _, e := range endpoints.List {
		g.printf("func make%s(s %s) endpoint.Endpoint {\n", e.Name, endpoints.ServiceName)
		g.printf("\treturn func(ctx context.Context, request interface{}) (interface{}, error) {\n")

		if len(e.Request.Feilds) > 0 {
			g.printf("\treq := request.(%s)\n", e.Request.Name)
		}

		if len(e.Response.Feilds) > 0 {
			g.printf("\t")
			for i, f := range e.Response.Feilds {
				if i > 0 {
					g.printf(",")
				}
				g.printf(f.Field.Name)
			}

			g.printf(" := ")
		}

		g.printf("s.%s", e.Method.Name)

		g.printf("(")

		for i, f := range e.Request.Feilds {
			if i > 0 {
				g.printf(",")
			}

			if f.Field.TypeWithPkg() == "context.Context" {
				g.printf("ctx")
			} else {
				g.printf("req.%s", f.Name)
			}

		}

		g.printf(")\n")

		if len(e.Response.Feilds) > 0 {
			g.printf("\treturn %s{\n", e.Response.Name)
			for _, f := range e.Response.Feilds {
				g.printf("\t%s: %s,\n", f.Name, f.Field.Name)
			}
			g.printf("}, nil")
		} else {
			g.printf("\treturn nil, nil")
		}

		g.printf("\t}\n")

		g.printf("}\n\n")
	}
}

func (g *EndpointGenerator) declareImports(endpoints Endpoints, mapImport parser.MapImport) {
	g.printf("import(\n")

	g.printf("\t\"github.com/go-kit/kit/endpoint\"\n")

	for _, e := range endpoints.List {
		for _, f := range e.Request.Feilds {
			if f.Field.Pkg != "" && f.Field.Pkg != endpoints.Pkg {
				if importName, ok := mapImport.Get(f.Field.Pkg); ok {
					g.printf("\t\"%s\"\n", importName)
				}
			}
		}
		for _, f := range e.Response.Feilds {
			if f.Field.Pkg != "" && f.Field.Pkg != endpoints.Pkg {
				if importName, ok := mapImport.Get(f.Field.Pkg); ok {
					g.printf("\t\"%s\"\n", importName)
				}
			}
		}
	}

	g.printf(")\n\n")
}

func (g *EndpointGenerator) declareMethods(result parser.Result) {
	for _, m := range result.Methods {
		g.printf("// %s implemented interface.\n", m.Name)
		g.printf("func (s %s) %s", endpointStructName, m.Name)
		g.printf("(")

		for i, p := range m.Params {
			if i > 0 {
				g.printf(",")
			}

			g.printf("%s %s", p.Name, p.TypeWithPkg())
		}
		g.printf(")")

		if len(m.Results) > 1 {
			g.printf("(")
		}

		for i, r := range m.Results {
			if i > 0 {
				g.printf(",")
			}
			g.printf("%s", r.Type)

		}

		if len(m.Results) > 1 {
			g.printf(")")
		}

		g.printf("{\n")

		g.printf("panic(\"endpoint not implemented %s\")", m.Name)

		g.printf("}\n\n")
	}
}

func (g *EndpointGenerator) declareFailer() {
	g.printf("\ntype errorer interface {\n\tError() error\n}\n\n")
}

func (g *EndpointGenerator) Generate(result parser.Result) ([]byte, error) {
	g.printf("package %s\n\n", result.Pkg)

	endpoints := Endpoints{
		Pkg:         result.Pkg,
		ServiceName: result.ServiceName,
	}
	for _, m := range result.Methods {
		lcName := utils.LcFirst(m.Name)

		var respFields []EndpointTransportDataField
		var reqFields []EndpointTransportDataField

		for _, f := range m.Params {
			reqFields = append(reqFields, EndpointTransportDataField{
				Name:  utils.UcFirst(f.Name),
				Field: f,
			})
		}

		for _, f := range m.Results {
			respFields = append(respFields, EndpointTransportDataField{
				Name:  utils.UcFirst(f.Name),
				Field: f,
			})
		}

		endpoints.List = append(endpoints.List, Endpoint{
			Name:   m.Name + "Endpoint",
			Method: m,
			Request: EndpointTransportData{
				Name:   lcName + "Request",
				Feilds: reqFields,
			},
			Response: EndpointTransportData{
				Name:   lcName + "Response",
				Feilds: respFields,
			},
		})
	}

	g.declareImports(endpoints, result.MapImport)
	g.declareFailer()
	g.declareEndpoint(endpoints)
	g.declareMethods(result)
	g.declareMakeFuncs(endpoints)
	g.declareRequestResponse(endpoints)

	return g.format()
}

func NewEndpoint() *EndpointGenerator {
	return &EndpointGenerator{}
}
