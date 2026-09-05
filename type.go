package main

import (
	"database/sql"
	"time"
)

type NullTime sql.NullTime

type Notification struct {
	UUID        string
	Title       string
	ScheduledAt time.Time
	SendedAt    NullTime
}
