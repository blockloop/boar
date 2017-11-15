package boar

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPErrorShouldPrintCause(t *testing.T) {
	err := &httpError{
		cause:  io.ErrClosedPipe,
		status: 100,
	}
	str := err.Error()
	assert.Contains(t, str, io.ErrClosedPipe.Error())
}

func TestHTTPErrorShouldPrintStatus(t *testing.T) {
	err := &httpError{
		cause:  io.ErrClosedPipe,
		status: 100,
	}
	str := err.Error()
	assert.Contains(t, str, "100")
}

func TestHTTPErrorMarshalJSONCreatesErrorFieldWithCause(t *testing.T) {
	e := &httpError{
		status: 500,
		cause:  io.ErrClosedPipe,
	}

	byts, err := e.MarshalJSON()
	assert.NoError(t, err)

	var ermsg struct {
		Error string `json:"error"`
	}

	require.NoError(t, json.Unmarshal(byts, &ermsg))

	assert.Equal(t, io.ErrClosedPipe.Error(), ermsg.Error)
}

func TestValidationErrorMarshalJSONCreatesErrorFieldWithFieldNameWithCauses(t *testing.T) {
	e := &ValidationError{
		status:    500,
		fieldName: "query",
		Errors:    []error{io.ErrClosedPipe},
	}

	byts, err := e.MarshalJSON()
	assert.NoError(t, err)

	var ermsg struct {
		Errors struct {
			Query []string `json:"query"`
		} `json:"errors"`
	}

	require.NoError(t, json.Unmarshal(byts, &ermsg))

	assert.Equal(t, []string{io.ErrClosedPipe.Error()}, ermsg.Errors.Query)
}

func TestValidationErrorCauseShowsAllCauses(t *testing.T) {
	ers := []error{io.ErrClosedPipe, os.ErrInvalid}
	e := &ValidationError{
		status:    500,
		fieldName: "query",
		Errors:    ers,
	}

	err := e.Cause().Error()
	for _, er := range ers {
		assert.Contains(t, err, er.Error())
	}
}
