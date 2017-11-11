package boar

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"github.com/asaskevich/govalidator"
	"github.com/blockloop/boar/bind"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

const (
	// QueryField is the name of the field for query parameters
	QueryField = "Query"
	// URLParamsField is the name of the field for the URL parameters
	URLParamsField = "URLParams"
	// BodyField is the name of the field for the Body parameters
	BodyField = "Body"
)

// JSON is a shortcut for map[string]interface{}
type JSON map[string]interface{}

// Router is an http router
type Router struct {
	r            *httprouter.Router
	errorHandler ErrorHandler
	mw           []Middleware
}

func nopHandler(Context) error {
	return nil
}

// RealRouter returns the httprouter.Router used for actual serving
func (rtr *Router) RealRouter() *httprouter.Router {
	return rtr.r
}

func setQuery(handler reflect.Value, qs url.Values) error {
	qf := handler.FieldByName(QueryField)
	if !qf.IsValid() {
		return nil
	}
	if qf.Kind() != reflect.Struct {
		return fmt.Errorf("%q field of %q must be a struct", QueryField, handler.Type().Name())
	}
	if !qf.CanSet() {
		return fmt.Errorf("%q field of %q is not setable", QueryField, handler.Type().Name())
	}
	if err := bind.QueryValue(qf, qs); err != nil {
		return err
	}
	ok, err := govalidator.ValidateStruct(qf.Addr().Interface())
	if !ok {
		verr, ok := err.(govalidator.Errors)
		if !ok {
			return verr
		}
		return NewValidationErrors(QueryField, verr.Errors())
	}
	return nil
}

func setURLParams(handler reflect.Value, params httprouter.Params) error {
	pf := handler.FieldByName(URLParamsField)
	if !pf.IsValid() {
		return nil
	}
	if pf.Kind() != reflect.Struct {
		return fmt.Errorf("%q field of %q must be a struct", QueryField, handler.Type().Name())
	}
	if !pf.CanSet() {
		return fmt.Errorf("%q field of %q is not setable", QueryField, handler.Type().Name())
	}
	if err := bind.ParamsValue(pf, params); err != nil {
		return err
	}
	ok, err := govalidator.ValidateStruct(pf.Addr().Interface())
	if !ok {
		verr, ok := err.(govalidator.Errors)
		if !ok {
			return verr
		}
		return NewValidationErrors(URLParamsField, verr.Errors())
	}
	return nil
}

func setBody(handler reflect.Value, c Context) error {
	bf := handler.FieldByName(BodyField)
	if !bf.IsValid() {
		return nil
	}
	if bf.Kind() != reflect.Struct {
		return fmt.Errorf("%q field of %q must be a struct", BodyField, handler.Type().Name())
	}
	if !bf.CanSet() {
		return fmt.Errorf("%q field of %q is not setable", BodyField, handler.Type().Name())
	}
	if err := c.ReadJSON(bf.Addr().Interface()); err != nil {
		return err
	}
	ok, err := govalidator.ValidateStruct(bf.Addr().Interface())
	if !ok {
		verr, ok := err.(govalidator.Errors)
		if !ok {
			return verr
		}
		return NewValidationErrors(BodyField, verr.Errors())
	}
	return nil
}

// Method is a path handler that uses a factory to generate the handler
// this is particularly useful for filling contextual information into a struct
// before passing it along to handle the request
func (rtr *Router) Method(method string, path string, createHandler GetHandlerFunc) {
	fn := func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	rtr.RealRouter().Handle(method, path, fn)
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
func (rtr *Router) Use(mw Middleware) {
	rtr.mw = append(rtr.mw, mw)
}

func (rtr *Router) withMiddlewares(next HandlerFunc) HandlerFunc {
	fn := next
	for i := len(rtr.mw) - 1; i >= 0; i-- {
		mw := rtr.mw[i]
		fn = mw(fn)
	}
	return fn
}

type simpleHandler struct {
	handle HandlerFunc
}

func (h *simpleHandler) Handle(c Context) error {
	return h.handle(c)
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
