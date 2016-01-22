package echoprint

import (
	"errors"
	"sync"

	"github.com/golang/glog"
)

// IngestResult represents the status of ingesting a fingerprint
type IngestResult struct {
	TrackID uint32      `json:"track_id"`
	Error   interface{} `json:"error"`
}

// ErrTrackIDExists is returned during ingestion when the provided TrackID already exists in the database
var ErrTrackIDExists = errors.New("TrackID already exists in the database")

// ErrTrackIDMissing is returned during ingestion when no TrackID is provided
var ErrTrackIDMissing = errors.New("Missing Track ID")

// IngestAll takes an array of CodegenFp and stores them in the database in parallel
func IngestAll(codegenList []*CodegenFp) []IngestResult {
	var results = make([]IngestResult, len(codegenList))
	var wg sync.WaitGroup

	for i, codegenFp := range codegenList {
		wg.Add(1)
		go func(group int, codegenFp *CodegenFp) {
			defer wg.Done()

			glog.Infof("Processing codegen %+v\n", codegenFp.Meta)

			fp, err := NewFingerprint(codegenFp)
			if err != nil {
				results[group] = IngestResult{Error: err.Error()}
				return
			}

			err = Ingest(fp)
			if err != nil {
				results[group] = IngestResult{Error: err.Error()}
				return
			}

			glog.Infof("Ingested Fingerprint %+v", fp.Meta)
			results[group] = IngestResult{TrackID: fp.Meta.TrackID}
		}(i, codegenFp)
	}

	wg.Wait()

	return results
}

// Ingest takes a single CodegenFp and stores it in the database for matching
func Ingest(fp *Fingerprint) error {

	if fp.Meta.TrackID == 0 {
		glog.V(3).Info("TrackID is missing, aborting ingestion")
		return ErrTrackIDMissing
	}

	exists, err := db.checkTrackExists(fp.Meta.TrackID)
	if err != nil {
		glog.Error(err)
		return err
	}

	if exists {
		glog.V(3).Infof("TrackID=%d already exists, aborting ingestion", fp.Meta.TrackID)
		return ErrTrackIDExists
	}

	glog.V(3).Infof("TrackID=%d does not exist, starting ingestion", fp.Meta.TrackID)

	err = db.save(fp)

	return nil
}
