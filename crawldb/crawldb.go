package crawldb

import (
	"encoding/json"
	"fmt"
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/s-rah/onionscan/model"
	"time"
)

type CrawlDB struct {
	myDB *db.DB
}

func (cdb *CrawlDB) NewDB(dbdir string) {
	db, err := db.OpenDB(dbdir)
	if err != nil {
		panic(err)
	}
	cdb.myDB = db

	//If we have just created this db then it will be empty
	if len(cdb.myDB.AllCols()) == 0 {
		cdb.Initialize()
	}

}

func (cdb *CrawlDB) Initialize() {
	if err := cdb.myDB.Create("crawls"); err != nil {
		panic(err)
	}

	crawls := cdb.myDB.Use("crawls")
	if err := crawls.Index([]string{"URL"}); err != nil {
		panic(err)
	}

}

func (cdb *CrawlDB) InsertCrawlRecord(url string, page *model.Page) (int, error) {
	crawls := cdb.myDB.Use("crawls")
	docID, err := crawls.Insert(map[string]interface{}{
		"URL":       url,
		"Timestamp": time.Now(),
		"Page":      page})
	return docID, err
}

type CrawlRecord struct {
	URL       string
	Timestamp time.Time
	Page      model.Page
}

func (cdb *CrawlDB) GetCrawlRecord(id int) (CrawlRecord, error) {
	crawls := cdb.myDB.Use("crawls")
	readBack, err := crawls.Read(id)
	if err == nil {
		out, err := json.Marshal(readBack)
		if err == nil {
			var crawlRecord CrawlRecord
			json.Unmarshal(out, &crawlRecord)
			return crawlRecord, nil
		}
		return CrawlRecord{}, err
	}
	return CrawlRecord{}, err
}

func (cdb *CrawlDB) HasCrawlRecord(url string, duration time.Duration) (bool, int) {
	var query interface{}
	before := time.Now().Add(duration)

	q := fmt.Sprintf(`{"eq":"%v", "in": ["URL"]}`, url)
	json.Unmarshal([]byte(q), &query)

	queryResult := make(map[int]struct{}) // query result (document IDs) goes into map keys
	crawls := cdb.myDB.Use("crawls")
	if err := db.EvalQuery(query, crawls, &queryResult); err != nil {
		panic(err)
	}

	for id := range queryResult {
		// To get query result document, simply read it
		readBack, err := crawls.Read(id)
		if err == nil {
			out, err := json.Marshal(readBack)
			if err == nil {
				var crawlRecord CrawlRecord
				json.Unmarshal(out, &crawlRecord)

				if crawlRecord.Timestamp.After(before) {
					return true, id
				}
			}
		}

	}

	return false, 0
}
