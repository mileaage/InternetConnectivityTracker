package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDatabase() error {
	db, err := sql.Open("sqlite3", "./trackdata.db")
	if err != nil {
		log.Fatal("Failed to open database")
	}

	defer db.Close()

	// test connections
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to database")
	return nil
}

func StoreDowntime() error {
	return nil
}
