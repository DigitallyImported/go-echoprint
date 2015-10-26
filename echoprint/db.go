package echoprint

import (
	"strconv"
	"strings"
	"time"

	"github.com/AudioAddict/go-solr/solr"
	"github.com/boltdb/bolt"
	"github.com/golang/glog"
)

type dbResult struct {
	fp         *Fingerprint
	trackID    uint32
	upc        string
	isrc       string
	score      float32
	ingestedAt string
}

type dbConnection struct {
	boltDb *bolt.DB
	si     *solr.SolrInterface
}

var db *dbConnection

func newDBResult(trackID uint32) dbResult {
	return dbResult{
		fp:      &Fingerprint{},
		trackID: trackID,
	}
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

		//db.si, err = solr.NewSolrInterface("http://72.251.236.164:8983/solr", "echoprint")
		db.si, err = solr.NewSolrInterface("http://vagrant-env-platform:8980/solr", "echoprint")
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
func (db *dbConnection) Query(fp *Fingerprint, rows int, minScore float32) ([]dbResult, error) {

	// build the unique set of codes for scoring
	//var codeListParams = make([]string, len(codes))
	var querySet = make(map[uint32]struct{})
	for _, code := range fp.Codes {
		querySet[uint32(code)] = struct{}{}
		//		codeListParams[i] = strconv.Itoa(int(code))
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

	s := db.si.Search(query)
	r, err := s.Result(nil)
	if err != nil {
		return nil, err
	}

	glog.V(1).Infof("Number of documents matched: %d", r.Results.NumFound)

	var results []dbResult
	var codeInt uint32
	for _, doc := range r.Results.Docs {
		result := newDBResult(uint32(doc.Get("trackId").(float64)))
		result.ingestedAt = doc.Get("ingestedAt").(string)

		codes := doc.Get("codes").([]interface{})
		times := doc.Get("times").([]interface{})

		result.fp.Codes = make([]uint32, len(codes))
		result.fp.Times = make([]uint32, len(times))

		// construct a unique array of codepoints on the matching fp to calculate the code score
		matchSet := make(map[uint32]struct{})
		for i, val := range codes {
			codeInt = uint32(val.(float64))
			matchSet[codeInt] = struct{}{}
			result.fp.Codes[i] = codeInt
			result.fp.Times[i] = uint32(times[i].(float64))
		}

		result.score = calculateCodeScore(querySet, matchSet)
		if result.score >= minScore {
			glog.V(1).Info("DB Match above minimum threshold, Score=", result.score, " Filename=", doc.Get("filename").(string))
			results = append(results, result)
		} else {
			glog.V(2).Info("DB Match below minimum threshold, Score=", result.score, " Filename=", doc.Get("filename").(string))
		}
	}

	return results, nil
}

// Ingest stores fingerprints in the database for matching
func (db *dbConnection) Ingest() {

}

// Purge deletes everything from the databases, used for testing
func (db *dbConnection) Purge() {

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

//
// import (
//   "github.com/szferi/gomdb"
// )
//
// // create a directory to hold the database
// path, err := ioutil.TempDir("", "mdb_test")
// dieOrNah(err)
//
// log.Println("Setup DB in:", path)
// defer os.RemoveAll(path)
//
// // open the db
// env, err := mdb.NewEnv()
// env.SetMapSize(1 << 40) // max file size
// env.Open(path, 0, 0664)
// defer env.Close()
//
// txn, _ := env.BeginTxn(nil, 0)
// dbi, _ := txn.DBIOpen(nil, 0)
// defer env.DBIClose(dbi)
// txn.Commit()
//
// fpList, err := echoprint.ParseCodegenFile(os.Args[1])
// dieOrNah(err)
//
// for _, codegenFp := range fpList {
//
//   fp, err := echoprint.NewFingerprint(fpList[0].Code, fpList[0].Metadata.Version)
//   dieOrNah(err)
//
//   txn, err = env.BeginTxn(nil, 0)
//   dieOrNah(err)
//
//   trackID := make([]byte, 4)
//   binary.LittleEndian.PutUint32(trackID, codegenFp.TrackID)
//
//   // bval, err := txn.Get(dbi, key)
//   // if err != mdb.NotFound {
//   // 	dieOrNah(err)
//   // }
//
//   trackData := make([]byte, len(fp.Codes)*4)
//   for i, code := range fp.Codes {
//     codeBytes := make([]byte, 4)
//     binary.LittleEndian.PutUint32(codeBytes, code)
//
//     offset := i * 4
//     trackData[offset] = codeBytes[0]
//     trackData[offset+1] = codeBytes[1]
//     trackData[offset+2] = codeBytes[2]
//     trackData[offset+3] = codeBytes[3]
//
//     codeValues, err := txn.Get(dbi, codeBytes)
//     if err != mdb.NotFound {
//       dieOrNah(err)
//     }
//
//     codeValues = append(codeValues, trackID...)
//     err = txn.Put(dbi, codeBytes, codeValues, 0)
//     dieOrNah(err)
//   }
//
//   err = txn.Put(dbi, trackID, trackData, 0)
//   dieOrNah(err)
//
//   txn.Commit()
// }
//
// // inspect the database
// stat, _ := env.Stat()
// fmt.Println(stat.Entries)
//
// // scan the database
// txn, _ = env.BeginTxn(nil, mdb.RDONLY)
// defer txn.Abort()
// cursor, _ := txn.CursorOpen(dbi)
// defer cursor.Close()
// for {
//   bkey, bval, err := cursor.Get(nil, nil, mdb.NEXT)
//   if err == mdb.NotFound {
//     break
//   }
//   if err != nil {
//     panic(err)
//   }
//   fmt.Printf("%b: %d bytes\n", bkey, len(bval))
// }
