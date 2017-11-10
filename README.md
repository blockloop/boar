# Boar
Boar is a tiny HTTP framework (<1k LOC) that provides helpers to functionality commonly repeated in HTTP handlers: 

1. Parsing a query string to a struct with static types and validation
2. Parsing URL Path parameters to a struct with static types and validation
3. Parsing JSON requests to a struct with static types and validation

Each HTTP handler returns an `error` for all paths except for happy path. If validation or parsing fail then the handler returns an error and a global handler writes the response back to the client. This keeps Handlers clean and straight forward and provides separation between validation, parsing, processing, response writing, and error handling. 

## Factory
A factory is given contextual information and creates handlers. It solves several problems. Often HTTP handlers consist of constants and variable dependencies. Constant dependencies are things like configuration, databases, and http/gRPC clients. They remain the same throughout the lifecycle of the application. Variable dependencies are things that are built based on the request context. Variable dependencies consist of things like, request object (parsed as JSON, etc), query strings, loggers with context, authentication information, etc. A factory is responsible for filling out these pieces of information before the handler is activated. This is achieved by creating a Handler struct. 

## Handlers
A Handler is a struct that has  `func Handle(boar.Context) error`. When an HTTP request is received it is passed to the Factory to generate a handler. The Handler is then executed with the `boar.Context` for processing. Handlers are generally limited in scope. They are provided _all_ of the required components to process the request (i.e. user/auth information, datastores, loggers, etc)

## Context
`boar.Context` is provided for http handler or middleware. While it is a useful wrapper, it provides full control over the request and response objects with `c.Request() *http.Request` and `c.Response() http.ResponseWriter`. Along with WriteJSON and a few other small helpers the following are what I consider most useful

1. `ReadQuery(interface{}) error`
2. `ReadURLParams(interface{}) error`
3. `ReadJSON(interface{}) error`

Each of these funcs parse the query string, URL path parameters, and Body as JSON, respectively. Along with providing static type classing it provides validation provided by [asaskevich/govalidator](https://github.com/asaskevich/govalidator). It is a comprehensive validation engine for validating struct fields. It is not required by any means, but it is free if the tags are present. Each of these funcs return a `boar.HTTPError` if validation fails by either invalid types or govalidator failures. 

## Paths and Routes
All paths are routes are entirely handled by [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter). Boar uses a small wrapper to create `boar.Context` and handle errors. 

## HTTPError
HTTPError is a special error that is easily formatted and understood in an HTTP response. It consists of an HTTP status code and an underlying error. 

## ValidationError
ValidationError is an HTTPError that was caused by validation. It consists of nothing more than an embedded HTTPError with a status code of 400 - Bad Request. It is useful for differentiating between validation errors and internal errors (client or server caused)

## Example
```go
// main.go
func main() {
    factory := handlers.NewFactory(db, appLogger)
    r := boar.NewRouter()
    r.Get("/person/:id", factory.ItemGetHandler)
    panic(r.ListenAndServe(":3000"))
}

// handlers/factory.go
type Factory struct {
    db store.Store
    log log.Interface
}
func NewFactory(db *sql.DB) Factory {
    return &Factory{
        db: store.New(db),
        log: log,
    }
}
func (f *Factory) ItemGetHandler(c boar.Context) (Handler, error) {
    h := &getUser{
        db: f.db,
        log: f.log.WithFields(log.F{
            "request.id": f.RequestID(c),
            "user.id": f.UserID(c),
        }),
    }
    err := c.ParseURLParams(&h.URLParams)
    if err != nil {
        // err is a ValidationError and will be seralized as JSON and sent
        // to the client as a 400 status code
        return nil, err
    }
    return h, nil
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
    user, err := db.FindUserByID(h.URLParams.ID)
    if err != nil {
        log.WithError(err).Error("error listing items from datastore")
        // logs request.id and user.id
        return err
    }

    if user == nil {
        return boar.NewHTTPError(http.StatusNotFound, fmt.Errorf("user not found"))
        // err will be seralized as JSON and sent to the client as a 404 with a body
        // of { "error": "user not found" }
    }

    c.WriteJSON(http.StatusOK, map[string]interface{}{
        "user": user,
    })
}
```

