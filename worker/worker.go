package main

import (
	"time"

	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"upspin.io/client"
	"upspin.io/config"
	"upspin.io/transports"
	"upspin.io/upspin"
)

// Worker is all the stuff this worker needs to do its job
type Worker struct {
	db         *sqlx.DB
	log        *logrus.Logger
	us         upspin.Client
	logURL     string
	upspinPath string
	aws        AWS
}

// NewWorker creates a brand new worker
func NewWorker(db string, logURL string, upspinPath string, aws AWS) *Worker {
	var err error
	w := &Worker{
		logURL:     logURL,
		upspinPath: upspinPath,
		aws:        aws,
	}

	w.log = logrus.New()
	w.log.Formatter = new(logrus.TextFormatter)
	w.log.Level = logrus.DebugLevel

	for {
		w.db, err = sqlx.Connect("mysql", db)
		if err != nil {
			w.log.Warnln("Problem connecting to DB", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	conf, err := config.FromFile("./upspin/config")
	if err != nil {
		w.log.Fatalln("Could not load config file", err)
	}
	transports.Init(conf)
	w.us = client.New(conf)

	return w
}

// Close down our worker properly
func (w *Worker) Close() error {
	w.db.Close()
	return nil
}
