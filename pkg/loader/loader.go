package loader

import (
	"github.com/l-vitaly/gokitgen/pkg/config"
)

type Loader interface {
	Load(filename string) error
	SetConfig(c *config.Config)
	Supports(filename string) bool
}

type LoaderResolver struct {
	loaders []Loader
}

func (r *LoaderResolver) Add(l Loader) {
	r.loaders = append(r.loaders, l)
}

func (r *LoaderResolver) Resolve(filename string) Loader {
	for _, l := range r.loaders {
		if l.Supports(filename) {
			return l
		}
	}
	return nil
}

func NewLoaderResolver() *LoaderResolver {
	return &LoaderResolver{}
}
