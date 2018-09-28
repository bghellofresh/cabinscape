package storage

import (
	"database/sql"
	"fmt"
)

// PgDB represents the database connection
type PgDB struct {
	dbConn *sql.DB
}

// DBConfig represents the storage configuration
type DBConfig struct {
	DBHost string
	DBName string
	DBUser string
	DBPass string
}

// Init initialises the db connection
func Init(config *DBConfig) (*PgDB, error) {
	connectString := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable",
		config.DBUser, config.DBPass, config.DBHost, config.DBName)
	if dbConn, err := sql.Open("postgres", connectString); err != nil {
		return nil, err
	} else {
		p := &PgDB{dbConn: dbConn}
		if err := p.dbConn.Ping(); err != nil {
			return nil, err
		}
		return p, nil
	}
}

// CreateTablesIfNotExist creates the DB schema
func (p *PgDB) CreateTablesIfNotExist() error {
	createSQL := `
       CREATE TABLE IF NOT EXISTS events (
       uid TEXT NOT NULL PRIMARY KEY,
       summary TEXT NOT NULL,
       dtstart TIMESTAMP NOT NULL,
	   dtend TIMESTAMP NOT NULL);`

	if rows, err := p.dbConn.Query(createSQL); err != nil {
		return err
	} else {
		rows.Close()
	}

	return nil
}

// InsertOrUpdateEvent adds or updates rows in the database
func (p *PgDB) InsertOrUpdateEvent(uid string, start string, end string, summary string) error {
	lookupSQL := `
	SELECT uid, summary, dtstart, dtend FROM events
	WHERE dtstart = $1 AND dtend = $2
`
	exists := true
	var oldSummary, oldUID string
	var oldStart, oldEnd string
	newStart := start
	newEnd := end
	r := p.dbConn.QueryRow(lookupSQL, newStart, newEnd)
	err := r.Scan(&oldUID, &oldSummary, &oldStart, &oldEnd)
	if err == sql.ErrNoRows {
		exists = false
	} else if err != nil {
		return err
	}
	if !exists {
		insertSQL := `
		INSERT INTO events
		(uid, dtstart, dtend, summary)
		VALUES ($1, $2, $3, $4)
	`
		_, err := p.dbConn.Exec(insertSQL, uid, newStart, newEnd, summary)
		if err != nil {
			return err
		}
	} else {
		if newStart != oldStart || newEnd != oldEnd || summary != oldSummary {
			updateSQL := `
			UPDATE events
			SET dtstart = $3, dtend = $4, summary = $5, uid = $2
			WHERE uid = $1
		`
			_, err := p.dbConn.Exec(updateSQL, oldUID, uid, newStart, newEnd, summary)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetEvents returns events from storage
func (p *PgDB) GetEvents() (*sql.Rows, error) {
	lookupSQL := `
        SELECT uid, summary, dtstart, dtend FROM events
  	`
	r, err := p.dbConn.Query(lookupSQL)
	if err != nil {
		return nil, err
	}
	return r, nil
}
