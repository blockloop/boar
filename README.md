# Boar

[![GoDoc](https://godoc.org/github.com/blockloop/boar?status.svg)](https://godoc.org/github.com/blockloop/boar)
[![Build Status](https://travis-ci.org/blockloop/boar.svg?branch=master)](https://travis-ci.org/blockloop/boar)
[![Coverage Status](https://coveralls.io/repos/github/blockloop/boar/badge.svg?branch=master&t=872674523461)](https://coveralls.io/github/blockloop/boar?branch=master)


Boar aims to streamline the design of HTTP server applications by providing separation of concerns usually crammed into HTTP handler functions or chained middlewares. Some of the most common actions are performed automatically before the HTTP handler is even executed.

## Thanks

Boar is only possible because of the many frameworks and ideas that have inspired me. I want to specifically thank [mustafakin/gongular](https://github.com/mustafaakin/gongular) for the core parsing idea.

Other thanks to

- [Gin](https://github.com/gin-gonic/gin)
- [govalidator](https://github.com/asaskevich/govalidator)

## Handlers

With Boar, handlers are not functions. A Handler in Boar is an interface that has `Handle(Context) error`. This provides the ability to build request context around the handler _before_ the handler is executed. An HTTP handler looks something like this:

```go
type createPersonHandler struct {
    db store.Persons
    log log.Interface

    // Parse the JSON body automatically
    Body struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }
}

func (h *listTagsHandler) Handle(boar.Context) error {
    // h.log.Info("received request")
    // id, err := h.db.Create(h.Body.Name, h.Body.Age)
}
```

## Query String

Query strings are parsed by adding a Query field to your handler struct. The Query field within the handler will be automaticaly populated with the contents of the query string (if they exist). Not only does this provide an obvious structure to the query string, it provides static type parsing by default. Rather than littering your handler func with `r.URL.Query().Get("page")`, `strconv.ParseInt()`, and writing bad request errors, the framework handles these operations automatically. If the client sends `?page=aaa` then parsing will fail and the framework will automatically respond to the client with an error explaining that `page` must be an integer. listTagsHandler.Handle() is never executed.

Validation is also provided. See [Validation](#Validation) for more details.


## Error Handling

_All_ error handling is left to the global error handler. This includes validation errors, internal errors, etc. This provides a clean working space for handlers and a single-responsibility handler for errors. The default error handler logs to the golang log with raw errors and is not recommended for use in production environments.

Each Handler returns `error`. If validation or parsing fail then the handler returns an error and a global handler writes the response back to the client. This keeps Handlers clean and straight forward and provides separation between validation, parsing, processing, response writing, and error handling. 

[HTTPError](https://godoc.org/github.com/blockloop/boar#HTTPError) is used to automatically format responses back to the client. It provides a status code an a cause.

 ValidationError - an HTTPError that was caused by validation. It is an HTTPError with a status code of 400 - Bad Request. It prints errors as JSON for the client to fix their request. It is useful for differentiating between validation errors and internal errors (client or server caused). A ValidationError is formatted as follows

```json
{
    "errors": {
        "query": [
            "limit must be an int"
        ],
        "body": [
            "age must be between 18 and 150"
        ]
    }
}
```

There are a few helpful errors pre-defined within boar that will help with middleware.


## Middleware

Middleware behaves just like middleware with any other framework. Once you create a router you execute `(*Router).Use(Middleware)`. When requests are received the middleware will be executed in the order which they were added. However, because of the [custom response writer](response_writer.go) access to all writer methods are available *even after the primary handler has written it's response*. This is particularly useful where middlewares add new headers to responses, log/check the status code of the response, etc. The response message is not written to the client until all middlewares and the primary handler have been executed. At this point the buffer is flushed to the net/http.ResponseWriter. Middleware adheres to the following signature:

```go
type HandlerFunc func(Context) error

func(next HandlerFunc) HandlerFunc {
    // ...
}
```

## Recommended Project Layout

```
.
├── main.go                // main should create handler factory, and register routes to boar
├── auth/                  // other packages which handlers use such as authentication
│   └── user.go
├── clients/
│   └── external_client_impl.go
├── handlers/              // handlers should have their own package and each handler should
│   │                      // be private. Only the factory should expose funcs which build handlers
│   │
│   ├── user_factory.go    // a factory which creates user handlers for execution
│   │                      // it is recommended to use one factory per area of interest
│   ├── get_user.go        // each handler should have it's own file
│   └── set_user.go
└── store/
    └── location.go        // you should implement an interface for every database
```

## Example Project

```go
// main.go
func main() {
    // ... setup constant variables
    factory := handlers.NewFactory(db, appLogger)
    r := boar.NewRouter()
    r.Get("/person/:id", factory.GetPersonByID)
    log.WithError(r.ListenAndServe(":3000")).Fatal("application exit")
}

// handlers/factory.go
type Factory struct {
    db store.Store
    log log.Interface
}
func NewFactory(db store.Store) Factory {
    return &Factory{
        db: db,
        log: log,
    }
}
func (f *Factory) GetPersonByID(c boar.Context) (Handler, error) {
    return &getUser{
        db: f.db,
        log: f.log.WithFields(log.F{
            "request.id": f.RequestID(c),
            "user.id": f.UserID(c),
        }),
    }, nil
}

// handlers/get_user.go
type getUser struct {
    db store.Store
    log log.Interface
    URLParams struct {
        ID int `url:"id"`
    }
}
func (h *getUser) Handle(c boar.Context) error {
    // at this point all validation, authentication, and other setup processes
    // have completed successfully
    user, err := db.FindUserByID(h.URLParams.ID)
    if err != nil {
        // logs request.id and user.id
        log.WithError(err).Error("error listing items from datastore")
        // return the error (as-is) to be handled by the global error handler
        // which will write a 500 internal server error by default.
        return err
    }

    if user == nil {
        // shortcut to a 404 error message with a message to signify that the route
        // was okay, but the entity was not found
        return boar.ErrEntityNotFound
    }

    // if WriteJSON fails for some reason let the global error handler handle
    // the logging
    return c.WriteJSON(http.StatusOK, &Response{
        User: user,
    })
}
```

