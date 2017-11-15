// Package boar provides HTTP middleware for semantic and organized HTTP server applications
package boar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blockloop/boar/bind"
	"github.com/gorilla/schema"
	"github.com/julienschmidt/httprouter"
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

	// ReadForm reads the contents of the request form and populates the values of v.
	ReadForm(v interface{}) error

	// WriteJSON writes the status code and then sends a json response message
	WriteJSON(status int, v interface{}) error

	// WriteStatus is an alias to c.Response().WriteHeader(status)
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
		urlParams:  ps,
		formParser: schema.NewDecoder(),
	}
}

type requestContext struct {
	response   http.ResponseWriter
	request    *http.Request
	urlParams  httprouter.Params
	formParser *schema.Decoder
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
	r.response.WriteHeader(status)
	return nil
}

func (r *requestContext) Request() *http.Request {
	return r.request
}

func (r *requestContext) Response() http.ResponseWriter {
	return r.response
}

func (r *requestContext) ReadJSON(v interface{}) error {
	if err := json.NewDecoder(r.Request().Body).Decode(v); err != nil {
		return NewValidationError(bodyField, err)
	}
	return nil
}

func (r *requestContext) ReadForm(v interface{}) error {
	if err := r.Request().ParseForm(); err != nil {
		return NewValidationError(bodyField, err)
	}

	if err := r.formParser.Decode(v, r.Request().Form); err != nil {
		return NewValidationError(bodyField, err)
	}
	return nil
}

func (r *requestContext) WriteJSON(status int, v interface{}) error {
	r.response.Header().Set("content-type", "application/json")
	r.response.WriteHeader(status)
	if err := json.NewEncoder(r.Response()).Encode(v); err != nil {
		return fmt.Errorf("could not encode JSON response: %+v", err)
	}
	return nil
}

func (r *requestContext) ReadQuery(v interface{}) error {
	if err := bind.Query(v, r.Request().URL.Query()); err != nil {
		return NewValidationError(queryField, err)
	}
	return nil
}
