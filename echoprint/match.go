package echoprint

import (
	"log"
	"strconv"
	"strings"

	"github.com/AudioAddict/go-solr/solr"
)

const (
	maxIngestDuration  = 60 * 60 * 4
	histogramMatchSlop = 2

	minMatchScorePercent = 0.05 * 100
	minMatchConfidence   = 0.25 * 100
	bestMatchDiff        = 0.25
)

// MatchMaker holds information for making matches
type MatchMaker struct {
	si *solr.SolrInterface
}

// MatchResult represents a response from the fingerprint matching algorithm
type MatchResult struct {
	fp    *Fingerprint
	score float32
}

func newMatchResult() MatchResult {
	return MatchResult{
		fp: &Fingerprint{},
	}
}

// NewMatchMaker returns an instance of MatchMaker
func NewMatchMaker() (m *MatchMaker, err error) {
	m = &MatchMaker{}

	// TODO: config
	m.si, err = solr.NewSolrInterface("http://vagrant-env-platform:8980/solr", "echoprint")
	return
}

// Match attempts to find the fingerprint provided in the database
// and returns an array of MatchResult
func (mm *MatchMaker) Match(fp *Fingerprint) ([]MatchResult, error) {
	t := trackTime("Match")
	defer t.finish(true)

	if !fp.clamped {
		fp = fp.NewClamped()
	}

	matches, err := mm.fpQuery(fp.Codes, 30, minMatchScorePercent)
	return matches, err
}

func (mm *MatchMaker) fpQuery(codes []uint32, rows int, minScore float32) ([]MatchResult, error) {
	var querySet = make(map[uint32]struct{})

	// build the unique set of codes for scoring
	for _, code := range codes {
		querySet[uint32(code)] = struct{}{}
	}

	var codeListParams = make([]string, len(querySet))
	var i int
	for code := range querySet {
		codeListParams[i] = strconv.Itoa(int(code))
		i++
	}

	query := solr.NewQuery()
	query.Q("codes:" + strings.Join(codeListParams, " "))
	query.Rows(rows)
	//query.FieldList("id,score")
	s := mm.si.Search(query)
	r, err := s.Result(nil)
	if err != nil {
		return nil, err
	}

	log.Printf("Number of documents: %d", r.Results.NumFound)

	var matches []MatchResult
	var codeInt uint32
	for _, doc := range r.Results.Docs {
		m := newMatchResult()
		matchSet := make(map[uint32]struct{})
		codes := doc.Get("codes").([]interface{})
		times := doc.Get("times").([]interface{})

		m.fp.Codes = make([]uint32, len(codes))
		m.fp.Times = make([]uint32, len(times))

		for i, val := range codes {
			codeInt = uint32(val.(float64))
			matchSet[codeInt] = struct{}{}
			m.fp.Codes[i] = codeInt
			m.fp.Times[i] = uint32(times[i].(float64))
		}

		score := calculateScore(querySet, matchSet)
		//log.Println("Score=", m.score)
		if score >= minScore {
			m.score = score
			matches = append(matches, m)
		}
	}

	return matches, nil
}

func calculateScore(qSet, mSet map[uint32]struct{}) float32 {
	t := trackTime("calculateScore")
	defer t.finish(false)

	var count int
	for code := range qSet {
		if _, ok := mSet[code]; ok {
			count++
		}
	}

	return float32(count) / float32(len(qSet)) * 100.00
}
