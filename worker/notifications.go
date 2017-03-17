package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"

	"upspin.io/path"
	"upspin.io/upspin"
)

// Run contains details of a previous run
type Run struct {
	Date       time.Time `db:"date"`
	Hash       string    `db:"final_hash"`
	Records    int64     `db:"record_count"`
	BadRecords int64     `db:"bad_record"`
}

// sendNotifications writes out a log line to an upspin file
// it'll tweet too if there are inconsistencies
func (w *Worker) sendNotifications() error {
	var err error

	// Get the last run
	var run Run
	if err := w.db.Get(&run, "SELECT `date`, `final_hash`, `record_count`, `bad_record` FROM `run` ORDER BY `date` DESC LIMIT 1"); err != nil {
		return err
	}

	// Get the number of bad records from the run before this
	// We'll bail if the number of bad records hasn't increased since then
	var prevBad int64
	if err := w.db.Get(&prevBad, "SELECT `bad_record` FROM `run` ORDER BY `date` DESC LIMIT 1, 1"); err != nil {
		return err
	}
	if run.BadRecords == 0 || run.BadRecords == prevBad {
		w.log.Infof("No notifications to send we have %d bad records and 0 new bad records\n", run.BadRecords)
		return nil
	}

	w.log.Errorln("We have %d new bad records, sending notifications", run.BadRecords-prevBad)

	// Save to an upspin directory
	filename := strings.Replace(run.Date.String(), " ", "_", -1)
	path := path.Join(upspin.PathName(w.upspinPath), filename)
	message := []byte(fmt.Sprintf("I've seen %d records and %d have bad hashes\n", run.Records, run.BadRecords))
	_, err = w.us.Put(path, message)
	if err != nil {
		w.log.Errorln("Damn, failed to save to upspin")
	}

	// Tweet about it
	if err = w.saveToTwitter(message); err != nil {
		w.log.Errorln("Uh oh, failed to save to twitter")
	}

	// Email people who have registered to receive notifications
	if err = w.emailUsers(string(message)); err != nil {
		w.log.Errorln("Uh oh, failed to email users")
	}

	return err
}

// saveToTwitter will use upspin2tweet.com to tweet
// It must only be used if there is an inconsistency found
func (w *Worker) saveToTwitter(tweet []byte) error {
	_, err := w.us.Put("upspin@upspin2tweet.com/KeyserverLog/tweet", tweet)
	return err
}

// emailUsers is used to email all registered users. This must only be used
// when an inconsistency in the log is found.
func (w *Worker) emailUsers(message string) error {
	var bccAddresses []*string
	if err := w.db.Select(&bccAddresses, "SELECT `address` FROM `email` WHERE `confirmed`=true"); err != nil {
		return fmt.Errorf("couldn't load email addresses %v", err)
	}
	w.sendEmail(
		"info@jn.gl", bccAddresses, "info@jn.gl",
		"Upspin Keyserver Log integrity problem",
		"The service running at https://logaudit.jn.gl has detected a problem with the integrity of the upspin Keyserver log at https://key.upspin.io/log: "+message,
	)
	return nil
}

// sendEmail sends an email using AWS SES
func (w *Worker) sendEmail(dest string, bcc []*string, from string, subject string, body string) error {
	awsSession := session.New(&aws.Config{
		Region:      aws.String(w.aws.Region),
		Credentials: credentials.NewStaticCredentials(w.aws.Key, w.aws.Secret, ""),
	})
	sesSession := ses.New(awsSession)
	sesEmailInput := &ses.SendEmailInput{
		Destination: &ses.Destination{
			BccAddresses: bcc,
			ToAddresses:  []*string{aws.String(dest)},
		},
		Source:           aws.String(from),
		ReplyToAddresses: []*string{aws.String(from)},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{Data: aws.String(body)},
			},
			Subject: &ses.Content{Data: aws.String(subject)},
		},
	}
	_, err := sesSession.SendEmail(sesEmailInput)
	return err
}
