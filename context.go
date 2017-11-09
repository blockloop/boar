package boar

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/blockloop/boar/bind"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

// Context is an http handler context
type Context interface {

	// Context is a shortcut for Request().Context()
	Context() context.Context

	// Request returns the underlying http.Request
	Request() *http.Request

	// Response returns the underlying http.ResponseWriter
	Response() http.ResponseWriter

	// ReadQuery parses the query string from the request into a struct
	// if the query string has invalid types (e.g. alpha for an int field)
	// then a ValidationError will be returned with a status code of 400
	ReadQuery(v interface{}) error

	// ReadParams parses the url parameters into a struct
	// ReadParams(v interface{}) error

	// ReadForm(v interface{}) error
	ReadJSON(v interface{}) error

	// WriteJSON sets the status code and then sends a json response message
	WriteJSON(status int, v interface{}) error

	// Status returns the currently set status. This is useful for middlewares
	Status() int

	// WriteStatus sets the status and returns without a body
	WriteStatus(status int) error

	// URLParams returns all params as a key/value pair for quick lookups
	URLParams() httprouter.Params

	// ReadURLParams maps all URL parameters to struct fields of v and returns
	// a validation error if there are any type mismatches
	ReadURLParams(v interface{}) error
}

// NewContext creates a new Context based on the rquest and response writer given
func NewContext(r *http.Request, w http.ResponseWriter, ps httprouter.Params) Context {
	return newContext(r, w, ps)
}

func newContext(r *http.Request, w http.ResponseWriter, ps httprouter.Params) *requestContext {
	return &requestContext{
		response:   w,
		request:    r,
		onceStatus: &sync.Once{},
		status:     http.StatusOK,
		urlParams:  ps,
	}
}

type requestContext struct {
	response   http.ResponseWriter
	request    *http.Request
	status     int
	onceStatus *sync.Once
	urlParams  httprouter.Params
}

func (r *requestContext) Context() context.Context {
	return r.Request().Context()
}

func (r *requestContext) ReadURLParams(v interface{}) error {
	return bind.Params(v, r.URLParams())
}

func (r *requestContext) URLParams() httprouter.Params {
	return r.urlParams
}

func (r *requestContext) WriteStatus(status int) error {
	r.status = status
	r.response.WriteHeader(status)
	return nil
}

func (r *requestContext) Status() int {
	return r.status
}

func (r *requestContext) Request() *http.Request {
	return r.request
}

func (r *requestContext) Response() http.ResponseWriter {
	return r.response
}

func (r *requestContext) ReadJSON(v interface{}) error {
	err := json.NewDecoder(r.Request().Body).Decode(v)
	if err != nil {
		return NewHTTPError(http.StatusBadRequest, err)
	}

	ok, err := govalidator.ValidateStruct(v)
	if !ok {
		return NewValidationError(err)
	}
	return nil
}

func (r *requestContext) WriteJSON(status int, v interface{}) error {
	r.status = status
	r.response.Header().Set("content-type", "application/json")
	r.response.WriteHeader(status)
	err := json.NewEncoder(r.Response()).Encode(v)
	return errors.Wrap(err, "could not encode JSON response")
}

func (r *requestContext) ReadQuery(v interface{}) error {
	if err := bind.Query(v, r.Request().URL.Query()); err != nil {
		return NewValidationError(err)
	}
	return nil
}
