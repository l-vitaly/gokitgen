package helloservice

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// ErrBadRequest bad request.
var ErrBadRequest = errors.New("bad request")

func NewHTTPHandler(svc Service) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(errorEncoder),
	}

	sayHandler := kithttp.NewServer(
		makeSayEndpoint(svc),
		decodeHTTPSayRequest,
		encodeHTTPSayResponse,
		opts...,
	)

	withoutParamsHandler := kithttp.NewServer(
		makeWithoutParamsEndpoint(svc),
		decodeHTTPWithoutParamsRequest,
		encodeHTTPWithoutParamsResponse,
		opts...,
	)

	withoutAllHandler := kithttp.NewServer(
		makeWithoutAllEndpoint(svc),
		decodeHTTPWithoutAllRequest,
		encodeHTTPWithoutAllResponse,
		opts...,
	)

	r := mux.NewRouter()

	return r
}

func decodeHTTPSayRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	panic("not implement decodeHTTPSayRequest")
}

func encodeHTTPSayResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	panic("not implement decodeHTTPSayRequest")
}

func decodeHTTPWithoutParamsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	panic("not implement decodeHTTPWithoutParamsRequest")
}

func encodeHTTPWithoutParamsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	panic("not implement decodeHTTPWithoutParamsRequest")
}

func decodeHTTPWithoutAllRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	panic("not implement decodeHTTPWithoutAllRequest")
}

func encodeHTTPWithoutAllResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	panic("not implement decodeHTTPWithoutAllRequest")
}

func errorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case ErrBadRequest:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}
