package boar

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"

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
	if httperr.Status() > 499 {
		// log internal server errors
		log.Printf("ERROR: %+v", httperr)
	}

	if werr := c.WriteJSON(httperr.Status(), httperr); werr != nil {
		log.Printf("ERROR: could not serialize json: %s\n%s", werr, string(debug.Stack()))
	}
	return
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

// Router is an http router
type Router struct {
	r            *httprouter.Router
	errorHandler ErrorHandler
	mw           []Middleware
}

// RealRouter returns the httprouter.Router used for actual serving
func (rtr *Router) RealRouter() *httprouter.Router {
	return rtr.r
}

// Method is a path handler that uses a factory to generate the handler
// this is particularly useful for filling contextual information into a struct
// before passing it along to handle the request
func (rtr *Router) Method(method string, path string, createHandler HandlerProviderFunc) {
	fn := rtr.makeHandler(method, path, createHandler)

	rtr.RealRouter().Handle(method, path, fn)
}

func (rtr *Router) makeHandler(method string, path string, createHandler HandlerProviderFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c := newContext(r, w, ps)
		h, err := createHandler(c)
		if err != nil {
			rtr.errorHandler(c, err)
			return
		}
		if h == nil {
			rtr.errorHandler(c, errors.New("handler cannot be nil"))
			return
		}

		handlerValue := reflect.Indirect(reflect.ValueOf(h))

		if err := setQuery(handlerValue, r.URL.Query()); err != nil {
			rtr.errorHandler(c, err)
			return
		}

		if err := setURLParams(handlerValue, ps); err != nil {
			rtr.errorHandler(c, err)
			return
		}

		if r.ContentLength > 0 {
			if err := setBody(handlerValue, c); err != nil {
				rtr.errorHandler(c, err)
				return
			}
		}

		handle := rtr.withMiddlewares(h.Handle)
		if err := handle(c); err != nil {
			rtr.errorHandler(c, err)
			return
		}
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
			panic(fmt.Sprintf("cannot use nil middleware at %d: ", i))
		}
	}
	rtr.mw = append(mw, rtr.mw...)
}

func (rtr *Router) withMiddlewares(next HandlerFunc) HandlerFunc {
	fn := next
	for i := 0; i < len(rtr.mw); i++ {
		mw := rtr.mw[i]
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

// Connect is a handler that accepts only CONNECT requests
func (rtr *Router) Connect(path string, h HandlerProviderFunc) {
	rtr.Method(http.MethodConnect, path, h)
}

// ListenAndServe is a handler that accepts only LISTENANDSERVE requests
func (rtr *Router) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, rtr.RealRouter())
}

// SetErrorHandler sets the error handler. Any route that returns
// an error will get routed to this error handler
func (rtr *Router) SetErrorHandler(h ErrorHandler) {
	rtr.errorHandler = h
}

type simpleHandler struct {
	handle HandlerFunc
}

func (h *simpleHandler) Handle(c Context) error {
	return h.handle(c)
}
