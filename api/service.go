package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	// blank import for side effect

	_ "github.com/go-sql-driver/mysql"
	"github.com/wagslane/go-rabbitmq"
)

type Service struct {
	db         *sql.DB
	amqpConn   *rabbitmq.Conn
	amqpPub    *rabbitmq.Publisher
	httpClient *http.Client
}

func (s *Service) Close() {
	s.db.Close()
	s.amqpPub.Close()
	s.amqpConn.Close()
}

func InitService() Service {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	amqpConn, err := rabbitmq.NewConn(
		os.Getenv("RABBIT_CONN"),
		rabbitmq.WithConnectionOptionsLogging,
	)

	if err != nil {
		log.Fatal(err)
	}

	amqpPub, err := rabbitmq.NewPublisher(
		amqpConn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName("amq.direct"),
	)

	if err != nil {
		log.Fatal(err)
	}

	return Service{
		db,
		amqpConn,
		amqpPub,
		&http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
