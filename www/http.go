package main

import (
	"html/template"
	"net/http"
)

// ErrorResponse is used when a json object needs to be returned with just an error
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

const errorPage = `
<html>
    <head>
        <title>Upspin Keyserver Log Audit</title>
    </head>
    <body>
        <h1>Oops! - {{ .Code }} - {{ .Message }}</h1>
    </body>
</html>
`

// errorResponse is a wrapper around templateResponse for returning errors
func errorResponse(w http.ResponseWriter, message string, status int) error {
	w.WriteHeader(status)
	t, _ := template.New("errorPage").Parse(errorPage)
	t.Execute(w, ErrorResponse{Error: message, Code: status})
	return nil
}

// templateResponse returns the template populated with the struct
func templateResponse(w http.ResponseWriter, filename string, v interface{}, status int) error {
	t, err := template.ParseFiles(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		t, _ := template.New("errorPage").Parse(errorPage)
		t.Execute(w, ErrorResponse{Error: "Failed to load tempalte", Code: http.StatusInternalServerError})
		return err
	}
	w.WriteHeader(status)
	t.Execute(w, v)
	return nil
}
