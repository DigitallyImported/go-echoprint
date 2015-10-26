package main

import (
	"fmt"
	"net/http"
)

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Nothing to see here, move along")
}
