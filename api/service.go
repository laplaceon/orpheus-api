package api

import (
	"database/sql"
	"log"
	"os"

	// blank import for side effect

	_ "github.com/go-sql-driver/mysql"
)

type Service struct {
	Database *sql.DB
}

func InitService() Service {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return Service{db}
}
