package api

import (
	"database/sql"

	"log"

	// blank import for side effect

	_ "github.com/go-sql-driver/mysql"
)

const dbFileName = "root:leo869636@tcp(localhost:3306)/orpheus?parseTime=true"

type Service struct {
	Database *sql.DB
}

func InitService() Service {
	db, err := sql.Open("mysql", dbFileName)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return Service{db}
}
