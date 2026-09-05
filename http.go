package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	conf Config

	db *sql.DB
}

type ValidationErr struct {
	field   string
	message string
	err     error
}

func (e *ValidationErr) Error() string {
	return fmt.Sprintf("%s field is invalid: %s", e.field, e.message)
}

func (e *ValidationErr) Is(other *ValidationErr) bool {
	return e.field == other.field && e.err.Error() == other.err.Error() && e.message == other.message
}

func (e *ValidationErr) Unwrap() error {
	return e.err
}

func newHandler(db *sql.DB, conf Config) *Handler {
	h := &Handler{db: db, conf: conf}

	return h
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func accepted(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	response := map[string]string{
		"status": "accepted",
	}

	json.NewEncoder(w).Encode(response)
}

type CreateNotificationRequest struct {
	UUID        string `json:"uuid"`
	Title       string `json:"title"`
	ScheduledAt string `json:"scheduled_at"`
}

func (r *CreateNotificationRequest) ToNotification() (Notification, error) {
	err := ValidateUUID(r.UUID)
	if err != nil {
		return Notification{}, &ValidationErr{
			field:   "uuid",
			err:     err,
			message: "invalid format for uuidv4",
		}
	}

	if strings.TrimSpace(r.Title) == "" {
		return Notification{}, &ValidationErr{
			field:   "title",
			message: "title must be not empty",
		}
	}

	scheduledAt, err := time.Parse(time.DateTime, r.ScheduledAt)
	if err != nil {
		return Notification{}, &ValidationErr{
			field:   "scheduled_at",
			err:     err,
			message: fmt.Sprintf("invalid datetime format. expected: %s", time.DateTime),
		}
	}

	return Notification{
		UUID:        r.UUID,
		Title:       r.Title,
		ScheduledAt: scheduledAt,
		SendedAt:    NullTime{Valid: false},
	}, nil
}

func (h *Handler) createNotificationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	notificationRequest := &CreateNotificationRequest{}

	err := json.NewDecoder(r.Body).Decode(notificationRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %s", err.Error()), http.StatusBadRequest)

		return
	}

	notification, err := notificationRequest.ToNotification()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		internalServerError(w)

		return
	}
	defer tx.Rollback()

	err = SaveNotification(ctx, tx, notification)
	if err != nil {
		if errors.Is(err, NotificationAlreadyExistsErr) {
			http.Error(w, "notification with given uuid already exists", http.StatusBadRequest)

			return
		}

		slog.Error("failed to save notification", "err", err.Error())
		internalServerError(w)

		return
	}

	err = tx.Commit()
	if err != nil {
		internalServerError(w)

		return
	}

	accepted(w)
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
