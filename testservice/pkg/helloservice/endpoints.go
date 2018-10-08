package helloservice

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type errorer interface {
	Error() error
}

// Set collects all of the endpoints that compose an Service service.
type set struct {
	SayEndpoint           endpoint.Endpoint
	WithoutParamsEndpoint endpoint.Endpoint
	WithoutAllEndpoint    endpoint.Endpoint
}

// Say implemented interface.
func (s set) Say(name string) (Message, error) {
	panic("endpoint not implemented Say")
}

// WithoutParams implemented interface.
func (s set) WithoutParams() error {
	panic("endpoint not implemented WithoutParams")
}

// WithoutAll implemented interface.
func (s set) WithoutAll() {
	panic("endpoint not implemented WithoutAll")
}

func makeSayEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(sayRequest)
		message, err := s.Say(req.Name)
		return sayResponse{
			Message: message,
			Err:     err,
		}, nil
	}
}

func makeWithoutParamsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		err := s.WithoutParams()
		return withoutParamsResponse{
			Err: err,
		}, nil
	}
}

func makeWithoutAllEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		s.WithoutAll()
		return nil, nil
	}
}

type sayRequest struct {
	Name string
}

type sayResponse struct {
	Message Message
	Err     error
}

func (r sayResponse) Error() error { return r.Err }

type withoutParamsResponse struct {
	Err error
}

func (r withoutParamsResponse) Error() error { return r.Err }
