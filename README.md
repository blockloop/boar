# Boar

[![GoDoc](https://godoc.org/github.com/blockloop/boar?status.svg)](https://godoc.org/github.com/blockloop/boar)
[![Build Status](https://travis-ci.org/blockloop/boar.svg?branch=master)](https://travis-ci.org/blockloop/boar)
[![Coverage Status](https://coveralls.io/repos/github/blockloop/boar/badge.svg?branch=master&t=872674523461)](https://coveralls.io/github/blockloop/boar?branch=master)


Boar aims to streamline the design of HTTP server applications by providing separation of concerns usually crammed into HTTP handler functions or chained middlewares. Some of the most common actions are performed automatically before the HTTP handler is even executed.

## Handlers

With Boar, handlers are not functions. A Handler in Boar is an interface that has `Handle(Context) error`. This provides the ability to build request context around the handler _before_ the handler is executed. An HTTP handler looks something like this:

```go
type createPersonHandler struct {
    db store.Persons
    log log.Interface

    Body struct {
        Name string `query:"name"`
        Age  int    `query:"age"`
    }
}

func (h *listTagsHandler) Handle(boar.Context) error {
    // h.log.Info("received request")
    // id, err := h.db.Create(h.Body.Name, h.Body.Age)
}
```

The Query field within the handler will be automaticaly populated with the contents of the query string (if they exist). Not only does this provide an obvious structure to the query string, it provides static type parsing by default. Rather than littering your handler func with `r.URL.Query().Get("page")`, `strconv.ParseInt()`, and writing bad request errors, the framework handles these operations automatically. If the client sends `?page=aaa` then parsing will fail and the framework will automatically respond to the client with an error explaining that `page` must be an integer. listTagsHandler.Handle() is never executed.

Validation is also provided. See [Validation](#Validation) for more details.

## Body parsing



## Error Handling

Each HTTP handler returns an `error` for all paths except for happy path. If validation or parsing fail then the handler returns an error and a global handler writes the response back to the client. This keeps Handlers clean and straight forward and provides separation between validation, parsing, processing, response writing, and error handling. 

- HTTPError - interface. A special error that is easily formatted and understood in an HTTP response. It consists of an HTTP status code and an underlying error. 
- ValidationError - an HTTPError that was caused by validation. It is an HTTPError with a status code of 400 - Bad Request. It is useful for differentiating between validation errors and internal errors (client or server caused)

## Example
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

