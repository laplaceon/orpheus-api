package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/wagslane/go-rabbitmq"
)

type ActionRequest struct {
	UserId int    `json:"user_id"`
	Data   string `json:"data"`
}

func createGenreTransferRequest(actionRequest ActionRequest, pub *rabbitmq.Publisher) (err error) {
	arB, err := msgpack.Marshal(actionRequest)

	if err != nil {
		log.Println(err)
		return
	}

	err = pub.Publish(
		arB,
		[]string{"actions"},
		rabbitmq.WithPublishOptionsExchange("amq.direct"),
	)

	return err
}

func (s *Service) CreateGenreTransferRequest(c *gin.Context) {
	var actionRequest ActionRequest
	err := c.BindJSON(&actionRequest)

	if err != nil {
		log.Println(err)
	}

	err = createGenreTransferRequest(actionRequest, s.amqpPub)

	if err != nil {
		log.Println(err)
	}

	c.JSON(200, gin.H{"success": true})
}
