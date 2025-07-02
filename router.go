package checkpoint

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Router interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	Handle(string, http.Handler)
}
type RouterAdapter struct {
	Mux interface{}
}

func (g *RouterAdapter) Handle(pattern string, handler http.Handler) {
	switch m := g.Mux.(type) {
	case *mux.Router:
		m.Handle(pattern, handler)
	}
}
func (g *RouterAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch m := g.Mux.(type) {
	case *mux.Router:
		m.ServeHTTP(w, r)
	default:
		http.Error(w, "Unsupported router type", http.StatusInternalServerError)
	}
}
