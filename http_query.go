package main

import (
	"io/ioutil"
	"net/http"

	"github.com/AudioAddict/go-echoprint/echoprint"
	"github.com/golang/glog"
)

func queryHandler(w http.ResponseWriter, r *http.Request) {
	jsonData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Error(err)
		apiError(w, err)
		return
	}

	matches, err := peformQuery(jsonData)
	if err != nil {
		apiError(w, err)
		return
	}
	renderResponse(w, matches)
}

func peformQuery(jsonData []byte) ([][]*echoprint.MatchResult, error) {
	codegenList, err := echoprint.ParseCodegen(jsonData)
	if err != nil {
		return nil, err
	}

	matches := echoprint.MatchAll(codegenList)
	return matches, nil
}
