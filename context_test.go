package boar

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blockloop/boar/mocks"
	"github.com/stretchr/testify/assert"
	. "github.com/stretchr/testify/mock"
)

func TestWriteJSONSetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	c := NewContext(nil, w, nil)
	c.WriteJSON(http.StatusTeapot, JSON{})
	w.Flush()
	assert.Equal(t, w.Result().Header.Get("content-type"), "application/json")
}

func TestWriteJSONReturnsErrorWhenJSONEncodeFails(t *testing.T) {
	w := &mocks.ResponseWriter{}
	expected := io.ErrClosedPipe
	w.On("Write", Anything).Return(0, expected)
	w.On("Header", Anything).Return(http.Header{})
	w.On("WriteHeader", Anything).Return(0)

	c := NewContext(nil, w, nil)
	err := c.WriteJSON(http.StatusTeapot, JSON{})
	assert.Error(t, err)
}

func TestWriteJSONDoesNotReturnErrorWhenJSONEncodePasses(t *testing.T) {
	w := &mocks.ResponseWriter{}
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
