# checkpoint

### Description
Checkpoint is package meant to assist in running integration testing of REST APIs handlers. It provides a way to construct an HTTP requests along including middleware and evaluate the resulting response.

### Installation
```go get github.com/rkuprov/checkpoint@latest```

### How to use
Because url query and path parameters tend to be implemented differently by different routers, the Check function needs to be initialized with the router you're intending to use. For this reason a Router interface was introduced that implements two methods:
* ServerHTTP(w http.ResponseWriter, r *http.Request)
* Handle(pattern string, handler *http.Handler)

Fulfilling this interface allows parsing of query and path parameters.

Some routers, such as one provided by `github.com/gorilla/mux` do not match the Router interface exactly, so an adapter must be used (see router.go).
The list of implemented routers is here:

**Work out of the box**
* Stdlib's `http.ServeMux`
* `chi`

**Works with the adapter**
* `gorilla`'s mux.
