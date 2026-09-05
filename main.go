package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

/*
   MVP user paths:
   - Client send notification create request with http api
   - This request protected with basic auth, just for case someone find out my open URL
   - Deamon once in 10 seconds reads not sended notificatons and send it to specific telegram chat ID with specific bot token, both defined as OS.env
*/

const appDir = ".notifier"
const appDBFile = "notifier.db"

type Config struct {
	Username string
	Password string
}

func initConfig() (Config, error) {
	username := os.Getenv("NOTIFIER_USERNAME")
	if username == "" {
		return Config{}, errors.New("NOTIFIER_USERNAME env is not set")
	}

	password := os.Getenv("NOTIFIER_PASSWORD")
	if password == "" {
		return Config{}, errors.New("NOTIFIER_PASSWORD env is not set")
	}

	return Config{
		Username: username,
		Password: password,
	}, nil
}

func main() {
	conf, err := initConfig()

	if err != nil {
		panic(err)
	}

	db, err := openAppDB()
	if err != nil {
		panic(err)
	}

	h := newHandler(db, conf)
	registerRoutes(h)

	slog.Info("running http server", "host:port", ":8090")
	err = http.ListenAndServe(":8090", nil) // TODO: move parameter to config
	if err != nil {
		panic(err)
	}
}

func registerRoutes(h *Handler) {
	http.Handle("POST /notification/{$}", h.basicAuth(http.HandlerFunc(h.createNotificationHandler)))
}

func openAppDB() (*sql.DB, error) {
	appDir, err := createAppDirIfNotExists()
	if err != nil {
		return nil, fmt.Errorf("failed to create application directory: %w", err)
	}

	db, err := sql.Open("sqlite3", filepath.Join(appDir, appDBFile))
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	err = CreateSchema(context.TODO(), db)
	if err != nil {
		return nil, fmt.Errorf("failed to create database schema: %w", err)
	}

	return db, nil
}

func createAppDirIfNotExists() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(homeDir, appDir)
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}

	return path, nil
}
