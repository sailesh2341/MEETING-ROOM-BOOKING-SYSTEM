package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func Open(databaseURL string) (*sql.DB, error) {
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	database.SetMaxOpenConns(20)
	database.SetMaxIdleConns(10)
	database.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = database.PingContext(ctx)
	if err != nil {
		database.Close()
		return nil, err
	}

	return database, nil
}

func Close(database *sql.DB) error {
	if database == nil {
		return nil
	}
	return database.Close()
}
