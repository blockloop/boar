package boar

import (
	"log"
	"net/http"
	"runtime/debug"
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

func defaultErrorHandler(c Context, err error) {
	if err == nil {
		return
	}

	httperr, ok := err.(HTTPError)
	if !ok {
		httperr = NewHTTPError(http.StatusInternalServerError, err)
	}

	if err := c.WriteJSON(httperr.Status(), httperr); err != nil {
		log.Printf("ERROR: could not serialize json: %s\n%s", err, string(debug.Stack()))
	}
	return
}
