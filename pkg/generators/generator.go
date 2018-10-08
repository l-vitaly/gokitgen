package generators

import "github.com/l-vitaly/gokitgen/pkg/parser"

type Generator interface {
	Generate(result parser.Result) ([]byte, error)
}
