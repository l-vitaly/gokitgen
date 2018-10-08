package generators

import (
	"bytes"
	"fmt"
	"go/format"

	"github.com/l-vitaly/gokitgen/pkg/parser"
	"github.com/l-vitaly/gokitgen/pkg/utils"
)

// HTTPGeneratorOption http generator option.
type HTTPGeneratorOption func(g *httpGenerator)

// HTTPGeneratorZipkin zpikin.
func HTTPGeneratorZipkin(zipkin bool) HTTPGeneratorOption {
	return func(g *httpGenerator) {
		g.zipkin = zipkin
	}
}

// HTTPGeneratorLogger zpikin.
func HTTPGeneratorLogger(logger bool) HTTPGeneratorOption {
	return func(g *httpGenerator) {
		g.logger = logger
	}
}

// HTTPGeneratorClient client.
func HTTPGeneratorClient(client bool) HTTPGeneratorOption {
	return func(g *httpGenerator) {
		g.client = client
	}
}

// HTTPGeneratorGenericResponse generic responce.
func HTTPGeneratorGenericResponse(genericResponse bool) HTTPGeneratorOption {
	return func(g *httpGenerator) {
		g.genericResponse = genericResponse
	}
}

// HTTPGeneratorGenericRequest generic responce.
func HTTPGeneratorGenericRequest(genericRequest bool) HTTPGeneratorOption {
	return func(g *httpGenerator) {
		g.genericRequest = genericRequest
	}
}

type httpGenerator struct {
	buf             bytes.Buffer
	zipkin          bool
	client          bool
	genericResponse bool
	genericRequest  bool
	logger          bool
}

func (g *httpGenerator) printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *httpGenerator) format() ([]byte, error) {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		fmt.Println(g.buf.String())
		return nil, err
	}
	return src, nil
}

func (g *httpGenerator) declareImports(imports []string) {
	g.printf("import(\n")
	for _, i := range imports {
		g.printf(i)
		g.printf("\n")
	}
	g.printf(")\n\n")
}

func (g *httpGenerator) declareVars() {
	g.printf("// ErrBadRequest bad request.\n")
	g.printf("var ErrBadRequest = errors.New(\"bad request\")\n\n")
}

func (g *httpGenerator) declareNewServerHandler(result parser.Result) {
	g.printf("// NewHTTPHandler returns an HTTP handler.\n")
	g.printf("func NewHTTPHandler(svc %s", result.ServiceName)
	if g.zipkin {
		g.printf(", zipkinTracer *stdzipkin.Tracer")
	}
	if g.logger {
		g.printf(", logger log.Logger")
	}
	g.printf(") http.Handler {\n")

	if g.zipkin {
		g.printf("zipkinServer := zipkin.HTTPServerTrace(zipkinTracer)\n\n")
	}

	g.printf("opts := []kithttp.ServerOption{\n")
	g.printf("kithttp.ServerErrorEncoder(errorHTTPEncoder),\n")
	if g.logger {
		g.printf("kithttp.ServerErrorLogger(logger),\n")
	}
	if g.zipkin {
		g.printf("zipkinServer,\n")
	}

	g.printf("}\n\n")

	g.declareServerHandlers(result)

	g.printf("r := mux.NewRouter()\n\n")

	g.printf("return r\n")

	g.printf("}\n\n")
}

func (g *httpGenerator) declareNewClientHandler(result parser.Result) {
	if g.client {
		g.printf("// NewHTTPClient returns an %s backed by an HTTP server living at the remote instance.\n", result.ServiceName)
		g.printf("func NewHTTPClient(instance string")

		if g.zipkin {
			g.printf(", zipkinTracer *stdzipkin.Tracer")
		}
		if g.logger {
			g.printf(", logger log.Logger")
		}

		g.printf(") (%s, error) {\n", result.ServiceName)

		g.printf("if !strings.HasPrefix(instance, \"http\") {\n")
		g.printf("instance = \"http://\" + instance\n")
		g.printf("}\n")
		g.printf("u, err := url.Parse(instance)\n")
		g.printf("if err != nil {\n")
		g.printf("return nil, err\n")
		g.printf("}\n")

		if g.zipkin {
			g.printf("zipkinClient := zipkin.HTTPClientTrace(zipkinTracer)\n\n")
		}

		g.printf("opts := []kithttp.ClientOption{\n")
		if g.zipkin {
			g.printf("zipkinClient,\n")
		}
		g.printf("}\n\n")

		g.declareClientEndpoints(result)

		g.printf("}\n\n")
	}
}

func (g *httpGenerator) declareServerHandlers(result parser.Result) {
	for _, m := range result.Methods {
		g.printf("%sHandler := kithttp.NewServer(\n", utils.LcFirst(m.Name))
		g.printf("make%sEndpoint(svc),\n", m.Name)
		g.printf("decodeHTTP%sRequest,\n", m.Name)
		if g.genericResponse {
			g.printf("encodeHTTPGenericResponse,\n")
		} else {
			g.printf("encodeHTTP%sResponse,\n", m.Name)
		}
		g.printf("opts...,\n")
		g.printf(")\n\n")
	}
}

func (g *httpGenerator) declareClientEndpoints(result parser.Result) {
	for _, m := range result.Methods {
		g.printf("%sEndpoint := kithttp.NewClient(\n", utils.LcFirst(m.Name))
		g.printf("\"GET\",\n")
		g.printf("copyURL(u, \"\"),\n")
		if g.genericRequest {
			g.printf("encodeHTTPGenericRequest,\n")
		} else {
			g.printf("encodeHTTP%sRequest,\n", m.Name)
		}
		g.printf("decodeHTTP%sResponse,\n", m.Name)
		g.printf("opts...,\n")
		g.printf(").Endpoint()\n\n")
	}

	g.printf("return &%s{\n", endpointStructName)
	for _, m := range result.Methods {
		g.printf("%sEndpoint: %sEndpoint,\n", m.Name, utils.LcFirst(m.Name))
	}
	g.printf("}, nil\n")
}

func (g *httpGenerator) declareDecodeEncode(result parser.Result) {
	for _, m := range result.Methods {
		g.printf("func decodeHTTP%sRequest(ctx context.Context, r *http.Request) (interface{}, error) {\n", m.Name)
		g.printf(`panic("not implement decodeHTTP%sRequest")`, m.Name)
		g.printf("\n}\n\n")

		if !g.genericResponse {
			g.printf("func encodeHTTP%sResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {\n", m.Name)
			g.printf(`panic("not implement decodeHTTP%sRequest")`, m.Name)
			g.printf("\n}\n\n")
		}

		if g.client {
			if !g.genericRequest {
				g.printf("func encodeHTTP%sRequest(ctx context.Context, r *http.Request, request interface{})  error {", m.Name)
				g.printf(`panic("not implement encodeHTTP%sRequest")`, m.Name)
				g.printf("\n}\n\n")
			}
			g.printf("func decodeHTTP%sResponse(ctx context.Context, r *http.Response) (interface{}, error) {", m.Name)
			g.printf(`panic("not implement encodeHTTP%sRequest")`, m.Name)
			g.printf("\n}\n\n")
		}

	}

	if g.genericResponse {
		g.printf("func encodeHTTPGenericResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {\n")
		g.printf("if f, ok := response.(errorer); ok && f.Error() != nil {\n")
		g.printf("errorHTTPEncoder(ctx, f.Error(), w)\n")
		g.printf("return nil\n")
		g.printf("}\n")
		g.printf("w.Header().Set(\"Content-Type\", \"application/json; charset=utf-8\")\n")
		g.printf("return json.NewEncoder(w).Encode(response)\n")
		g.printf("}\n\n")
	}

	if g.genericRequest {
		g.printf("func encodeHTTPGenericRequest(ctx context.Context, r *http.Request, request interface{}) error {\n")
		g.printf("var buf bytes.Buffer\n")
		g.printf("if err := json.NewEncoder(&buf).Encode(request); err != nil {\n")
		g.printf("return err\n")
		g.printf("}\n")
		g.printf("r.Body = ioutil.NopCloser(&buf)\n")
		g.printf("return nil\n")
		g.printf("}\n\n")
	}
}

func (g *httpGenerator) declareEncodeError(result parser.Result) {
	g.printf("func errorHTTPEncoder(ctx context.Context, err error, w http.ResponseWriter) {\n")

	g.printf("w.Header().Set(\"Content-Type\", \"application/json; charset=utf-8\")\n")

	g.printf("switch err {\n")
	g.printf("case ErrBadRequest:\nw.WriteHeader(http.StatusBadRequest)\n")
	g.printf("default:\nw.WriteHeader(http.StatusInternalServerError)\n")
	g.printf("}\n\n")

	g.printf("json.NewEncoder(w).Encode(map[string]interface{}{\n")
	g.printf("\"error\": err.Error(),\n")
	g.printf("})\n")

	g.printf("}\n\n")
}

func (g *httpGenerator) declareCopyURL() {
	g.printf("func copyURL(base *url.URL, path string) *url.URL {\n")
	g.printf("next := *base\n")
	g.printf("next.Path = path\n")
	g.printf("return &next\n")
	g.printf("}\n\n")
}

func (g *httpGenerator) Generate(result parser.Result) ([]byte, error) {
	g.printf("package %s\n\n", result.Pkg)

	imports := []string{
		`"context"`,
		`"encoding/json"`,
		`"errors"`,
		`"net/http"`,
		`"net/url"`,
	}

	if g.client {
		imports = append(imports, `"strings"`)
	}

	imports = append(imports, []string{
		"",
		`kithttp "github.com/go-kit/kit/transport/http"`,
		`"github.com/gorilla/mux"`,
	}...)

	if g.zipkin {
		imports = append(
			imports,
			[]string{
				`"github.com/go-kit/kit/tracing/zipkin"`,
				`stdzipkin "github.com/openzipkin/zipkin-go"`,
			}...,
		)
	}

	if g.logger {
		imports = append(imports, `"github.com/go-kit/kit/log"`)
	}

	g.declareImports(imports)
	g.declareVars()
	g.declareNewServerHandler(result)
	g.declareNewClientHandler(result)
	g.declareDecodeEncode(result)
	g.declareEncodeError(result)
	g.declareCopyURL()

	return g.format()
}

// NewHTTPTransport creates a http transport generator.
func NewHTTPTransport(options ...HTTPGeneratorOption) Generator {
	g := &httpGenerator{}
	for _, o := range options {
		o(g)
	}
	return g
}
