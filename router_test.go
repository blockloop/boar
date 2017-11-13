package boar

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"testing"

	"github.com/blockloop/boar/mocks"
	"github.com/stretchr/testify/assert"
	. "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDefaultErrorHandlerShouldDoNothingToForNilError(t *testing.T) {
	mc := &mocks.Context{}
	defaultErrorHandler(mc, nil)
	mc.AssertNotCalled(t, "WriteJSON")
}

func TestDefaultErrorHandlerShouldWriteExistingHTTPError(t *testing.T) {
	mc := &mocks.Context{}
	status := 400
	err := NewHTTPErrorStatus(status)
	mc.On("WriteJSON", Anything, Anything).Return(nil).Run(func(args Arguments) {
		assert.Equal(t, status, args.Get(0))
		assert.Equal(t, err, args.Get(1))
	})

	defaultErrorHandler(mc, err)
}

func TestDefaultErrorHandlerShouldLogErrIfWriteJSONFails(t *testing.T) {
	mc := &mocks.Context{}
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
	mc := &mocks.Context{}
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

func TestMakeHandlerShouldCallErrorHandlerWhenSetURLParamsFaill(t *testing.T) {
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
