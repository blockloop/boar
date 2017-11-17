package boar

import (
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
	req.Header.Set("content-type", contentTypeJSON)

	w := httptest.NewRecorder()
	hndlr(w, req, nil)

	assert.Equal(t, 1, handler.Body.Age)
}

type badBodyHandler struct {
	Body int
}

func (*badBodyHandler) Handle(Context) error {
	return nil
}

func TestMakeHandlerShouldCallErrorHandlerWhenCantParseBody(t *testing.T) {
	var called bool

	r := NewRouter()
	r.SetErrorHandler(func(c Context, err error) {
		called = true
	})

	hndlr := r.makeHandler("POST", "/", func(Context) (Handler, error) {
		return &badBodyHandler{}, nil
	})

	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`a`))
	req.Header.Set("content-type", contentTypeJSON)
	req.Header.Set("content-length", "1")
	w := httptest.NewRecorder()
	hndlr(w, req, nil)
	assert.True(t, called)
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

func TestUseShouldPanicIfNilMiddleware(t *testing.T) {
	r := NewRouter()
	assert.Panics(t, func() {
		r.Use(nil)
	})
}

func TestUseShouldNotAddEmptyMiddlewares(t *testing.T) {
	r := NewRouter()
	r.Use(make([]Middleware, 0)...)
	assert.Len(t, r.mw, 0)
}

func TestRealRouterReturnsUnderlyingRouter(t *testing.T) {
	r := NewRouter()
	assert.NotNil(t, r.RealRouter())
}

func TestShouldCreateMethodHandlers(t *testing.T) {
	r := NewRouter()

	items := map[string]func(string, HandlerProviderFunc){
		http.MethodGet:     r.Get,
		http.MethodDelete:  r.Delete,
		http.MethodHead:    r.Head,
		http.MethodOptions: r.Options,
		http.MethodPatch:   r.Patch,
		http.MethodPost:    r.Post,
		http.MethodPut:     r.Put,
		http.MethodTrace:   r.Trace,
	}

	for method, handle := range items {
		// create a fake handler and assert that Handle was called when executing
		// a request with the provided method
		mh := &MockHandler{}

		mh.On("Handle", Anything).Run(func(args Arguments) {
			c := args.Get(0).(Context)
			assert.Equal(t, method, c.Request().Method)
		}).Return(nil)

		// call the router method Get,Delete,etc providing our mock handler
		handle("/", func(Context) (Handler, error) {
			return mh, nil
		})

		// startup a test server with our new router
		server := httptest.NewServer(r.RealRouter())
		defer server.Close()

		req, err := http.NewRequest(method, server.URL+"/", nil)
		require.NoError(t, err)

		_, err = http.DefaultClient.Do(req)
		require.NoError(t, err)

		mh.AssertCalled(t, "Handle", Anything)
	}
}

func TestShouldCallErrorHandlerWhenHandlerFails(t *testing.T) {
	var called bool
	r := NewRouter()
	herr := errors.New("asdf")

	r.SetErrorHandler(func(c Context, err error) {
		called = true
		assert.Equal(t, herr, err)
	})

	mh := &MockHandler{}
	mh.On("Handle", Anything).Return(herr)

	hndlr := r.makeHandler("GET", "/", func(Context) (Handler, error) {
		return mh, nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	hndlr(w, req, nil)
	assert.True(t, called)
	mh.AssertCalled(t, "Handle", Anything)
}

func TestSimpleHandlerShouldWork(t *testing.T) {
	r := NewRouter()

	handler := &MockHandler{}
	handler.On("Handle", Anything)

	r.MethodFunc("GET", "/", handler.Handle)

	server := httptest.NewServer(r.RealRouter())
	defer server.Close()
	http.Get(server.URL)
	handler.AssertCalled(t, "Handle", Anything)
}

func TestShouldValidatePostWhenEmptyBody(t *testing.T) {
	r := NewRouter()

	bh := &bodyHandler{}
	bh.handle = func(Context) error {
		assert.FailNow(t, "called Handle")
		return nil
	}

	r.Post("/", func(Context) (Handler, error) {
		return bh, nil
	})

	server := httptest.NewServer(r.RealRouter())
	defer server.Close()
	resp, err := http.Post(server.URL, contentTypeJSON, bytes.NewBufferString(""))
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestShouldErrorPostWhenNoContentType(t *testing.T) {
	r := NewRouter()

	bh := &bodyHandler{}
	bh.handle = func(Context) error {
		assert.FailNow(t, "called Handle")
		return nil
	}

	r.Post("/", func(Context) (Handler, error) {
		return bh, nil
	})

	server := httptest.NewServer(r.RealRouter())
	defer server.Close()
	resp, err := http.Post(server.URL, "", bytes.NewBufferString(""))
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestErrorsPostWhenEmptyBody(t *testing.T) {
	r := NewRouter()

	bh := &bodyHandler{}
	bh.handle = func(Context) error {
		assert.FailNow(t, "called Handle")
		return nil
	}

	r.Post("/", func(Context) (Handler, error) {
		return bh, nil
	})

	server := httptest.NewServer(r.RealRouter())
	defer server.Close()
	resp, err := http.Post(server.URL, contentTypeJSON, nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "EOF")
}
