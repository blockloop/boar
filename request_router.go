package boar

import (
	"net/http"

	"github.com/apex/log"
	"github.com/julienschmidt/httprouter"
)

// Router is an http router
type Router struct {
	r            *httprouter.Router
	errorHandler ErrorHandler
	log          log.Interface
}

// RealRouter returns the httprouter.Router used for actual serving
func (rtr *Router) RealRouter() *httprouter.Router {
	return rtr.r
}

// Method is a path handler that uses a factory to generate the handler
// this is particularly useful for filling contextual information into a struct
// before passing it along to handle the request
func (rtr *Router) Method(method string, path string, createHandler GetHandlerFunc) {
	rtr.RealRouter().Handle(method, path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c := newContext(r, w, ps)
		h, err := createHandler(c)
		if err != nil {
			rtr.errorHandler(c, err)
			return
		}

		if before, ok := h.(BeforeHandler); ok {
			if err := before.Before(c); err != nil {
				rtr.errorHandler(c, err)
				return
			}
		}

		if err := h.Handle(c); err != nil {
			rtr.errorHandler(c, err)
			return
		}

		if after, ok := h.(AfterHandler); ok {
			if err := after.After(c); err != nil {
				rtr.errorHandler(c, err)
				return
			}
		}
	})
}

// Head is a handler that acceps HEAD requests
func (rtr *Router) Head(path string, h GetHandlerFunc) {
	rtr.Method(http.MethodHead, path, h)
}

// Trace is a handler that accepts only TRACE requests
func (rtr *Router) Trace(path string, h GetHandlerFunc) {
	rtr.Method(http.MethodTrace, path, h)
}

// Delete is a handler that accepts only DELETE requests
func (rtr *Router) Delete(path string, h GetHandlerFunc) {
	rtr.Method(http.MethodDelete, path, h)
}

// Options is a handler that accepts only OPTIONS requests
// It is not recommended to use this as the router automatically
// handles OPTIONS requests by default
func (rtr *Router) Options(path string, h GetHandlerFunc) {
	rtr.Method(http.MethodOptions, path, h)
}

// Get is a handler that accepts only GET requests
func (rtr *Router) Get(path string, h GetHandlerFunc) {
	rtr.Method(http.MethodGet, path, h)
}

// Put is a handler that accepts only PUT requests
func (rtr *Router) Put(path string, h GetHandlerFunc) {
	rtr.Method(http.MethodPut, path, h)
}

// Post is a handler that accepts only POST requests
func (rtr *Router) Post(path string, h GetHandlerFunc) {
	rtr.Method(http.MethodPost, path, h)
}

// Patch is a handler that accepts only PATCH requests
func (rtr *Router) Patch(path string, h GetHandlerFunc) {
	rtr.Method(http.MethodPatch, path, h)
}

// Connect is a handler that accepts only CONNECT requests
func (rtr *Router) Connect(path string, h GetHandlerFunc) {
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
