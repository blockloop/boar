package boar

import (
	"fmt"
	"net/http"
)

// HTTPError is an error that is responded back to the requestor
type HTTPError struct {
	status int
	Err    error `json:"error"`
}

// NewHTTPError creates a new HTTPError that will be marshaled to the requestor
func NewHTTPError(status int, cause error) *HTTPError {
	return &HTTPError{
		status: status,
		Err:    cause,
	}
}

// Status returns the status code to be used with this error
func (h *HTTPError) Status() int {
	return h.status
}

func (h *HTTPError) Error() string {
	return fmt.Sprintf("status: %d, error: %s", h.status, h.Err)
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
	HTTPError
}

// NewValidationError creates a new ValidationError
func NewValidationError(err error) *ValidationError {
	cause := NewHTTPError(http.StatusBadRequest, err)
	return &ValidationError{*cause}
}
