package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorhill/cronexpr"
)

// Index contains data needed to render the index page
type Index struct {
	LastRun              time.Time `db:"date"`
	NextRun              time.Time
	TotalRecords         int `db:"record_count"`
	TotalInconsistencies int `db:"bad_record"`
	Records              []Record
}

// Record holds details of this record
type Record struct {
	Hash string `db:"claimed_hash"`
	Text string `db:"text"`
}

// next returns the next time this interval fires
func next(interval string) time.Time {
	fmt.Println("using interval", interval)
	expr, _ := cronexpr.Parse(interval)
	return expr.Next(time.Now()).UTC()
}

func (s *Server) getIndex() (Index, error) {
	var err error

	var index Index
	if err = s.db.Get(&index, "SELECT `record_count`, `bad_record`, `date` FROM run ORDER BY `id` DESC LIMIT 1"); err != nil {
		return Index{}, err
	}

	// Previous 5 records
	var records []Record
	if err = s.db.Select(&records, "SELECT `claimed_hash`, `text` FROM `record` ORDER BY `id` DESC LIMIT 5"); err != nil {
		return Index{}, err
	}
	index.Records = records
	index.NextRun = next(s.interval)

	return index, nil
}

// GetIndexHandler serves up the home page
func (s *Server) GetIndexHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.log.Debugln(r.Method, r.URL.Path, r.RemoteAddr)
		index, err := s.getIndex()
		if err != nil {
			errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templateResponse(w, "template/index.html", index, 200)
	})
}
