package main

import (
	"database/sql"
	"fmt"
	"net/http"
)

type Handler struct {
	conf Config

	db *sql.DB
}

func newHandler(db *sql.DB, conf Config) *Handler {
	h := &Handler{db: db, conf: conf}

	return h
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func (h *Handler) createNotificationHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement me.
	fmt.Fprintf(w, "Notification was created")
}

func (h *Handler) basicAuth(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)

			return
		}

		if username != h.conf.Username || password != h.conf.Password {
			http.Error(w, "Invalid username or/and password provided", http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(w, r)
	}
}

type errorResponse struct {
	ErrorMessage string `json:"error_message"`
}
