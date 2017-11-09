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

// MethodFunc is a path handler that uses a factory to generate the handler
// this is particularly useful for filling contextual information into a struct
// before passing it along to handle the request
func (rtr *Router) MethodFunc(method string, path string, createHandler GetHandlerFunc) {
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

func (rtr *Router) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, rtr.RealRouter())
}

// SetErrorHandler sets the error handler. Any route that returns
// an error will get routed to this error handler
func (rtr *Router) SetErrorHandler(h ErrorHandler) {
	rtr.errorHandler = h
}
