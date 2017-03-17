package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"html/template"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/gorilla/mux"
)

// Signup is used to populate the response after processing the signup form
type Signup struct {
	Message string
}

// VerificationEmail is used to populate the verification email sent to the user
type VerificationEmail struct {
	Address string
	Code    string
}

// generateToken creates a brand new random token for email verification
func generateToken() string {
	dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, 64)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

// PostSignupHandler registers a new user
func (s *Server) PostSignupHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.log.Debugln(r.Method, r.URL.Path, r.RemoteAddr)

		r.ParseForm()
		email := r.FormValue("email")
		code := generateToken()

		_, err := s.db.Exec("INSERT INTO `email` (`address`, `code`) VALUES (?, ?)", email, code)
		if err != nil {
			failed := Signup{Message: fmt.Sprintf("Failed to register your email address %s - %v", email, err)}
			templateResponse(w, "template/signup.html", failed, http.StatusBadRequest)
			return
		}

		t, err := template.ParseFiles("template/emailverification.html")
		if err != nil {
			failed := Signup{Message: fmt.Sprintf("Failed to generate verification email to your email address %s - %v", email, err)}
			templateResponse(w, "template/signup.html", failed, http.StatusBadRequest)
			return
		}
		var bodyBuffer bytes.Buffer
		t.Execute(&bodyBuffer, VerificationEmail{Address: email, Code: code})
		err = s.sendEmail(email, "upspin@jn.gl", "Upspin.jn.gl email verification", bodyBuffer.String())
		if err != nil {
			failed := Signup{Message: fmt.Sprintf("Failed to send verification email to your email address %s - %v", email, err)}
			templateResponse(w, "template/signup.html", failed, http.StatusBadRequest)
			return
		}

		success := Signup{Message: fmt.Sprintf("Your email address %s has been registered, a verification email has been sent.", email)}
		templateResponse(w, "template/signup.html", success, http.StatusOK)
	})
}

// GetSignupHandler completes the registration process
func (s *Server) GetSignupHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.log.Debugln(r.Method, r.URL.Path, r.RemoteAddr)
		vars := mux.Vars(r)
		code := vars["code"]
		res, err := s.db.Exec("UPDATE `email` SET `confirmed`=true WHERE CODE=?", code)
		if err != nil {
			failed := Signup{Message: fmt.Sprintf("Failed to verify your email address - %v", err)}
			templateResponse(w, "template/signup.html", failed, http.StatusBadRequest)
			return
		}
		rows, err := res.RowsAffected()
		if rows == 0 || err != nil {
			failed := Signup{Message: "The email address is either already verified or your code is incorrect"}
			templateResponse(w, "template/signup.html", failed, http.StatusBadRequest)
			return
		}
		success := Signup{Message: fmt.Sprintf("Your email address been verified.")}
		templateResponse(w, "template/signup.html", success, http.StatusOK)
	})
}

// sendEmail sends an email using AWS SES
func (s *Server) sendEmail(dest string, from string, subject string, body string) error {
	awsSession := session.New(&aws.Config{
		Region:      aws.String(s.aws.Region),
		Credentials: credentials.NewStaticCredentials(s.aws.Key, s.aws.Secret, ""),
	})
	sesSession := ses.New(awsSession)
	sesEmailInput := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{aws.String(dest)},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Data: aws.String(body)},
			},
			Subject: &ses.Content{
				Data: aws.String(subject),
			},
		},
		Source: aws.String(from),
		ReplyToAddresses: []*string{
			aws.String(from),
		},
	}
	_, err := sesSession.SendEmail(sesEmailInput)
	return err
}
