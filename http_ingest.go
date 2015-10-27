package main

import (
	"net/http"

	"github.com/AudioAddict/go-echoprint/echoprint"
)

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	var jsonData []byte
	r.Body.Read(jsonData)

	results, err := peformIngest(jsonData)
	if err != nil {
		apiError(w, err)
		return
	}
	renderResponse(w, results)
}

func peformIngest(jsonData []byte) ([]echoprint.IngestResult, error) {
	codegenList, err := echoprint.ParseCodegen(jsonData)
	if err != nil {
		return nil, err
	}

	results := echoprint.IngestAll(codegenList)
	return results, nil
}
