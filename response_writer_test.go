package boar

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestNewCreatesEmptyBuffer(t *testing.T) {
	w := NewBufferedResponseWriter(nil)
	assert.Len(t, w.Body(), 0)
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
	w.body = []byte(exp)

	require.NoError(t, w.Flush())
	rec.Flush()

	body, err := ioutil.ReadAll(rec.Result().Body)
	require.NoError(t, err)

	assert.Equal(t, exp, string(body))
}

func TestCloseFlushesAndClosesTheBuffer(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewBufferedResponseWriter(rec)

	w.Write([]byte("asldfhasdf"))
	require.NoError(t, w.Close())

	assert.Empty(t, w.Body())
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

	assert.Equal(t, w.Status(), exp)
}

func TestBodyReturnsWrittenBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewBufferedResponseWriter(rec)

	exp := "kajshdfalsdf"
	fmt.Fprintf(w, exp)

	assert.Equal(t, string(w.Body()), exp)
}

func TestLenReturnsWrittenBytesLength(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewBufferedResponseWriter(rec)

	exp := []byte("alksjdhflkajsdf")
	w.Write(exp)

	assert.Equal(t, w.Len(), len(exp))
}

func TestWriteHeaderDoesNotWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewBufferedResponseWriter(rec)

	w.WriteHeader(100)
	rec.Flush()

	assert.NotEqual(t, rec.Result().StatusCode, 100)
}

func TestWriteHeaderSetsStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewBufferedResponseWriter(rec)

	w.WriteHeader(100)

	assert.Equal(t, w.Status(), 100)
}

func TestHeaderSetsHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewBufferedResponseWriter(rec)

	w.Header().Set("hello", "world")
	rec.Flush()

	val := rec.Result().Header.Get("hello")
	assert.Equal(t, "world", val)
}
