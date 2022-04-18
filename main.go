// Gluten Free Horoscopes
package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Handle("/", Handler(IndexHandler))
	r.Handle("/print", Handler(PrintHandler))

	http.ListenAndServe(":8111", r)

}
