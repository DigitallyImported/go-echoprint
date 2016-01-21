package echoprint

import (
	"encoding/binary"
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/glog"
	"github.com/rtt/Go-Solr"
)

const (
	// TODO: config
	maxSolrBooleanTerms = 4096
)

type dbResult struct {
	fp         *Fingerprint
	score      float32
	ingestedAt string
}

type dbConnection struct {
	boltDb   *bolt.DB
	solrConn *solr.Connection
}

var errTrackNotFound = errors.New("Failed to find Track in database")
var db *dbConnection

// Purge deletes everything from the databases, used for testing
func Purge() error {
	DBDisconnect()
	os.Remove("echoprint.db")
	db.solrDelete("*:*")
	db = nil

	return DBConnect()
}

// DBConnect establishes necessary databases connections
// TODO: config for db
func DBConnect() error {
	var err error

	if db == nil {
		db = &dbConnection{}
		db.boltDb, err = bolt.Open("echoprint.db", 0600, &bolt.Options{Timeout: 5 * time.Second})
		if err != nil {
			return err
		}

		//db.solrConn, err = solr.Init("72.251.236.164", 8983, "echoprint")
		db.solrConn, err = solr.Init("vagrant-env-platform", 8980, "echoprint")
	}

	return err
}

// DBDisconnect closes database connections
func DBDisconnect() {
	if db != nil {
		db.boltDb.Close()
	}
}

// Query matches fingerprints against the database that meet the minimum code score
func (db *dbConnection) query(fp *Fingerprint, start int, rows int, minScore float32) ([]dbResult, error) {
	t := trackTime("dbConnection.Query")
	defer t.finish()

	glog.V(2).Infof("Querying database rows from %d to %d", start, start+rows)

	// build the unique set of codes for scoring
	var querySet = make(map[uint32]struct{})
	for _, code := range fp.Codes {
		querySet[uint32(code)] = struct{}{}
	}
	glog.V(3).Infof("%d Unique codes for matching", len(querySet))

	numCodes := len(querySet)
	if numCodes > maxSolrBooleanTerms {
		numCodes = maxSolrBooleanTerms - 1
	}

	var codeListParams = make([]string, numCodes)
	var i int
	for code := range querySet {
		codeListParams[i] = strconv.Itoa(int(code))
		i++
		if i >= numCodes {
			break
		}
	}

	q := solr.Query{
		Params: solr.URLParamMap{
			"q": []string{"codes:" + strings.Join(codeListParams, " ")},
		},
		Rows:  rows,
		Start: start,
	}

	resp, err := db.solrSelect(&q)
	if err != nil {
		return nil, err
	}

	glog.V(1).Infof("Solr Matched %d documents in %dms", resp.Results.Len(), resp.QTime)

	var results []dbResult
	for i := 0; i < resp.Results.Len(); i++ {
		doc := resp.Results.Get(i)
		fp, err := db.load(uint32(doc.Field("trackId").(float64)))
		if err != nil {
			return nil, err
		}

		result := dbResult{
			fp:         fp,
			ingestedAt: doc.Field("ingestedAt").(string),
		}

		// construct a unique array of codepoints on the matching fp to calculate the code score
		matchSet := make(map[uint32]struct{})
		for _, code := range result.fp.Codes {
			matchSet[code] = struct{}{}
		}

		result.score = calculateCodeScore(querySet, matchSet)
		if result.score >= minScore {
			glog.V(2).Infof("DB Match above minimum threshold, Score=%f, Meta=%+v", result.score, fp.Meta)
			results = append(results, result)
		} else {
			glog.V(3).Infof("DB Match below minimum threshold, Score=%f, Meta=%+v", result.score, fp.Meta)
		}
	}

	return results, nil
}

// save stores the fingerprint in the database for matching
func (db *dbConnection) save(fp *Fingerprint) error {
	t := trackTime("dbConnection.save")
	defer t.finish()

	doc := map[string]interface{}{
		"add": []interface{}{
			map[string]interface{}{"trackId": fp.Meta.TrackID, "codes": fp.Codes},
		},
	}

	err := db.solrUpdate(doc, false)
	if err != nil {
		return err
	}

	trackIDKey := make([]byte, 4)
	binary.LittleEndian.PutUint32(trackIDKey, fp.Meta.TrackID)

	err = db.boltDb.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(trackIDKey)
		if err != nil {
			return err
		}

		err = b.Put([]byte("codes"), uint32ArrayToBytes(fp.Codes))
		err = b.Put([]byte("times"), uint32ArrayToBytes(fp.Times))
		err = b.Put([]byte("version"), float64ToBytes(fp.Meta.Version))
		err = b.Put([]byte("upc"), []byte(fp.Meta.UPC))
		err = b.Put([]byte("isrc"), []byte(fp.Meta.ISRC))
		err = b.Put([]byte("filename"), []byte(fp.Meta.Filename))
		return err
	})

	if err != nil {
		db.solrDeleteTrack(fp.Meta.TrackID)
	}

	return err
}

func (db *dbConnection) load(trackID uint32) (*Fingerprint, error) {
	t := trackTime("dbConnection.loadMeta")
	defer t.finish()

	fp := &Fingerprint{}
	err := db.boltDb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(uint32ToBytes(trackID))
		if b == nil {
			return errTrackNotFound
		}

		fp.Codes = bytesToUint32Array(b.Get([]byte("codes")))
		fp.Times = bytesToUint32Array(b.Get([]byte("times")))
		fp.Meta.TrackID = trackID
		fp.Meta.Version = bytesTofloat64(b.Get([]byte("version")))
		fp.Meta.UPC = string(b.Get([]byte("upc")))
		fp.Meta.ISRC = string(b.Get([]byte("isrc")))
		fp.Meta.Filename = string(b.Get([]byte("filename")))
		return nil
	})

	return fp, err
}

func (db *dbConnection) checkTrackExists(trackID uint32) (bool, error) {
	var exists bool
	err := db.boltDb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(uint32ToBytes(trackID))
		exists = b != nil
		return nil
	})

	return exists, err
}

func (db *dbConnection) solrDeleteTrack(trackID uint32) error {
	return db.solrDelete("trackId:" + strconv.Itoa(int(trackID)))
}

func (db *dbConnection) solrDelete(q string) error {
	doc := map[string]interface{}{
		"delete": map[string]interface{}{
			"query": q,
		},
	}

	return db.solrUpdate(doc, true)
}

func (db *dbConnection) solrUpdate(document map[string]interface{}, commit bool) (err error) {
	resp, err := db.solrConn.Update(document, commit)
	if err == nil && !resp.Success {
		err = errors.New(resp.String())
	}

	return
}

func (db *dbConnection) solrSelect(q *solr.Query) (resp *solr.SelectResponse, err error) {
	resp, err = db.solrConn.Select(q)
	if err == nil && resp.Status != 0 {
		err = errors.New("Solr select() failed")
	}

	return
}

// calculateCodeScore does a basic intersection on the unique code values
// to determine if we should even consider it for a histogram match
func calculateCodeScore(qSet, mSet map[uint32]struct{}) float32 {
	t := trackTime("calculateCodeScore")
	defer t.finish()

	var count int
	for code := range qSet {
		if _, ok := mSet[code]; ok {
			count++
		}
	}

	return float32(count) / float32(len(qSet)) * 100.00
}

func uint32ToBytes(i uint32) []byte {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, i)
	return bytes
}

func bytesTofloat64(bytes []byte) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(bytes))
}

func float64ToBytes(float float64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, math.Float64bits(float))
	return bytes
}

func uint32ArrayToBytes(data []uint32) []byte {
	bytes := make([]byte, len(data)*4)
	for i, val := range data {
		offset := i * 4
		binary.LittleEndian.PutUint32(bytes[offset:offset+4], val)
	}

	return bytes
}

func bytesToUint32Array(bytes []byte) []uint32 {
	data := make([]uint32, len(bytes)/4)
	for i := range data {
		offset := i * 4
		data[i] = binary.LittleEndian.Uint32(bytes[offset : offset+4])
	}
	return data
}
