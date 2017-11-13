package boar

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCreatesEmptyBuffer(t *testing.T) {
	w := NewBufferedResponseWriter(nil)
	assert.NotNil(t, w.body)
	assert.Len(t, w.body.Bytes(), 0)
}

func TestFlushWritesHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewBufferedResponseWriter(rec)
	w.status = 100
	assert.NoError(t, w.Flush())
	rec.Flush()
	assert.Equal(t, 100, rec.Result().StatusCode)
}

func TestFlushCopiesBodyToBase(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewBufferedResponseWriter(rec)

	exp := "hello"
	w.body = bytes.NewBufferString(exp)

	require.NoError(t, w.Flush())
	rec.Flush()

	body, err := ioutil.ReadAll(rec.Result().Body)
	require.NoError(t, err)

	assert.Equal(t, exp, string(body))
}

func TestWriteSetsStatusToOKIfUnset(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewBufferedResponseWriter(rec)
	fmt.Fprintf(w, "")

	assert.Equal(t, w.status, http.StatusOK)
}

func TestWriteDoesNotSetStatusIfAlreadySet(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewBufferedResponseWriter(rec)

	exp := 199
	w.status = exp

	fmt.Fprintf(w, "")

	assert.Equal(t, w.status, exp)
}
