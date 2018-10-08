package helloservice

type Message struct {
	Value string
}

type Service interface {
	Say(name string) (message Message, err error)
	WithoutParams() (err error)
	WithoutAll()
}
