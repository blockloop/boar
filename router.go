package boar

// HandlerFunc is a function that handles an HTTP request
type HandlerFunc func(Context) error

// Router is an http router interface
type Router interface {
	HeadFunc(path string, h HandlerFunc, mw ...HandlerFunc)
	Head(path string, h Handler)

	TraceFunc(path string, h HandlerFunc, mw ...HandlerFunc)
	Trace(path string, h Handler)

	Delete(path string, h HandlerFunc, mw ...HandlerFunc)
	DeleteFunc(path string, h Handler)

	OptionsFunc(path string, h HandlerFunc, mw ...HandlerFunc)
	Options(path string, h Handler)

	GetFunc(path string, h HandlerFunc, mw ...HandlerFunc)
	Get(path string, h Handler)

	PutFunc(path string, h HandlerFunc, mw ...HandlerFunc)
	Put(path string, h Handler)

	PostFunc(path string, h HandlerFunc, mw ...HandlerFunc)
	Post(path string, h Handler)

	PatchFunc(path string, h HandlerFunc, mw ...HandlerFunc)
	Patch(path string, h Handler)

	ConnectFunc(path string, h HandlerFunc, mw ...HandlerFunc)
	Connect(path string, h Handler)

	MethodFunc(method string, path string, h HandlerFunc, mw ...HandlerFunc)
	Method(method string, path string, h Handler)
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
