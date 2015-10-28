package main

import (
	"io/ioutil"
	"net/http"

	"github.com/AudioAddict/go-echoprint/echoprint"
	"github.com/golang/glog"
)

type queryResult struct {
	Matches []*echoprint.MatchResult `json:"matches"`
	Status  string                   `json:"status"`
}

func newQueryResult(matches []*echoprint.MatchResult) queryResult {
	qr := queryResult{Matches: matches}
	if len(matches) > 0 {
		if matches[0].Best {
			qr.Status = "BEST_MATCH"
		} else {
			qr.Status = "POSSIBLE_MATCH"
		}
	} else {
		qr.Status = "NO_MATCH"
	}

	return qr
}

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

func peformQuery(jsonData []byte) ([]queryResult, error) {
	codegenList, err := echoprint.ParseCodegen(jsonData)
	if err != nil {
		return nil, err
	}

	matchGroups := echoprint.MatchAll(codegenList)
	result := make([]queryResult, len(matchGroups))
	for i, group := range matchGroups {
		result[i] = newQueryResult(group)
	}
	return result, nil
}
