package echoprint

import (
	"sort"
	"sync"

	"github.com/golang/glog"
)

const (
	maxIngestDuration  = 60 * 60 * 4
	histogramMatchSlop = 2

	minMatchScorePercent = 0.05 * 100
	minMatchConfidence   = 0.25 * 100
	bestMatchDiff        = 0.25
	maxConfidence        = 100.00
)

// MatchResult represents a response from the fingerprint matching algorithm
type MatchResult struct {
	fp         *Fingerprint
	Best       bool    `json:"best"`
	TrackID    uint32  `json:"track_id"`
	UPC        string  `json:"upc"`
	ISRC       string  `json:"isrc"`
	Confidence float32 `json:"confidence"`
	IngestedAt string  `json:"ingested_at"`
}

// MatchError holds information for a match that failed due to an error during processing
type MatchError struct {
	Error string `json:"error"`
}

// implement sort.Interface for MatchResults to sort by confidence (descending)
type byConfidence []*MatchResult

func (m byConfidence) Len() int           { return len(m) }
func (m byConfidence) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m byConfidence) Less(i, j int) bool { return m[i].Confidence > m[j].Confidence }

func newMatchResult(r dbResult) *MatchResult {
	return &MatchResult{
		fp:         r.fp,
		TrackID:    r.trackID,
		IngestedAt: r.ingestedAt,
	}
}

// MatchAll performs mutiple matches in parallel, results are grouped by the index of the
// fingerprint list so they may be returned in the order they are received
func MatchAll(codegenList []CodegenFp) ([]interface{}, error) {
	var allMatches = make([]interface{}, len(codegenList))
	var wg sync.WaitGroup

	for i, codegenFp := range codegenList {
		wg.Add(1)
		go func(group int, codegenFp CodegenFp) {
			defer wg.Done()

			glog.Infof("Processing codegen TrackID=%d, Version=%f, Filename=%s\n",
				codegenFp.TrackID, codegenFp.Metadata.Version, codegenFp.Metadata.Filename)

			fp, err := NewFingerprint(codegenFp.Code, codegenFp.Metadata.Version)
			if err != nil {
				glog.Info(err)
				allMatches[group] = MatchError{err.Error()}
				return
			}

			matches, err := Match(fp)
			if err != nil {
				glog.Info(err)
				allMatches[group] = MatchError{err.Error()}
				return
			}

			glog.Info("Number of matches found:", len(matches))
			allMatches[group] = matches
		}(i, codegenFp)
	}

	wg.Wait()

	return allMatches, nil
}

// Match attempts to find the fingerprint provided in the database and returns an array of MatchResult
func Match(fp *Fingerprint) ([]*MatchResult, error) {
	t := trackTime("Match")
	defer t.finish()

	if !fp.clamped {
		fp = fp.NewClamped()
	}

	var matches []*MatchResult
	results, err := db.Query(fp, 500, minMatchScorePercent)

	for _, r := range results {
		match := newMatchResult(r)
		match.Confidence = calculateConfidence(fp, match.fp, uint32(histogramMatchSlop))
		if match.Confidence >= minMatchConfidence {
			glog.V(1).Info("Match result above minimum threshold, Confidence=", match.Confidence, " TrackID=", match.TrackID)
			matches = append(matches, match)
		} else {
			glog.V(2).Info("Match result below minimum threshold, Confidence=", match.Confidence, " TrackID=", match.TrackID)
		}
	}

	numMatches := len(matches)

	if numMatches > 0 {
		sort.Sort(byConfidence(matches))
		determineBestMatch(matches)
		clampMatchConfidence(matches)
	}

	return matches, err
}

// determine if we have a "best" match
func determineBestMatch(matches []*MatchResult) {
	if len(matches) == 1 {
		matches[0].Best = true
		glog.V(2).Infof("Single good match, marking as best: %+v", matches[0])
	} else {
		// top match is different enough to call it best
		if matches[0].Confidence-matches[1].Confidence >= matches[0].Confidence*bestMatchDiff {
			matches[0].Best = true
			glog.V(2).Infof("Multiple good matches, top result is different enough, marking as best: %+v", matches[0])
		} else {
			glog.V(2).Info("Multiple good matches, top result is not different enough, no best match found")
		}
	}
}

func clampMatchConfidence(matches []*MatchResult) {
	for _, match := range matches {
		if match.Confidence > maxConfidence {
			match.Confidence = maxConfidence
		}
	}
}

func calculateConfidence(fp *Fingerprint, matchFp *Fingerprint, slop uint32) float32 {
	t := trackTime("calculateActualScore")
	defer t.finish()

	timeDiffs := make(map[int]uint16)
	matchCodeMap := getCodeTimeMap(matchFp, slop)
	for i, code := range fp.Codes {
		fpTime := fp.Times[i] / slop * slop

		if matchTimes, ok := matchCodeMap[code]; ok {
			for _, matchTime := range matchTimes {
				dist := int(fpTime - matchTime)
				if dist < 0 {
					dist = -dist
				}
				timeDiffs[dist]++
			}
		}
	}

	var timeDiffVals []int
	for _, key := range timeDiffs {
		timeDiffVals = append(timeDiffVals, int(key))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(timeDiffVals)))

	var score int
	if len(timeDiffVals) > 0 {
		score = timeDiffVals[0]
		if len(timeDiffVals) > 1 {
			score += timeDiffVals[1]
		}
	}

	return float32(score) / float32(len(fp.Codes)) * 100.00
}

func getCodeTimeMap(fp *Fingerprint, slop uint32) map[uint32][]uint32 {
	codeMap := make(map[uint32][]uint32, len(fp.Codes))

	for i, code := range fp.Codes {
		time := fp.Times[i] / slop * slop
		codeMap[code] = append(codeMap[code], time)
	}

	return codeMap
}
