package generators

import (
	"bytes"
	"fmt"
	"go/format"

	"github.com/l-vitaly/gokitgen/pkg/parser"
)

// LoggingGeneratorOption ...
type LoggingGeneratorOption func(g *loggingGenerator)

// LoggingGeneratorEnableStackTrace enable stack trace loggin.
func LoggingGeneratorEnableStackTrace(stackTrace bool) LoggingGeneratorOption {
	return func(g *loggingGenerator) {
		g.stackTrace = stackTrace
	}
}

type loggingGenerator struct {
	buf        bytes.Buffer
	stackTrace bool
}

func (g *loggingGenerator) printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *loggingGenerator) format() ([]byte, error) {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		fmt.Println(g.buf.String())
		return nil, err
	}
	return src, nil
}

func (g *loggingGenerator) declareImports(imports []string) {
	g.printf("import(\n")
	for _, i := range imports {
		g.printf("%s\n", i)
	}
	g.printf(")\n")
}

func (g *loggingGenerator) declareStruct(result parser.Result) {
	g.printf("type logging%s struct{\n", result.ServiceName)
	g.printf("next %s\n", result.ServiceName)
	g.printf("logger log.Logger\n")
	g.printf("}\n\n")
}

func (g *loggingGenerator) declareMethods(result parser.Result) {
	for _, m := range result.Methods {
		g.printf("func (s *logging%s) %s(", result.ServiceName, m.Name)

		for i, p := range m.Params {
			if i > 0 {
				g.printf(",")
			}
			g.printf("%s %s", p.Name, p.TypeWithPkg())
		}

		g.printf(")")

		if len(m.Results) > 0 {
			g.printf("(")
		}

		for i, r := range m.Results {
			if i > 0 {
				g.printf(",")
			}
			g.printf("%s %s", r.Name, r.TypeWithPkg())
		}

		if len(m.Results) > 0 {
			g.printf(")")
		}

		g.printf("{\n")

		g.printf("defer func(begin time.Time) {\n")

		g.printf("s.logger.Log(\n")

		g.printf("\"method\",\"%s\",\n", m.Name)

		if g.stackTrace {
			errName := ""
			for _, p := range m.Results {
				if p.Type == "error" {
					errName = p.Name
					break
				}
			}
			if errName != "" {
				g.printf("\"stackTrace\",getStackTrace(%s),\n", errName)
			}
		}

		for _, p := range m.Params {
			g.printf("\"%s\",%s,\n", p.Name, p.Name)
		}

		g.printf(")\n")

		g.printf("}(time.Now())\n\n")

		if len(m.Results) > 0 {
			g.printf("return ")
		}

		g.printf("s.%s(", m.Name)

		for i, p := range m.Params {
			if i > 0 {
				g.printf(",")
			}
			g.printf(p.Name)
		}

		g.printf(")\n")

		g.printf("}\n\n")
	}
}

func (g *loggingGenerator) declareStactTraceFn() {
	g.printf("type stackTracer interface {\n")
	g.printf("StackTrace() errors.StackTrace\n")
	g.printf("}\n\n")
	g.printf("func getStackTrace(err error) string {\n")
	g.printf("if err, ok := err.(stackTracer); ok {\n")
	g.printf(`return fmt.Sprintf("%%+v\n", err.StackTrace())`)
	g.printf("\n")
	g.printf("}\n")
	g.printf("return \"\"\n")
	g.printf("}\n\n")
}

func (g *loggingGenerator) declareNewLogging(result parser.Result) {
	g.printf("// NewLogging%s creates a logging service middleware.\n", result.ServiceName)
	g.printf("func NewLogging%[1]s(next %[1]s, logger log.Logger) %[1]s {\n", result.ServiceName)
	g.printf("return &logging%s{next: next, logger: logger}\n", result.ServiceName)
	g.printf("}\n\n")
}

func (g *loggingGenerator) Generate(result parser.Result) ([]byte, error) {
	g.printf("package %s\n\n", result.Pkg)

	imports := []string{
		`"fmt"`,
		`"time"`,
		"\n",
		`"github.com/go-kit/kit/log"`,
	}
	if g.stackTrace {
		imports = append(imports, `"github.com/pkg/errors"`)
	}

	g.declareImports(imports)
	g.declareStruct(result)
	g.declareMethods(result)
	if g.stackTrace {
		g.declareStactTraceFn()
	}
	g.declareNewLogging(result)
	return g.format()
}

// NewLogging cerates a logginh generate.
func NewLogging(options ...LoggingGeneratorOption) Generator {
	g := &loggingGenerator{}
	for _, o := range options {
		o(g)
	}
	return g
}
