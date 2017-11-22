package boar

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	. "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDefaultErrorHandlerShouldDoNothingToForNilError(t *testing.T) {
	mc := &MockContext{}
	ErrorHandler(func(c Context) error {
		return nil
	})(mc)
	mc.AssertNotCalled(t, "WriteJSON")
}

func TestDefaultErrorHandlerWritesExistingHTTPErrorIfNotAlreadyWritten(t *testing.T) {
	mr := &MockResponseWriter{}
	mr.On("Len").Return(0)

	mc := &MockContext{}
	mc.On("Response").Return(mr)

	status := http.StatusBadRequest
	err := NewHTTPErrorStatus(status)
	mc.On("WriteJSON", Anything, Anything).Return(nil).Run(func(args Arguments) {
		assert.Equal(t, status, args.Get(0))
		assert.Equal(t, err, args.Get(1))
	})

	ErrorHandler(func(c Context) error {
		return err
	})(mc)
}

func TestDefaultErrorHandlerReturnsWriteErrIfWriteJSONFails(t *testing.T) {
	mc := &MockContext{}
	status := 400
	err := NewHTTPErrorStatus(status)
	writeErr := errors.New("something went wrong")

	mc.On("WriteJSON", Anything, Anything).Return(writeErr)

	mr := &MockResponseWriter{}
	mr.On("Flush").Return(nil)
	mr.On("Len").Return(0)
	mc.On("Response").Return(mr)

	actual := ErrorHandler(func(c Context) error {
		return err
	})(mc)

	assert.Contains(t, actual.Error(), writeErr.Error())
}

func TestDefaultErrorHandlerMakesNonHTTPErrorsIntoHTTPErrors(t *testing.T) {
	err := errors.New("something went wrong")
	mr := &MockResponseWriter{}
	mr.On("Len").Return(0)

	mc := &MockContext{}
	mc.On("Response").Return(mr)
	mc.On("WriteJSON", Anything, Anything).Return(nil)

	actual := ErrorHandler(func(c Context) error {
		return err
	})(mc)

	assert.IsType(t, &httpError{}, actual)
}

func TestDefaultErrorHandlerDoesNotWriteIfAlreadyWritten(t *testing.T) {
	mrw := &MockResponseWriter{}
	mrw.On("Len").Return(1)

	mc := &MockContext{}
	mc.On("Response").Return(mrw)
	mc.On("WriteJSON", Anything, Anything).Return(nil)

	ErrorHandler(func(c Context) error {
		return errors.New("hello, world")
	})(mc)

	mrw.AssertCalled(t, "Len")
	mc.AssertNotCalled(t, "WriteJSON", Anything)
}

func TestRequestParserMiddlewarePanicsWhenNilHandler(t *testing.T) {
	handle := requestParserMiddleware(func(Context) (Handler, error) {
		return nil, nil
	})

	assert.Panics(t, func() {
		handle(nil)
	})
}

func TestMakeHandlerReturnsErrorWhenErrorOnCreateHandler(t *testing.T) {
	err := errors.New("something broke")
	handle := requestParserMiddleware(func(Context) (Handler, error) {
		return nil, err
	})

	actual := handle(nil)
	assert.Equal(t, err, actual)
}

type badQueryHandler struct {
	Query int
}

func (h *badQueryHandler) Handle(Context) error { return nil }

func TestRequestParserMiddlewareReturnsErrorWhenSetQueryFails(t *testing.T) {
	handle := requestParserMiddleware(func(Context) (Handler, error) {
		return &badQueryHandler{}, nil
	})

	req := httptest.NewRequest("GET", "/?hello=world", nil)

	mc := &MockContext{}
	mc.On("Request").Return(req)

	err := handle(mc)
	assert.Error(t, err)
}

type badURLParamsHandler struct {
	URLParams int
}

func (h *badURLParamsHandler) Handle(Context) error { return nil }

func TestRequestParserMiddlewareReturnsErrorWhenSetURLParamsFails(t *testing.T) {
	handle := requestParserMiddleware(func(Context) (Handler, error) {
		return &badURLParamsHandler{}, nil
	})

	req := httptest.NewRequest("GET", "/", nil)

	mc := &MockContext{}
	mc.On("Request").Return(req)
	mc.On("URLParams").Return(httprouter.Params{})

	err := handle(mc)
	assert.Error(t, err)
}

type bodyHandler struct {
	handle HandlerFunc
	Body   struct {
		Age int `valid:"required"`
	}
}

func (h *bodyHandler) Handle(c Context) error { return h.handle(c) }

type badBodyHandler struct {
	Body int
}

func (*badBodyHandler) Handle(Context) error {
	return nil
}

func TestRequestParserMiddlewareReturnsErrorWhenSetBodyFails(t *testing.T) {
	handle := requestParserMiddleware(func(Context) (Handler, error) {
		return &badBodyHandler{}, nil
	})

	req := httptest.NewRequest("POST", "/", bytes.NewBufferString("{}"))
	req.Header.Set("content-type", contentTypeJSON)

	mc := &MockContext{}
	mc.On("Request").Return(req)
	mc.On("URLParams").Return(httprouter.Params{})

	err := handle(mc)
	assert.Error(t, err)
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
			err := next(c)
			items = append(items, "first")
			return err
		}
	})
	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			err := next(c)
			items = append(items, "second")
			return err
		}
	})

	mh := &MockHandler{}
	mh.On("Handle", Anything).Run(func(args Arguments) {
		args.Get(0).(Context).WriteStatus(http.StatusOK)
	}).Return(nil)

	r.Get("/", func(Context) (Handler, error) {
		return mh, nil
	})

	server := httptest.NewServer(r)
	defer server.Close()

	_, err := http.Get(server.URL)
	require.NoError(t, err)

	assert.Equal(t, []string{"first", "second"}, items)
}

func TestUseShouldPanicIfNilMiddleware(t *testing.T) {
	r := NewRouter()
	assert.Panics(t, func() {
		r.Use(nil)
	})
}

func TestUseShouldNotAddNilMiddlewares(t *testing.T) {
	r := NewRouter()
	start := len(r.mw)
	r.Use(make([]Middleware, 0)...)
	assert.Len(t, r.mw, start)
}

func TestRealRouterReturnsUnderlyingRouter(t *testing.T) {
	r := NewRouter()
	assert.NotNil(t, r)
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
		server := httptest.NewServer(r)
		defer server.Close()

		req, err := http.NewRequest(method, server.URL+"/", nil)
		require.NoError(t, err)

		_, err = http.DefaultClient.Do(req)
		require.NoError(t, err)

		mh.AssertCalled(t, "Handle", Anything)
	}
}

// func TestShouldCallErrorHandlerWhenHandlerFails(t *testing.T) {
// 	var called bool
// 	r := NewRouter()
// 	herr := errors.New("asdf")

// 	r.SetErrorHandler(func(c Context, err error) {
// 		called = true
// 		assert.Equal(t, herr, err)
// 	})

// 	mh := &MockHandler{}
// 	mh.On("Handle", Anything).Return(herr)

// 	hndlr := r.makeHandler("GET", "/", func(Context) (Handler, error) {
// 		return mh, nil
// 	})

// 	req := httptest.NewRequest("GET", "/", nil)
// 	w := httptest.NewRecorder()
// 	hndlr(w, req, nil)
// 	assert.True(t, called)
// 	mh.AssertCalled(t, "Handle", Anything)
// }

func TestSimpleHandlerShouldWork(t *testing.T) {
	r := NewRouter()

	handler := &MockHandler{}
	handler.On("Handle", Anything).Return(nil)

	r.MethodFunc("GET", "/", handler.Handle)

	server := httptest.NewServer(r)
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

	server := httptest.NewServer(r)
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

	server := httptest.NewServer(r)
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

	server := httptest.NewServer(r)
	defer server.Close()
	resp, err := http.Post(server.URL, contentTypeJSON, nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "EOF")
}

func TestCallsErrHandlerAsFirstMiddleware(t *testing.T) {
	var errHandlerCalled bool
	ErrorHandler = func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			errHandlerCalled = true
			return next(c)
		}
	}

	r := NewRouter()

	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			err := next(c)
			require.True(t, errHandlerCalled)
			return err
		}
	})

	mh := &MockHandler{}
	mh.On("Handle", Anything).Return(nil)

	r.Get("/", func(Context) (Handler, error) {
		return mh, nil
	})

	server := httptest.NewServer(r)
	defer server.Close()

	_, err := http.Get(server.URL + "/")
	require.NoError(t, err)
}
