package boar

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// HandlerFunc is a function that handles an HTTP request
type HandlerFunc func(Context) error

// Router is an http router interface
type Router interface {
	HeadFunc(path string, h ...HandlerFunc)
	Head(path string, h Handler)

	TraceFunc(path string, h ...HandlerFunc)
	Trace(path string, h Handler)

	DeleteFunc(path string, h ...HandlerFunc)
	Delete(path string, h Handler)

	OptionsFunc(path string, h ...HandlerFunc)
	Options(path string, h Handler)

	GetFunc(path string, h ...HandlerFunc)
	Get(path string, h Handler)

	PutFunc(path string, h ...HandlerFunc)
	Put(path string, h Handler)

	PostFunc(path string, h ...HandlerFunc)
	Post(path string, h Handler)

	PatchFunc(path string, h ...HandlerFunc)
	Patch(path string, h Handler)

	ConnectFunc(path string, h ...HandlerFunc)
	Connect(path string, h Handler)

	MethodFunc(method string, path string, h ...HandlerFunc)
	Method(method string, path string, h Handler)

	ListenAndServe(addr string) error

	// RealRouter returns the httprouter.Router used for actual serving
	RealRouter() *httprouter.Router
}

// Handler is an http Handler
type Handler interface {
	Handle(Context) error
}

// BeforeHandler is a Handler that executes Middlewares BEFORE calling Handle
type BeforeHandler interface {
	Before() []HandlerFunc
}

// AfterHandler is a Handler that executes Middlewares AFTER calling Handle
type AfterHandler interface {
	After() []HandlerFunc
}

// BeforeAfterHandler is a Handler that executes Middlewares BEFORE and AFTER calling Handle
type BeforeAfterHandler interface {
	BeforeHandler
	Handler
	AfterHandler
}

type requestRouter struct {
	r *httprouter.Router
}

// NewRouter creates a new router for handling http requests
func NewRouter() Router {
	return newRouter()
}

func newRouter() *requestRouter {
	return &requestRouter{
		r: httprouter.New(),
	}
}

func (r *requestRouter) RealRouter() *httprouter.Router {
	return r.r
}

func (r *requestRouter) Method(method string, path string, h Handler) {
	r.RealRouter().Handle(method, path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c := newContext(r, w, ps)
		h.Handle(c)
	})
}

func (r *requestRouter) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, r.RealRouter())
}

func (r *requestRouter) MethodFunc(method string, path string, h ...HandlerFunc) {
	panic("not implemented")
}

func (r *requestRouter) HeadFunc(path string, h ...HandlerFunc) {
	panic("not implemented")
}

func (r *requestRouter) Head(path string, h Handler) {
	r.Method(http.MethodHead, path, h)
}

func (r *requestRouter) TraceFunc(path string, h ...HandlerFunc) {
	panic("not implemented")
}

func (r *requestRouter) Trace(path string, h Handler) {
	r.Method(http.MethodTrace, path, h)
}

func (r *requestRouter) DeleteFunc(path string, h ...HandlerFunc) {
	panic("not implemented")
}

func (r *requestRouter) Delete(path string, h Handler) {
	r.Method(http.MethodDelete, path, h)
}

func (r *requestRouter) OptionsFunc(path string, h ...HandlerFunc) {
	panic("not implemented")
}

func (r *requestRouter) Options(path string, h Handler) {
	r.Method(http.MethodOptions, path, h)
}

func (r *requestRouter) GetFunc(path string, h ...HandlerFunc) {
	panic("not implemented")
}

func (r *requestRouter) Get(path string, h Handler) {
	r.Method(http.MethodGet, path, h)
}

func (r *requestRouter) PutFunc(path string, h ...HandlerFunc) {
	panic("not implemented")
}

func (r *requestRouter) Put(path string, h Handler) {
	r.Method(http.MethodPut, path, h)
}

func (r *requestRouter) PostFunc(path string, h ...HandlerFunc) {
	panic("not implemented")
}

func (r *requestRouter) Post(path string, h Handler) {
	r.Method(http.MethodPost, path, h)
}

func (r *requestRouter) PatchFunc(path string, h ...HandlerFunc) {
	panic("not implemented")
}

func (r *requestRouter) Patch(path string, h Handler) {
	r.Method(http.MethodPatch, path, h)
}

func (r *requestRouter) ConnectFunc(path string, h ...HandlerFunc) {
	panic("not implemented")
}

func (r *requestRouter) Connect(path string, h Handler) {
	r.Method(http.MethodConnect, path, h)
}
