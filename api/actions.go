package api

import (
	"github.com/gin-gonic/gin"
)

type ActionRequest struct {
	UserId int    `json:"user_id"`
	Data   string `json:"data"`
}

func (s *Service) CreateGenreTransferRequest(c *gin.Context) {

}
