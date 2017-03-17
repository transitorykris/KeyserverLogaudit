package main

import (
	"bufio"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// getKeyServerLog gets the Keyserver log at the given URL
func getKeyserverLog(logURL string) (*bufio.Reader, error) {
	resp, err := http.Get(logURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to retrieve log, got status code %d", resp.StatusCode)
	}
	return bufio.NewReader(resp.Body), nil
}

// Record is a key server record
type Record struct {
	Seen     time.Time `db:"seen"`
	Claimed  string    `db:"claimed_hash"`
	Previous string    `db:"previous_hash"`
	Hash     string    `db:"actual_hash"`
	Text     string    `db:"text"`
}

// nextRecord returns the log message and the claimed SHA256 hash of it
func nextRecord(reader *bufio.Reader) (*Record, error) {
	// Get the record
	record := &Record{}
	line, err := reader.ReadString('\n')
	if err != nil {
		return record, err
	}
	record.Text = strings.TrimSpace(line)

	// Get the claimed hash
	line, err = reader.ReadString('\n')
	if err != nil {
		return record, err
	}
	record.Claimed = strings.TrimPrefix(strings.TrimSpace(line), "SHA256:")

	return record, nil
}

// GenerateHash calculates the actual hash of this record. The hash is the text of the
// log line with a new line followed by the previous hash with no new line.
func (r *Record) GenerateHash(previous *Record) {
	r.Hash = fmt.Sprintf("%x",
		sha256.Sum256(append([]byte(r.Text+"\n"), []byte(previous.Hash)...)),
	)
}

// lastRecord returns the last known good record in the DB
func (w *Worker) lastRecord() (Record, error) {
	var r Record
	err := w.db.Get(&r, "SELECT `claimed_hash`, `previous_hash`, `text` FROM `record` ORDER BY `id` DESC LIMIT 1")
	if err == sql.ErrNoRows {
		return r, nil
	}
	return r, err
}

// saveRecord inserts this log record into the DB
// this should only be used for a record with the correct hash
func (w *Worker) saveRecord(r *Record, prev *Record) error {
	_, err := w.db.Exec(
		"INSERT INTO `record` (`claimed_hash`, `previous_hash`, `actual_hash`, `text`) VALUES (?, ?, ?, ?)",
		r.Claimed, prev.Hash, r.Hash, r.Text,
	)
	return err
}

// saveBadRecord inserts this log record into a table containing records that failed their hash check
func (w *Worker) saveBadRecord(r *Record, prev *Record) error {
	_, err := w.db.Exec(
		"INSERT INTO `bad_record` (`claimed_hash`, `previous_hash`, `actual_hash`, `text`) VALUES (?, ?, ?, ?)",
		r.Claimed, prev.Hash, r.Hash, r.Text,
	)
	return err
}

// getPreviousHashes returns a slice of hashes ordered by insertion
func (w *Worker) getPreviousHashes() ([]string, error) {
	var hashes []string
	err := w.db.Select(&hashes, "SELECT claimed_hash FROM record")
	return hashes, err
}
