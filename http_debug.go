package main

import "net/http"

func debugHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		op := r.FormValue("debug_op")
		data := r.FormValue("data")
		switch op {
		case "Ingest":
			// TODO: ingest
		case "Query":
			matches, err := peformQuery([]byte(data))
			if err != nil {
				apiError(w, err)
				return
			}
			renderResponse(w, matches)
		}
	} else {
		renderView(w, "debug", nil)
	}
}
