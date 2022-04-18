package main

import (
	"log"
	"net/http"
	"regexp"
	"text/template"

	"github.com/mitchellh/go-wordwrap"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		// handle returned error here.
		w.WriteHeader(503)
		w.Write([]byte("bad"))
	}
}

func render(w http.ResponseWriter, data map[string]interface{}, patterns ...string) error {
	err := parse(patterns...).ExecuteTemplate(w, "main", data)
	if err != nil {
		http.Error(w, err.Error(), 0)
		log.Print("Template executing error: ", err)
	}
	return err
}

func parse(patterns ...string) *template.Template {
	for i := 0; i < len(patterns); i++ {
		patterns[i] = "templates/" + patterns[i]
	}
	return template.Must(template.New("main").Funcs(funcs).ParseFiles(patterns...))
}

var funcs = template.FuncMap{
	"wrap": func(s string) string {
		wrapped := wordwrap.WrapString(s, 29)
		re := regexp.MustCompile(`\r?\n`)
		wrapped = re.ReplaceAllString(wrapped, "\n    ")
		return wrapped
	},
}
