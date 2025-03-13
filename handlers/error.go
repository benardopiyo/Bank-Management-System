package handlers

import (
	"net/http"
	"text/template"
)

func ErrorPage(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}

	errorData := struct {
		StatusCode int
		Message    string
	}{
		StatusCode: statusCode,
		Message:    message,
	}

	tmpl := template.Must(template.ParseFiles("templates/error.html"))
	tmpl.Execute(w, errorData)
}

func ErrorPageTrans(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}

	errorData := struct {
		StatusCode int
		Message    string
	}{
		StatusCode: statusCode,
		Message:    message,
	}

	tmpl := template.Must(template.ParseFiles("templates/errorTrans.html"))
	tmpl.Execute(w, errorData)
}