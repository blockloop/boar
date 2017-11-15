package boar

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	. "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWriteJSONSetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	c := NewContext(nil, w, nil)
	c.WriteJSON(http.StatusTeapot, JSON{})
	w.Flush()
	assert.Equal(t, w.Result().Header.Get("content-type"), "application/json")
}

func TestWriteJSONReturnsErrorWhenJSONEncodeFails(t *testing.T) {
	w := &MockResponseWriter{}
	expected := io.ErrClosedPipe
	w.On("Write", Anything).Return(0, expected)
	w.On("Header", Anything).Return(http.Header{})
	w.On("WriteHeader", Anything).Return(0)

	c := NewContext(nil, w, nil)
	err := c.WriteJSON(http.StatusTeapot, JSON{})
	require.Error(t, err, "WriteJSON")
	assert.Contains(t, err.Error(), expected.Error())
}

func TestWriteJSONDoesNotReturnErrorWhenJSONEncodePasses(t *testing.T) {
	w := &MockResponseWriter{}
	w.On("Write", Anything).Return(0, nil)
	w.On("Header", Anything).Return(http.Header{})
	w.On("WriteHeader", Anything).Return(0)

	c := NewContext(nil, w, nil)
	err := c.WriteJSON(http.StatusTeapot, JSON{})
	assert.NoError(t, err)
}

func TestWriteJSONSetsStatus(t *testing.T) {
	w := httptest.NewRecorder()
	c := NewContext(nil, w, nil)
	c.WriteJSON(http.StatusTeapot, JSON{})
	w.Flush()
	assert.Equal(t, http.StatusTeapot, w.Result().StatusCode)
}

func TestWriteStatusSetsResponseStatus(t *testing.T) {
	w := httptest.NewRecorder()
	c := NewContext(nil, w, nil)
	c.WriteStatus(http.StatusTeapot)
	w.Flush()
	assert.Equal(t, http.StatusTeapot, w.Result().StatusCode)
}

func TestContextShouldReturnRequestContext(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	c := NewContext(req, nil, nil)
	assert.Equal(t, req.Context(), c.Context())
}

func TestReadURLParamsShouldBindParamsFromRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	params := httprouter.Params{
		{Key: "name", Value: "Brett"},
		{Key: "age", Value: "123"},
	}

	c := NewContext(req, nil, params)

	var fields struct {
		Name string `url:"name"`
		Age  int    `url:"age"`
	}

	err = c.ReadURLParams(&fields)
	require.NoError(t, err)

	assert.Equal(t, fields.Name, "Brett")
	assert.Equal(t, fields.Age, 123)
}

func TestReadFormShouldBindFormFields(t *testing.T) {
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString("Name=Brett&Age=123"))
	require.NoError(t, err)
	req.Header.Set("content-type", contentTypeFormEncoded)

	c := NewContext(req, nil, nil)

	var fields struct {
		Name string
		Age  int
	}

	err = c.ReadForm(&fields)
	require.NoError(t, err)

	assert.Equal(t, fields.Name, "Brett")
	assert.Equal(t, fields.Age, 123)
}

func TestReadFormReturnsErrorIfParseFormFails(t *testing.T) {
	req, err := http.NewRequest("POST", "/", nil)
	require.NoError(t, err)
	req.Header.Set("content-type", contentTypeFormEncoded)

	c := NewContext(req, nil, nil)

	var fields struct {
		Name string
	}

	err = c.ReadForm(&fields)
	assert.IsType(t, &ValidationError{}, err)
}

func TestReadFormReturnsErrorIfMismatchTypes(t *testing.T) {
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString("Name=Brett&Age=abcd"))
	require.NoError(t, err)
	req.Header.Set("content-type", contentTypeFormEncoded)

	c := NewContext(req, nil, nil)

	var fields struct {
		Name string
		Age  int
	}

	err = c.ReadForm(&fields)
	assert.IsType(t, &ValidationError{}, err)
}

func TestReadQueryShouldBindFields(t *testing.T) {
	req, err := http.NewRequest("GET", "/?Name=Brett&Age=123", nil)
	require.NoError(t, err)

	c := NewContext(req, nil, nil)

	var fields struct {
		Name string
		Age  int
	}

	err = c.ReadQuery(&fields)
	require.NoError(t, err)

	assert.Equal(t, fields.Name, "Brett")
	assert.Equal(t, fields.Age, 123)
}

func TestReadQueryShouldReturnValidationErrorIfQueryTypesMismatch(t *testing.T) {
	req, err := http.NewRequest("GET", "/?Name=Brett&Age=abcd", nil)
	require.NoError(t, err)

	c := NewContext(req, nil, nil)

	var fields struct {
		Name string
		Age  int
	}

	err = c.ReadQuery(&fields)
	require.IsType(t, &ValidationError{}, err)
}
