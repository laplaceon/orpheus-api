package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	// blank import for side effect

	_ "github.com/go-sql-driver/mysql"
)

type Service struct {
	db *sql.DB
	// bgQueue    *rabbitmq.Conn
	httpClient *http.Client
}

func InitService() Service {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return Service{
		db,
		&http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
