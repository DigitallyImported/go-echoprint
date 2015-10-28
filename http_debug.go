package main

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/AudioAddict/go-echoprint/echoprint"
)

var statsInfo *stats

func init() {
	statsInfo = &stats{
		Memory: new(runtime.MemStats),
	}
}

type stats struct {
	Memory *runtime.MemStats
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		op := r.FormValue("debug_op")
		data := r.FormValue("data")

		var err error
		var results interface{}
		switch op {
		case "Ingest":
			results, err = peformIngest([]byte(data))
		case "Query":
			results, err = peformQuery([]byte(data))
		}

		if err != nil {
			apiError(w, err)
			return
		}
		renderResponse(w, results)

	} else {
		renderView(w, "debug", nil)
	}
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	runtime.ReadMemStats(statsInfo.Memory)

	renderResponse(w, statsInfo)
}

func purgeHandler(w http.ResponseWriter, r *http.Request) {
	echoprint.Purge()
	fmt.Fprint(w, "Done")
}
