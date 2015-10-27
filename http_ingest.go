package main

import (
	"io/ioutil"
	"net/http"

	"github.com/AudioAddict/go-echoprint/echoprint"
	"github.com/golang/glog"
)

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	jsonData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Error(err)
		apiError(w, err)
		return
	}

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
