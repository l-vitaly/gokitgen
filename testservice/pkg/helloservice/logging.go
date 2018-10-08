package helloservice

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type loggingService struct {
	next   Service
	logger log.Logger
}

func (s *loggingService) Say(name string) (message Message, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "Say",
			"stackTrace", getStackTrace(err),
			"name", name,
		)
	}(time.Now())

	return s.Say(name)
}

func (s *loggingService) WithoutParams() (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "WithoutParams",
			"stackTrace", getStackTrace(err),
		)
	}(time.Now())

	return s.WithoutParams()
}

func (s *loggingService) WithoutAll() {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "WithoutAll",
		)
	}(time.Now())

	s.WithoutAll()
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func getStackTrace(err error) string {
	if err, ok := err.(stackTracer); ok {
		return fmt.Sprintf("%+v\n", err.StackTrace())
	}
	return ""
}

// NewLoggingService creates a logging service middleware.
func NewLoggingService(next Service, logger log.Logger) Service {
	return &loggingService{next: next, logger: logger}
}
