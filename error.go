package boar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

var (
	// ErrUnauthorized is an HTTPError for StatusUnauthorized
	ErrUnauthorized = NewHTTPErrorStatus(http.StatusUnauthorized)

	// ErrForbidden is an HTTPError for StatusForbidden
	ErrForbidden = NewHTTPErrorStatus(http.StatusForbidden)

	// ErrNotFound is an HTTPError for StatusNotFound
	ErrNotFound = NewHTTPErrorStatus(http.StatusNotFound)

	// ErrNotAcceptable is an HTTPError for StatusNotAcceptable
	ErrNotAcceptable = NewHTTPErrorStatus(http.StatusNotAcceptable)

	// ErrUnsupportedMediaType is an HTTPError for StatusUnsupportedMediaType
	ErrUnsupportedMediaType = NewHTTPErrorStatus(http.StatusUnsupportedMediaType)

	// ErrGone is an HTTPError for StatusGone
	ErrGone = NewHTTPErrorStatus(http.StatusGone)

	// ErrTooManyRequests is an HTTPError for StatusTooManyRequests
	ErrTooManyRequests = NewHTTPErrorStatus(http.StatusTooManyRequests)

	// ErrEntityNotFound should be used to provide a more valuable 404 error
	// message to the client. Simply sending 404 with no body to the client
	// is confusing because it is not clear what was not found. Was the path
	// incorrect or was there simply no item in the datastore? ErrEntityNotFound
	// provides a distinction when URLs are currect, but there is simply
	// no record in the datastore
	ErrEntityNotFound = NewHTTPError(http.StatusNotFound, fmt.Errorf("entity not found"))
)

// HTTPError is an error that is communicated
type HTTPError interface {
	error
	Cause() error
	Status() int
	json.Marshaler
}

type httpError struct {
	status int
	cause  error
}

// NewHTTPErrorStatus creates a new HTTP Error with the given status code and
// uses the default status text for that status code. These are useful for concise
// errors such as "Forbidden" or "Unauthorized"
func NewHTTPErrorStatus(status int) error {
	return NewHTTPError(status, fmt.Errorf(http.StatusText(status)))
}

// NewHTTPError creates a new HTTPError that will be marshaled to the requestor
func NewHTTPError(status int, cause error) HTTPError {
	return &httpError{
		status: status,
		cause:  cause,
	}
}

// Status returns the status code to be used with this error
func (h *httpError) Status() int {
	return h.status
}

func (h *httpError) Cause() error {
	return h.cause
}

func (h *httpError) Error() string {
	return fmt.Sprintf("HTTPError: (status: %d, error: %s)", h.status, h.cause)
}

// MarshalJSON marshals this error to JSON
func (h *httpError) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"error": "%s"}`, h.cause.Error())), nil
}

// ValidationError is an HTTPError that was caused by validation. Validation
// errors are typically caused by valid tags or improper type mapping between
// input types and struct fields. These should always be considered 400 errors.
// This is useful when you want to control the flow of validation errors within
// your handlers.
//
// Example:
//    func Handle(c Context) error {
//        err := c.ReadJSON(&req)
//        if err != nil {
//            if ok, verr := err.(*ValidationError); ok {
//                return c.WriteJSON(http.StatusBadRequest, map[string]interface{}{
//                    "validationErrors": err.Error(),
//                })
//            }
//            return err
//        }
//    }
//
//
//
type ValidationError struct {
	location string
	status   int
	Errors   []error
}

var _ HTTPError = (*ValidationError)(nil)

// NewValidationError creates a new Validation error with a single reason.
// location is the area where the validation failed. It should be QueryField,
// BodyField, or URLParamsField
func NewValidationError(location string, err error) *ValidationError {
	return NewValidationErrors(location, []error{err})
}

// NewValidationErrors creates a new Validation error with reasons.
// location is the area where the validation failed. It should be QueryField,
// BodyField, or URLParamsField
func NewValidationErrors(location string, errs []error) *ValidationError {
	return &ValidationError{
		location: location,
		status:   http.StatusBadRequest,
		Errors:   errs,
	}
}

// Status is the http status to be used for responding to the client
func (e *ValidationError) Status() int {
	return e.status
}

// Cause is the underlying cause(s) of the validation error
func (e *ValidationError) Cause() error {
	return errors.New(e.Error())
}

func (e *ValidationError) Error() string {
	s := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		s[i] = err.Error()
	}
	return strings.Join(s, "; ")
}

// MarshalJSON allows overrides json.Marshal default behavior
func (e *ValidationError) MarshalJSON() ([]byte, error) {
	ers := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		ers[i] = err.Error()
	}

	return json.Marshal(JSON{
		"errors": JSON{
			e.location: ers,
		},
	})
}
