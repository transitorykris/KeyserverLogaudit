package main

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
)

type specification struct {
	Bind      string `envconfig:"bind" default:":8080"`
	DB        string `envconfig:"db" default:"root:secret@tcp(mysql:3306)/logaudit?parseTime=true"`
	Interval  string `envconfig:"interval" default:"23 */1 * * *"`
	AWSKey    string `envconfig:"aws_access_key" default:"ABCD1234"`
	AWSSecret string `envconfig:"aws_secret_key" default:"ZYXW9876"`
	AWSRegion string `envconfig:"aws_region" default:"us-west-2"`
}

// AWS holds access details for AWS
type AWS struct {
	Region string
	Key    string
	Secret string
}

func main() {
	var err error
	logger := logrus.New()

	var spec specification
	if err = envconfig.Process("APP", &spec); err != nil {
		logger.Fatalln(err)
	}
	logger.Info(spec)

	aws := AWS{
		Region: spec.AWSRegion,
		Key:    spec.AWSKey,
		Secret: spec.AWSSecret,
	}
	s, err := NewServer(spec.Interval, aws)
	if err != nil {
		logger.Fatalln(err)
	}

	for {
		if err = s.ConnectDB(spec.DB); err != nil {
			logger.WithField("func", "main").Warnln("Problem connecting to DB", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	defer s.Close()

	logger.Info("Starting")
	err = http.ListenAndServe(spec.Bind, s.Router())
	if err != nil {
		logger.Errorln(err)
	}
}
