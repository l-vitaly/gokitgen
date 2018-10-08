package config

type TransportOptions interface {
	Unmarshal(v interface{}) error
}

type transportOptions struct {
	unmarshal func(interface{}) error
}

func (o *transportOptions) UnmarshalYAML(unmarshal func(interface{}) error) error {
	o.unmarshal = unmarshal
	return nil
}

func (o *transportOptions) Unmarshal(v interface{}) error {
	return o.unmarshal(v)
}

type Config struct {
	Service string
	Path    string
}
