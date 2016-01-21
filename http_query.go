package main

import (
	"io/ioutil"
	"net/http"
	"runtime/debug"

	"github.com/AudioAddict/go-echoprint/echoprint"
	"github.com/golang/glog"
)

const (
	statusBestMatch      = "BEST_MATCH"
	statusDuplicateMatch = "DUPLICATE_MATCH"
	statusPossibleMatch  = "POSSIBLE_MATCH"
	statusNoMatch        = "NO_MATCH"
	statusError          = "ERROR"
)

type queryResult struct {
	Matches    []*echoprint.MatchResult `json:"matches"`
	Status     string                   `json:"status"`
	MatchCount int                      `json:"match_count"`
}

func newQueryResult(matches []*echoprint.MatchResult) queryResult {
	qr := queryResult{Matches: matches}
	qr.MatchCount = len(matches)

	if qr.MatchCount > 0 {
		if qr.MatchCount == 1 && matches[0].Error != nil {
			qr.Status = statusError
		} else if matches[0].Best {
			qr.Status = statusBestMatch
		} else {
			if qr.MatchCount > 1 && matches[0].Confidence == 100 && matches[1].Confidence == 100 {
				qr.Status = statusDuplicateMatch
			} else {
				qr.Status = statusPossibleMatch
			}
		}
	} else {
		qr.Status = statusNoMatch
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

	result, err := peformQuery(jsonData)
	if err != nil {
		apiError(w, err)
		return
	}

	renderResponse(w, result)
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

	debug.FreeOSMemory()
	return result, nil
}
