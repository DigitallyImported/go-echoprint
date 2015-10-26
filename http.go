package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
)

type errorResponse struct {
	Error string `json:"error"`
}

var views = template.Must(template.ParseGlob("views/*.html"))

func renderResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		httpError(w, err)
	}
}

func renderView(w http.ResponseWriter, viewName string, data interface{}) {
	if err := views.ExecuteTemplate(w, viewName+".html", data); err != nil {
		httpError(w, err)
	}
}

func httpError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func apiError(w http.ResponseWriter, err error) {
	w.WriteHeader(422)
	renderResponse(w, &errorResponse{err.Error()})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Nothing to see here, move along")
}
