package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/AudioAddict/go-echoprint/echoprint"
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

func queryHandler(w http.ResponseWriter, r *http.Request) {
	var jsonData []byte
	r.Body.Read(jsonData)

	// data := &queryResponse{apiResponse{Status: "success?!", Error: nil}}
	// renderResponse(w, data)
}

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Nothing to see here, move along")
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		op := r.FormValue("debug_op")
		data := r.FormValue("data")
		switch op {
		case "Ingest":
			// TODO: ingest
		case "Query":
			codegenList, err := echoprint.ParseCodegen([]byte(data))
			if err != nil {
				apiError(w, err)
				return
			}

			allMatches, err := echoprint.MatchAll(codegenList)
			if err != nil {
				apiError(w, err)
				return
			}

			renderResponse(w, allMatches)
		}
	} else {
		renderView(w, "debug", nil)
	}
}
