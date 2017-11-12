# Boar

[![Build Status](https://travis-ci.org/blockloop/boar.svg?branch=master)](https://travis-ci.org/blockloop/boar)
[![Coverage Status](https://coveralls.io/repos/github/blockloop/boar/badge.svg?branch=master)](https://coveralls.io/github/blockloop/boar?branch=master)

Boar is a small HTTP framework that aims to simplify and streamline the design of HTTP server applications. Boar automatically performs actions which are usually repeated across all handlers. Boar Automatically parses queryString, url parameters, and request body to a struct with static types and validation.

Each HTTP handler returns an `error` for all paths except for happy path. If validation or parsing fail then the handler returns an error and a global handler writes the response back to the client. This keeps Handlers clean and straight forward and provides separation between validation, parsing, processing, response writing, and error handling. 

## HTTPError

HTTPError is a special error that is easily formatted and understood in an HTTP response. It consists of an HTTP status code and an underlying error. 

## ValidationError

ValidationError is an HTTPError that was caused by validation. It consists of nothing more than an embedded HTTPError with a status code of 400 - Bad Request. It is useful for differentiating between validation errors and internal errors (client or server caused)

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

