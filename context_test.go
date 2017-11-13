package boar

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteJSONSetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	c := NewContext(nil, w, nil)
	c.WriteJSON(http.StatusTeapot, JSON{})
	w.Flush()
	assert.Equal(t, w.Result().Header.Get("content-type"), "application/json")
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
