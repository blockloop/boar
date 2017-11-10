package boar

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Middleware is a global middleware function
type Middleware func(HandlerFunc) HandlerFunc

// HandlerFunc is a function that handles an HTTP request
type HandlerFunc func(Context) error

// GetHandlerFunc is a prerequesite function that is used to generate handlers
// this is valuable to use like a factory
type GetHandlerFunc func(Context) (Handler, error)

// ErrorHandler is a handler func that is called when an error is returned from
// a route
type ErrorHandler func(Context, error)

// Handler is an http Handler
type Handler interface {
	Handle(Context) error
}

// NewRouterWithBase allows you to create a new http router with the provided
//  httprouter.Router instead of the default httprouter.New()
func NewRouterWithBase(r *httprouter.Router) *Router {
	return &Router{
		r:            r,
		errorHandler: defaultErrorHandler,
		mw:           make([]Middleware, 0),
	}
}

// NewRouter creates a new router for handling http requests
func NewRouter() *Router {
	return NewRouterWithBase(httprouter.New())
}

func defaultErrorHandler(c Context, err error) {
	httperr, ok := err.(*HTTPError)
	if !ok {
		httperr = &HTTPError{
			status: http.StatusInternalServerError,
			Err:    err,
		}
	}

	c.WriteJSON(httperr.Status(), httperr)
	return
}
