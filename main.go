// Gluten Free Horoscopes
package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	handler "github.com/nathanhollows/todo/handlers"
)

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Handle("/", handler.Handler(handler.IndexHandler))
	r.Handle("/print", handler.Handler(handler.PrintHandler))

	http.ListenAndServe(":8111", r)

}
