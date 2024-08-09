package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"
	_ "github.com/tursodatabase/go-libsql"
)

var db *sql.DB

func Get() (*sql.DB, error) {
	if db != nil {
		return db, nil
	}

	dbName := "file:./local.db"
	sqlInstance, err := sql.Open("libsql", dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s", err)
		return nil, err
	}
	db = sqlInstance
	return sqlInstance, nil
}

func GetRandomId() string {
	return uuid.New().String()
}
