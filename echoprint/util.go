package echoprint

import (
	"time"

	"github.com/golang/glog"
)

var timeTrackers map[string][]*timeTracker

func init() {
	timeTrackers = make(map[string][]*timeTracker, 0)
}

type timeTracker struct {
	Label   string
	Start   time.Time
	Elapsed time.Duration
}

func trackTime(label string) *timeTracker {
	return &timeTracker{label, time.Now(), 0}
}

func (tt *timeTracker) finish() {
	tt.Elapsed = time.Since(tt.Start)
	timeTrackers[tt.Label] = append(timeTrackers[tt.Label], tt)
	glog.V(3).Infof("-- %s took %s", tt.Label, tt.Elapsed)
}

// TotalTime returns the total duration of all timed functions
// func TotalTime() (total time.Duration) {
// 	for _, tt := range timeTrackers {
// 		total += tt.Elapsed
// 	}
//
// 	return
// }
