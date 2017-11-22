package boar

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

// JSON is a shortcut for map[string]interface{}
type JSON map[string]interface{}

// Middleware is a global middleware function
type Middleware func(HandlerFunc) HandlerFunc

// HandlerFunc is a function that handles an HTTP request
type HandlerFunc func(Context) error

// HandlerProviderFunc is a prerequesite function that is used to generate handlers
// this is valuable to use like a factory
type HandlerProviderFunc func(Context) (Handler, error)

// Handler is an http Handler
type Handler interface {
	Handle(Context) error
}

// ErrorHandler is a middleware that handles writing errors back to the client when an error
// an error occurs in the handler. It is the first middleware executed therefore It should
// always return the error that it handled
//
// ErrorHandler *must be set before the router is declared. When calling NewRouter the
// ErrorHandler is added as the first middleware in the chain. If ErrorHandler is modified
// after the router has been created then the old func will be used
var ErrorHandler Middleware = func(next HandlerFunc) HandlerFunc {
	return func(c Context) error {
		err := next(c)
		if err == nil {
			return nil
		}

		httperr, ok := err.(HTTPError)
		if !ok {
			httperr = NewHTTPError(http.StatusInternalServerError, err)
		}

		if c.Response().Len() == 0 {
			werr := c.WriteJSON(httperr.Status(), httperr)
			if werr != nil {
				return fmt.Errorf("unable to serialize JSON to response: %s", werr)
			}
		}

		return httperr
	}
}

// NewRouterWithBase allows you to create a new http router with the provided
//  httprouter.Router instead of the default httprouter.New()
func NewRouterWithBase(r *httprouter.Router) *Router {
	return &Router{
		base:        r,
		middlewares: []Middleware{ErrorHandler},
	}
}

// NewRouter creates a new router for handling http requests
func NewRouter() *Router {
	return NewRouterWithBase(httprouter.New())
}

// Router is an http router
type Router struct {
	base        *httprouter.Router
	middlewares []Middleware
}

// RealRouter returns the httprouter.Router used for actual serving
func (rtr *Router) RealRouter() *httprouter.Router {
	return rtr.base
}

func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rtr.RealRouter().ServeHTTP(w, r)
}

// Method is a path handler that uses a factory to generate the handler
// this is particularly useful for filling contextual information into a struct
// before passing it along to handle the request
func (rtr *Router) Method(method string, path string, createHandler HandlerProviderFunc) {
	rtr.RealRouter().Handle(method, path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c := newContext(r, w, ps)
		defer c.Response().Flush()

		wrappedHandler := withMiddlewares(rtr.middlewares, requestParserMiddleware(createHandler))
		wrappedHandler(c)
	})
}

// requestParserMiddleware provides the handler with request objects populated by request data such
// as query string, post body, and url parameters
func requestParserMiddleware(createHandler HandlerProviderFunc) HandlerFunc {
	return func(c Context) error {
		handler, err := createHandler(c)
		if err != nil {
			return err
		}
		if handler == nil {
			log.Panicf("nil handler provided for %q %q", c.Request().Method, c.Request().URL.Path)
		}

		handlerValue := reflect.Indirect(reflect.ValueOf(handler))

		if err := setQuery(handlerValue, c.Request().URL.Query()); err != nil {
			return err
		}

		if err := setURLParams(handlerValue, c.URLParams()); err != nil {
			return err
		}

		if err := setBody(handlerValue, c); err != nil {
			return err
		}
		return handler.Handle(c)
	}
}

// MethodFunc sets a HandlerFunc for a url with the given method. It is used for
// simple handlers that do not require any building. This is not a recommended
// for common use cases
func (rtr *Router) MethodFunc(method string, path string, h HandlerFunc) {
	rtr.Method(method, path, func(Context) (Handler, error) {
		return &simpleHandler{handle: h}, nil
	})
}

// Use injects a middleware into the http requests. They are executed in the
// order in which they are added.
func (rtr *Router) Use(mw ...Middleware) {
	if len(mw) == 0 {
		return
	}

	for i, m := range mw {
		if m == nil {
			log.Panicf("cannot use nil middleware at position %d: ", i)
		}
	}
	rtr.middlewares = append(rtr.middlewares, mw...)
}

func withMiddlewares(mws []Middleware, next HandlerFunc) HandlerFunc {
	fn := next
	for _, mw := range mws {
		fn = mw(fn)
	}
	return fn
}

// Head is a handler that acceps HEAD requests
func (rtr *Router) Head(path string, h HandlerProviderFunc) {
	rtr.Method(http.MethodHead, path, h)
}

// Trace is a handler that accepts only TRACE requests
func (rtr *Router) Trace(path string, h HandlerProviderFunc) {
	rtr.Method(http.MethodTrace, path, h)
}

// Delete is a handler that accepts only DELETE requests
func (rtr *Router) Delete(path string, h HandlerProviderFunc) {
	rtr.Method(http.MethodDelete, path, h)
}

// Options is a handler that accepts only OPTIONS requests
// It is not recommended to use this as the router automatically
// handles OPTIONS requests by default
func (rtr *Router) Options(path string, h HandlerProviderFunc) {
	rtr.Method(http.MethodOptions, path, h)
}

// Get is a handler that accepts only GET requests
func (rtr *Router) Get(path string, h HandlerProviderFunc) {
	rtr.Method(http.MethodGet, path, h)
}

// Put is a handler that accepts only PUT requests
func (rtr *Router) Put(path string, h HandlerProviderFunc) {
	rtr.Method(http.MethodPut, path, h)
}

// Post is a handler that accepts only POST requests
func (rtr *Router) Post(path string, h HandlerProviderFunc) {
	rtr.Method(http.MethodPost, path, h)
}

// Patch is a handler that accepts only PATCH requests
func (rtr *Router) Patch(path string, h HandlerProviderFunc) {
	rtr.Method(http.MethodPatch, path, h)
}

type simpleHandler struct {
	handle HandlerFunc
}

func (h *simpleHandler) Handle(c Context) error {
	return h.handle(c)
}
