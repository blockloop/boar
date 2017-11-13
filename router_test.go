package boar

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDefaultErrorHandlerShouldDoNothingToForNilError(t *testing.T) {
	mc := &MockContext{}
	defaultErrorHandler(mc, nil)
	mc.AssertNotCalled(t, "WriteJSON")
}

func TestDefaultErrorHandlerShouldWriteExistingHTTPError(t *testing.T) {
	mc := &MockContext{}
	status := 400
	err := NewHTTPErrorStatus(status)
	mc.On("WriteJSON", Anything, Anything).Return(nil).Run(func(args Arguments) {
		assert.Equal(t, status, args.Get(0))
		assert.Equal(t, err, args.Get(1))
	})

	defaultErrorHandler(mc, err)
}

func TestDefaultErrorHandlerShouldLogErrIfWriteJSONFails(t *testing.T) {
	mc := &MockContext{}
	status := 400
	err := NewHTTPErrorStatus(status)
	writeErr := errors.New("something went wrong")

	buf := bytes.NewBufferString("")
	log.SetOutput(buf)

	mc.On("WriteJSON", Anything, Anything).Return(writeErr)

	defaultErrorHandler(mc, err)
	logs, err := ioutil.ReadAll(buf)
	require.NoError(t, err, "reading log buffer")
	assert.Contains(t, string(logs), writeErr.Error())
}

func TestDefaultErrorHandlerShouldMakeNonHTTPErrorsIntoHTTPErrors(t *testing.T) {
	mc := &MockContext{}
	err := errors.New("something went wrong")
	mc.On("WriteJSON", Anything, Anything).Return(nil).Run(func(args Arguments) {
		herr := args.Get(1)
		assert.NotNil(t, herr)
		_, ok := herr.(HTTPError)
		assert.True(t, ok)
	})

	defaultErrorHandler(mc, err)
}

func TestMakeHandlerShouldCallErrorHandlerWhenNilHandler(t *testing.T) {
	var called bool

	r := NewRouter()
	r.SetErrorHandler(func(c Context, err error) {
		called = true
	})

	hndlr := r.makeHandler("GET", "/", func(Context) (Handler, error) {
		return nil, nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	hndlr(w, req, nil)
	assert.True(t, called)
}

func TestMakeHandlerShouldCallErrorHandlerWhenErrorOnCreateHandler(t *testing.T) {
	var called bool

	r := NewRouter()
	hErr := errors.New("")

	r.SetErrorHandler(func(c Context, err error) {
		called = true
		assert.Equal(t, hErr, err)
	})

	hndlr := r.makeHandler("GET", "/", func(Context) (Handler, error) {
		return nil, hErr
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	hndlr(w, req, nil)
	assert.True(t, called)
}

type badQueryHandler struct {
	Query int
}

func (h *badQueryHandler) Handle(Context) error { return nil }

func TestMakeHandlerShouldCallErrorHandlerWhenSetQueryFails(t *testing.T) {
	var called bool

	r := NewRouter()

	r.SetErrorHandler(func(c Context, err error) {
		called = true
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Query")
	})

	hndlr := r.makeHandler("GET", "/", func(Context) (Handler, error) {
		return &badQueryHandler{}, nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	hndlr(w, req, nil)
	assert.True(t, called)
}

type badURLParamsHandler struct {
	URLParams int
}

func (h *badURLParamsHandler) Handle(Context) error { return nil }

func TestMakeHandlerShouldCallErrorHandlerWhenSetURLParamsFails(t *testing.T) {
	var called bool

	r := NewRouter()

	r.SetErrorHandler(func(c Context, err error) {
		called = true
		require.Error(t, err)
		assert.Contains(t, err.Error(), "URL")
	})

	hndlr := r.makeHandler("GET", "/", func(Context) (Handler, error) {
		return &badURLParamsHandler{}, nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	hndlr(w, req, nil)
	assert.True(t, called)
}

type bodyHandler struct {
	handle HandlerFunc
	Body   struct {
		Age int
	}
}

func (h *bodyHandler) Handle(c Context) error { return h.handle(c) }

func TestMakeHandlerShouldNotSetBodyWhenContentLengthIsEmpty(t *testing.T) {
	var called bool

	r := NewRouter()

	handler := &bodyHandler{}
	handler.handle = func(Context) error {
		called = true
		return nil
	}

	hndlr := r.makeHandler("POST", "/", func(Context) (Handler, error) {
		return handler, nil
	})

	// the only way to set content-length is to use a raw request
	rawReq := `POST /post HTTP/1.1
Connection: close
Accept: */*
Content-Type: application/json
Content-Length: 0

{ "Age": 1 }`
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewBufferString(rawReq)))
	require.NoError(t, err)
	assert.Equal(t, "0", req.Header.Get("content-length"))

	w := httptest.NewRecorder()
	hndlr(w, req, nil)

	require.True(t, called)
	// even though Age was in the body, the request was not pared because no content-length
	// this is not a real-world scenario because the content-length is set by frameworks
	// and therefore will _always_ be set when there is a body. However, this allows
	// me to test the code which only triggers on ContentLength
	assert.NotEqual(t, 1, handler.Body.Age)
}

func TestMakeHandlerShouldSetBodyWhenContentLengthIsNotZero(t *testing.T) {
	r := NewRouter()

	handler := &bodyHandler{}
	handler.handle = func(Context) error {
		return nil
	}

	hndlr := r.makeHandler("POST", "/", func(Context) (Handler, error) {
		return handler, nil
	})

	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(`{ "Age": 1 }`))
	require.NoError(t, err)

	w := httptest.NewRecorder()
	hndlr(w, req, nil)

	assert.Equal(t, 1, handler.Body.Age)
}

type nopHandler struct{}

func (*nopHandler) Handle(Context) error {
	return nil
}

func TestShouldExecuteMiddlewaresInExactOrder(t *testing.T) {
	items := make([]string, 0)

	r := NewRouter()

	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			items = append(items, "a")
			return next(c)
		}
	})
	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			items = append(items, "b")
			return next(c)
		}
	})

	hndlr := r.makeHandler("GET", "/", func(Context) (Handler, error) {
		return &nopHandler{}, nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	hndlr(w, req, nil)
	assert.Len(t, items, 2)
	assert.Equal(t, items[0], "a")
	assert.Equal(t, items[1], "b")
}
