package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LoadProductsFromDatabase load product list from DB
func (s *Service) GetHistory(c *gin.Context) {
	history := []HistoryItem{}

	userId := c.Param("id")

	db := s.Database

	if err := db.Ping(); err != nil {
		log.Println(err)
		return
	}

	rows, err := db.Query("SELECT * FROM history WHERE user_id = ? ORDER BY created_at DESC", userId)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		h := HistoryItem{}
		if err := rows.Scan(&h.Id, &h.UserId, &h.ActionId, &h.InputSize, &h.CreatedAt); err != nil {
			log.Println(err)
			continue
		}
		history = append(history, h)
	}

	log.Printf("Queried %d history items", len(history))

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// return JSON
	c.JSON(http.StatusOK, history)
}
