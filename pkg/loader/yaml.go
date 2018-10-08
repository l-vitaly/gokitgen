package loader

import (
	"io/ioutil"
	"path"

	"github.com/l-vitaly/gokitgen/pkg/config"
	yaml "gopkg.in/yaml.v2"
)

type yamlLoader struct {
	c *config.Config
}

func (l *yamlLoader) Load(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, l.c)
	if err != nil {
		return err
	}
	return nil
}

func (l *yamlLoader) SetConfig(c *config.Config) {
	l.c = c
}

func (l *yamlLoader) Supports(filename string) bool {
	ext := path.Ext(filename)
	if ext == ".yaml" || ext == ".yml" {
		return true
	}
	return false
}

func NewYAML() Loader {
	return &yamlLoader{}
}
