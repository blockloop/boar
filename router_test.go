package boar

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	. "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDefaultErrorHandlerShouldDoNothingToForNilError(t *testing.T) {
	mc := &MockContext{}
	defaultErrorHandler(func(c Context) error {
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

	defaultErrorHandler(func(c Context) error {
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

	actual := defaultErrorHandler(func(c Context) error {
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

	actual := defaultErrorHandler(func(c Context) error {
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

	defaultErrorHandler(func(c Context) error {
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

type urlParamsHandler struct {
	handle    HandlerFunc
	URLParams struct {
		Age int `valid:"required"`
	}
}

func (h *urlParamsHandler) Handle(c Context) error { return h.handle(c) }

func TestRequestParserMiddlewareReturns404WhenSetURLParamsFailsValidation(t *testing.T) {
	r := NewRouter()
	r.Get("/users/:id", func(Context) (Handler, error) {
		return &urlParamsHandler{}, nil
	})

	req := httptest.NewRequest("GET", "/users/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	rec.Flush()

	assert.Equal(t, http.StatusNotFound, rec.Result().StatusCode)
}

func TestRequestParserMiddlewareDoesNotPrintErrorWhenValidationError(t *testing.T) {
	r := NewRouter()
	r.Get("/users/:id", func(Context) (Handler, error) {
		return &urlParamsHandler{}, nil
	})

	req := httptest.NewRequest("GET", "/users/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	rec.Flush()
	resp := rec.Result()

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	body := strings.ToLower(string(b))

	assert.NotContains(t, body, "age")
	assert.Contains(t, body, "not found")
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

func TestShouldExecuteMiddlewaresInReverseOrder(t *testing.T) {
	// Reverse order means they will essentially execute in sequential order because
	// each middleware executes the *next* item from within itself
	items := make([]string, 0, 3)

	r := NewRouter()

	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			items = append(items, "third")
			return next(c)
		}
	}, func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			items = append(items, "second")
			return next(c)
		}
	})

	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			items = append(items, "first")
			return next(c)
		}
	})

	r.MethodFunc(http.MethodGet, "/", func(c Context) error {
		return c.WriteStatus(200)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	assert.Equal(t, []string{"first", "second", "third"}, items)
}

func TestUseShouldPanicIfNilMiddleware(t *testing.T) {
	r := NewRouter()
	assert.Panics(t, func() {
		r.Use(nil)
	})
}

func TestUseShouldNotAddNilMiddlewares(t *testing.T) {
	r := NewRouter()
	start := len(r.middlewares)
	r.Use(make([]Middleware, 0)...)
	assert.Len(t, r.middlewares, start)
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

		req := httptest.NewRequest(method, "/", nil)

		r.ServeHTTP(httptest.NewRecorder(), req)

		mh.AssertCalled(t, "Handle", Anything)
	}
}

func TestSimpleHandlerShouldWork(t *testing.T) {
	r := NewRouter()

	handler := &MockHandler{}
	handler.On("Handle", Anything).Return(nil)

	r.MethodFunc(http.MethodGet, "/", handler.Handle)

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
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

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(""))
	req.Header.Set("content-type", contentTypeJSON)

	r.ServeHTTP(rec, req)

	rec.Flush()
	resp := rec.Result()

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

	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(""))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp := rec.Result()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestErrorHandlerSetsStatusWhenHandlerErrors(t *testing.T) {
	r := NewRouter()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	c := NewContext(req, w, nil)
	h := r.withMiddlewares(func(Context) error {
		return ErrUnauthorized
	})
	h(c)

	assert.Equal(t, http.StatusUnauthorized, c.Response().Status())
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

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("content-type", contentTypeJSON)

	r.ServeHTTP(rec, req)

	resp := rec.Result()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "EOF")
}

func TestHandlesErrorsBetweenMiddlewares(t *testing.T) {
	r := NewRouter()
	herr := errors.New("asdf")

	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			err := next(c)
			assert.IsType(t, (*httpError)(nil), err)
			assert.Contains(t, err.Error(), herr.Error())
			return herr
		}
	}, func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			err := next(c)
			assert.IsType(t, (*httpError)(nil), err)
			assert.Contains(t, err.Error(), herr.Error())
			return herr
		}
	})

	r.MethodFunc("POST", "/", func(Context) error {
		return herr
	})

	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
}

func TestHandleWritesErrorsBeforeMiddleware(t *testing.T) {
	called := make(chan string, 5)

	r := NewRouter()
	r.ErrorHandler = func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			err := next(c)
			called <- "ErrorHandler"
			return err
		}
	}

	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			err := next(c)
			called <- "Middleware"
			// require.Equal(t, err, ErrForbidden)
			// require.Equal(t, http.StatusForbidden, c.Response().Status())
			return err
		}
	})

	mh := &MockHandler{}
	mh.On("Handle", Anything).Run(func(Arguments) {
		called <- "Handler"
	}).Return(ErrForbidden)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	c := NewContext(req, w, nil)

	wrapped := r.withMiddlewares(mh.Handle)
	err := wrapped(c)
	w.Flush()
	require.IsType(t, &httpError{}, err)

	close(called)

	assert.Equal(t, "Handler", <-called)
	assert.Equal(t, "ErrorHandler", <-called)
	assert.Equal(t, "Middleware", <-called)
}

func TestPanicHandlerSets500StatusCode(t *testing.T) {
	r := NewRouter()
	r.Use(PanicMiddleware)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	err := fmt.Errorf("something broke")

	r.MethodFunc(http.MethodGet, "/", func(Context) error {
		panic(err)
	})

	r.ServeHTTP(rec, req)
	rec.Flush()

	resp := rec.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestPanicHandlerPreservesPanicMessage(t *testing.T) {
	r := NewRouter()
	r.Use(PanicMiddleware)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	err := fmt.Errorf("something broke")

	r.MethodFunc(http.MethodGet, "/", func(Context) error {
		panic(err)
	})

	r.ServeHTTP(rec, req)
	rec.Flush()

	resp := rec.Result()

	body, rerr := ioutil.ReadAll(resp.Body)
	require.NoError(t, rerr)

	assert.Contains(t, string(body), err.Error())
}

func TestPanicHandlerPreservesErrorWhenNoPanic(t *testing.T) {
	r := NewRouter()
	r.Use(PanicMiddleware)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	r.MethodFunc(http.MethodGet, "/", func(Context) error {
		return ErrForbidden
	})

	r.ServeHTTP(rec, req)
	rec.Flush()

	resp := rec.Result()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestPanicHandlerConvertsPanicStringsToHTTPError(t *testing.T) {
	r := NewRouter()
	r.Use(PanicMiddleware)

	done := &sync.WaitGroup{}
	done.Add(1)

	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			defer done.Done()
			err := next(c)
			assert.Implements(t, (*HTTPError)(nil), err)
			assert.Contains(t, err.Error(), "something broke")
			return err
		}
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	r.MethodFunc(http.MethodGet, "/", func(Context) error {
		panic("something broke")
	})

	r.ServeHTTP(rec, req)
	done.Wait()
	rec.Flush()
}

func TestNotFoundHandlerDoesNotPrintBody(t *testing.T) {
	r := NewRouter()
	r.Use(PanicMiddleware)

	done := &sync.WaitGroup{}
	done.Add(1)

	r.MethodFunc(http.MethodGet, "/hello", func(Context) error {
		return nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	r.ServeHTTP(rec, req)
	rec.Flush()
	resp := rec.Result()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Empty(t, body)
}

func TestMethodNodAllowedHandlerDoesNotPrintBody(t *testing.T) {
	r := NewRouter()
	r.Use(PanicMiddleware)

	done := &sync.WaitGroup{}
	done.Add(1)

	r.MethodFunc(http.MethodGet, "/", func(Context) error {
		return nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	r.ServeHTTP(rec, req)
	rec.Flush()
	resp := rec.Result()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Empty(t, body)
}
