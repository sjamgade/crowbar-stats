package sqlite

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	// This loads the postgres drivers.
	_ "github.com/mattn/go-sqlite3"

	"crowbar-stats/storage"
)

// New returns a postgres backed storage service.
func New(dbname string) (storage.Service, error) {
	// Coonect postgres
	db, err := sql.Open("sqlite3", dbname)
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	strQuery := "CREATE TABLE IF NOT EXISTS noderuns (id INTEGER PRIMARY KEY AUTOINCREMENT," +
		"nodename TEXT not NULL, " +
		"start DATE DEFAULT (datetime('now','localtime'))," +
		"finish DATE," +
		"datapath TEXT);"

	_, err = db.Exec(strQuery)
	if err != nil {
		return nil, err
	}
	folderpath := "./run.json.d"
	os.MkdirAll(folderpath, os.ModePerm)

	return &sqlite{db, folderpath}, nil
}

type sqlite struct {
	db         *sql.DB
	folderpath string
}

func (p *sqlite) Startrun(nodename string) (string, error) {
	var id int
	_, err := p.db.Exec("INSERT INTO noderuns(nodename) VALUES(?);", nodename)
	if err != nil {
		return "", err
	}
	find := "SELECT id FROM noderuns WHERE nodename is (?) ORDER BY datetime(start) DESC LIMIT 1;"
	row, err := p.db.Query(find, nodename)
	defer row.Close()
	row.Next()
	err = row.Scan(&id)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(id), nil
}

func (p *sqlite) Stoprun(rid string, payload []byte) (string, error) {
	var did int
	id, err := strconv.Atoi(rid)
	if err != nil {
		return "", fmt.Errorf("Unable to parse id %v: %v", rid, err)
	}

	find := "SELECT id FROM noderuns WHERE id = (?) AND datapath is null ORDER BY datetime(start) DESC LIMIT 1;"
	row, err := p.db.Query(find, id)
	row.Next()
	err = row.Scan(&did)
	if err != nil {
		return "", fmt.Errorf("run id does not exist %v or run already finished ", rid, err)
	}
	if id != did {
		return "", fmt.Errorf("run id does not match, or run id already finished")
	}
	row.Close()

	find = "UPDATE noderuns SET finish = ? , datapath = ? WHERE id IN" +
		"(SELECT id FROM noderuns WHERE id is (?) ORDER BY datetime(start) DESC LIMIT 1);"
	uuid, err := newUUID()
	jsonfilename := fmt.Sprintf("%v.json", uuid)

	err = ioutil.WriteFile(p.folderpath+"/"+jsonfilename, payload, 0644)
	if err != nil {
		return "", fmt.Errorf("Unable to write %v: %v", jsonfilename, err)
	}
	_, err = p.db.Exec(find, time.Now(), jsonfilename, id)
	if err != nil {
		return "", fmt.Errorf("File created %v, but not written to database %v", jsonfilename, err)
	}

	return "", nil
}

// close the db connection
func (p *sqlite) Close() error { return p.db.Close() }

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
