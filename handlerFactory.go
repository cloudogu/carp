package carp

import (
	"fmt"
	"net/http"
)

type handlerFactory struct {
	conf     Configuration
	handlers map[string]http.Handler
}

func createHandlerFactory() handlerFactory {
	return handlerFactory{
		handlers: make(map[string]http.Handler),
	}
}
func (f *handlerFactory) add(handlerId string, handler http.Handler) {
	switch handlerId {
	case handlerFactoryCasHandler:
		fallthrough
	case handlerFactoryRestHandler:
		f.handlers[handlerId] = handler
	default:
		panic("unknown request handler ID " + handlerId)
	}
}

func (f *handlerFactory) get(handlerId string) (http.Handler, error) {
	switch handlerId {
	case handlerFactoryCasHandler:
		fallthrough
	case handlerFactoryRestHandler:
		return f.handlers[handlerId], nil
	default:
		return nil, fmt.Errorf("unknown request handler ID " + handlerId)
	}
}
