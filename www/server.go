package main

import (
	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// Server represents our API server
type Server struct {
	db       *sqlx.DB
	log      *logrus.Logger
	interval string
	aws      AWS
}

// NewServer creates a new Goldblum server
func NewServer(interval string, aws AWS) (*Server, error) {
	// XXX REMOVE INTERVAL
	server := Server{
		log:      logrus.New(),
		interval: interval,
		aws:      aws,
	}
	return &server, nil
}

// ConnectDB connects our server to the given DB
func (s *Server) ConnectDB(db string) error {
	var err error
	if s.db, err = sqlx.Connect("mysql", db); err != nil {
		return err
	}
	return nil
}

// Router returns and HTTP router with the handlers for this server
func (s *Server) Router() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/signup", s.PostSignupHandler()).Methods("POST")
	r.Handle("/signup/{code}", s.GetSignupHandler()).Methods("GET")
	r.PathPrefix("/").Handler(s.GetIndexHandler()).Methods("GET")
	return r
}

// Close closes down the server
func (s *Server) Close() error {
	return s.db.Close()
}
