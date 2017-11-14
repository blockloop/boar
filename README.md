# Boar

[![GoDoc](https://godoc.org/github.com/blockloop/boar?status.svg)](https://godoc.org/github.com/blockloop/boar)
[![Build Status](https://travis-ci.org/blockloop/boar.svg?branch=master)](https://travis-ci.org/blockloop/boar)
[![Coverage Status](https://coveralls.io/repos/github/blockloop/boar/badge.svg?branch=master&t=872674523461)](https://coveralls.io/github/blockloop/boar?branch=master)


Boar aims to streamline the design of HTTP server applications by providing separation of concerns usually crammed into HTTP handler functions or chained middlewares. Some of the most common actions are performed automatically before the HTTP handler is even executed.

## Thanks

Boar is only possible because of the many frameworks and ideas that have inspired me. I want to specifically thank [mustafakin/gongular](https://github.com/mustafaakin/gongular) for the core parsing concepts.

Other thanks to

- [govalidator](https://github.com/asaskevich/govalidator) (used in Boar)
- [httprouter](https://github.com/julienschmidt/httprouter) (used in Boar)
- [Gin](https://github.com/gin-gonic/gin) (inspiration)

## Usage

See the wiki for usage and [Getting Started](http://github.com/blockloop/boar/wiki/Getting-Started)

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

