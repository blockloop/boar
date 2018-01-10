package boar

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSONSetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	c := newContext(nil, w, nil)
	c.WriteJSON(http.StatusOK, JSON{})
	assert.Equal(t, "application/json", w.Result().Header.Get("content-type"))
}

func TestWriteJSONReturnsErrorWhenJSONEncodeFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	w := NewMockResponseWriter(ctrl)
	expected := io.ErrClosedPipe
	w.EXPECT().Write(gomock.Any()).Return(0, expected)
	w.EXPECT().Header().Return(http.Header{})
	w.EXPECT().WriteHeader(gomock.Any())

	c := newContext(nil, nil, nil)
	c.response = w
	err := c.WriteJSON(http.StatusTeapot, JSON{})
	require.Error(t, err, "WriteJSON")
	assert.Contains(t, err.Error(), expected.Error())
}

func TestWriteJSONDoesNotReturnErrorWhenJSONEncodePasses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	w := NewMockResponseWriter(ctrl)
	w.EXPECT().Write(gomock.Any()).Return(0, nil)
	w.EXPECT().Header().Return(http.Header{})
	w.EXPECT().WriteHeader(gomock.Any())

	c := newContext(nil, nil, nil)
	c.response = w
	err := c.WriteJSON(http.StatusTeapot, JSON{})
	assert.NoError(t, err)
}

func TestWriteJSONSetsStatus(t *testing.T) {
	w := httptest.NewRecorder()
	c := newContext(nil, w, nil)

	require.NoError(t, c.WriteJSON(http.StatusTeapot, JSON{}))
	require.NoError(t, c.Response().Flush())
	w.Flush()

	assert.Equal(t, http.StatusTeapot, w.Result().StatusCode)
}

func TestWriteStatusSetsResponseStatus(t *testing.T) {
	w := httptest.NewRecorder()
	c := newContext(nil, w, nil)
	c.WriteStatus(http.StatusTeapot)
	c.Response().Flush()
	w.Flush()
	assert.Equal(t, http.StatusTeapot, w.Result().StatusCode)
}

func TestContextShouldReturnRequestContext(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	c := newContext(req, nil, nil)
	assert.Equal(t, req.Context(), c.Context())
}

func TestReadURLParamsShouldBindParamsFromRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	params := httprouter.Params{
		{Key: "name", Value: "Brett"},
		{Key: "age", Value: "123"},
	}

	c := newContext(req, nil, params)

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

	c := newContext(req, nil, nil)

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

	c := newContext(req, nil, nil)

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

	c := newContext(req, nil, nil)

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

	c := newContext(req, nil, nil)

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

	c := newContext(req, nil, nil)

	var fields struct {
		Name string
		Age  int
	}

	err = c.ReadQuery(&fields)
	require.IsType(t, &ValidationError{}, err)
}
