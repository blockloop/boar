package boar

import (
	"bytes"
	"io"
	"net/http"
)

// ResponseWriter is an http.ResponseWriter that captures the status code and body
// written for retrieval after the response has been sent
type ResponseWriter interface {
	http.ResponseWriter
	Status() int
	Body() []byte
	Len() int
}

var _ ResponseWriter = (*BufferedResponseWriter)(nil)

// BufferedResponseWriter is an http.ResponseWriter that captures the status code and body
// written for retrieval after the response has been sent
type BufferedResponseWriter struct {
	base   http.ResponseWriter
	body   *bytes.Buffer
	status int
}

// NewBufferedResponseWriter creates a new BufferedResponseWriter
func NewBufferedResponseWriter(base http.ResponseWriter) *BufferedResponseWriter {
	return &BufferedResponseWriter{
		base:   base,
		body:   bytes.NewBuffer(make([]byte, 0)),
		status: 0,
	}
}

// Status returns the currently set HTTP status code
func (w *BufferedResponseWriter) Status() int {
	return w.status
}

// Body returns everything that has been written to the response
func (w *BufferedResponseWriter) Body() []byte {
	return w.body.Bytes()
}

// Len returns the amount of bytes that have been written so far
func (w *BufferedResponseWriter) Len() int {
	return w.body.Len()
}

// Header returns the headers
func (w *BufferedResponseWriter) Header() http.Header {
	return w.base.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *BufferedResponseWriter) Write(b []byte) (n int, err error) {
	// This is what the http.ResponseWriter does by default. If the status code has
	// not already been set then it defaults to http.StatusOK
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return io.MultiWriter(w.body, w.base).Write(b)
}

// WriteHeader sends an HTTP response header with status code. If WriteHeader is
// not called explicitly, the first call to Write will trigger an implicit
// WriteHeader(http.StatusOK). Thus explicit calls to WriteHeader are mainly used
// to send error codes.
func (w *BufferedResponseWriter) WriteHeader(status int) {
	w.status = status
	w.base.WriteHeader(status)
}
