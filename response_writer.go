package boar

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

// ResponseWriter is an http.ResponseWriter that captures the status code and body
// written for retrieval after the response has been sent
type ResponseWriter interface {
	http.ResponseWriter
	io.Closer

	Flush() error
	Status() int
	Body() []byte
	Len() int
}

var _ ResponseWriter = (*BufferedResponseWriter)(nil)

// BufferedResponseWriter is an http.ResponseWriter that captures the status code and body
// written for retrieval after the response has been sent
type BufferedResponseWriter struct {
	base      http.ResponseWriter
	m         *sync.RWMutex
	body      []byte
	status    int
	flushOnce *sync.Once
}

// NewBufferedResponseWriter creates a new BufferedResponseWriter
func NewBufferedResponseWriter(base http.ResponseWriter) *BufferedResponseWriter {
	return &BufferedResponseWriter{
		base:      base,
		m:         &sync.RWMutex{},
		body:      make([]byte, 0),
		status:    0,
		flushOnce: &sync.Once{},
	}
}

// Flush flushes the buffer into the write stream and sends the body to the client
// Once flush has been called, there can be no more headers or anything sent to the client
//
// Flush is called internally by the Router once all middlewares, handlers, and error handlers
// have completely executed. This allows the middlewares access to writing headers, reading
// contents, etc.
func (w *BufferedResponseWriter) Flush() (err error) {
	w.m.RLock()
	defer w.m.RUnlock()
	w.flushOnce.Do(func() {
		w.base.WriteHeader(w.Status())
		_, err = io.Copy(w.base, bytes.NewBuffer(w.body))
	})
	return err
}

// Close flushes the response stream and closes all buffers. Subsequent calls to Body(), Len(),
// etc will yield no results
func (w *BufferedResponseWriter) Close() error {
	defer func() {
		w.m.Lock()
		w.body = make([]byte, 0)
		w.m.Unlock()
	}()
	return w.Flush()
}

// Status returns the currently set HTTP status code
func (w *BufferedResponseWriter) Status() int {
	return w.status
}

// Body returns everything that has been written to the response
func (w *BufferedResponseWriter) Body() []byte {
	w.m.RLock()
	defer w.m.RUnlock()
	return w.body
}

// Len returns the amount of bytes that have been written so far
func (w *BufferedResponseWriter) Len() int {
	w.m.RLock()
	defer w.m.RUnlock()
	return len(w.body)
}

// Header returns the header map that will be sent by WriteHeader. The Header map
// also is the mechanism with which Handlers can set HTTP trailers.
//
// Unlike the default http.ResponseWriter, headers can be written *after* the response
// body has been written because the response body is buffered and therefore not written
// to the stream until Flush or Close is called
func (w *BufferedResponseWriter) Header() http.Header {
	return w.base.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *BufferedResponseWriter) Write(b []byte) (n int, err error) {
	w.m.Lock()
	defer w.m.Unlock()
	// This is what the http.ResponseWriter does by default. If the status code has
	// not already been set then it defaults to http.StatusOK
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.body = append(w.body, b...)
	return len(b), nil
}

// WriteHeader sets the http status code. Unlike the default http.ResponseWriter, this does
// _not_ begin the response transaction. This will simply store the status code until Flush
// is executed
func (w *BufferedResponseWriter) WriteHeader(status int) {
	w.status = status
}
