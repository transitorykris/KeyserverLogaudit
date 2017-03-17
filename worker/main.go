package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"

	"github.com/transitorykris/hypnos"
)

type specification struct {
	DB         string `envconfig:"db" default:"root:secret@tcp(mysql:3306)/logaudit?parseTime=true"`
	LogURL     string `envconfig:"log_url" default:"https://key.upspin.io/log"`
	Interval   string `envconfig:"interval" default:"* * * * *"`
	UpspinPath string `envconfig:"upspin_path" default:"ann@example.com/logaudit"`
	AWSKey     string `envconfig:"aws_access_key" default:"ABCD1234"`
	AWSSecret  string `envconfig:"aws_secret_key" default:"ZYXW9876"`
	AWSRegion  string `envconfig:"aws_region" default:"us-west-2"`
}

// AWS holds access details for AWS
type AWS struct {
	Region string
	Key    string
	Secret string
}

func main() {
	var err error
	var spec specification
	if err = envconfig.Process("APP", &spec); err != nil {
		logrus.Fatalln(err)
	}
	logrus.Infoln(spec)

	aws := AWS{
		Region: spec.AWSRegion,
		Key:    spec.AWSKey,
		Secret: spec.AWSSecret,
	}

	w := NewWorker(spec.DB, spec.LogURL, spec.UpspinPath, aws)
	for {
		// Wait until the next time we need to run
		if _, _, err := hypnos.Sleep(spec.Interval); err != nil {
			logrus.Fatalln("invalid interval", err)
		}

		// Audit the log
		if err = w.Audit(); err != nil {
			logrus.Errorln("problem with run", err)
			// Do not send notifications, this run is problematic
			continue
		}

		// Save some details about the last audit
		if err := w.saveRun(); err != nil {
			w.log.Errorln("failed to save run", err)
			// Do not send notifications, this run is problematic
			continue
		}

		// Send notifications if there are any
		if err = w.sendNotifications(); err != nil {
			logrus.Errorln("problem with notifications", err)
		}
	}
}
