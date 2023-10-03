package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getAllHistory(userId int, db *sql.DB) (history []HistoryItem, err ClientError) {
	rows, err := db.Query(
		`SELECT history.id, user_id, plan_id, action_id, name, cost * (input_size / length) as cost, status, history.created_at FROM history 
			JOIN action_costs ON history.cost_id = action_costs.id
			JOIN actions ON action_costs.action_id = actions.id
			WHERE user_id = ?
			ORDER BY created_at DESC;`, userId)

	if err != nil {
		return nil, NewHttpError(err, http.StatusInternalServerError, "There was an error with the server.")
	}
	defer rows.Close()

	history = []HistoryItem{}
	for rows.Next() {
		h := HistoryItem{}
		if err := rows.Scan(&h.Id, &h.UserId, &h.PlanId, &h.ActionId, &h.ActionName, &h.Cost, &h.Status, &h.CreatedAt); err != nil {
			return nil, NewHttpError(err, http.StatusInternalServerError, "There was an error with the server.")
		}
		history = append(history, h)
	}

	return
}

func (s *Service) GetAllHistory(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))

	var history []HistoryItem
	if err == nil {
		history, err = getAllHistory(userId, s.db)
	} else {
		err = NewHttpError(err, http.StatusBadRequest, "Incorrect user id")
	}

	if err != nil {
		c.JSON(err.(HttpError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (s *Service) GetHistoryItem(c *gin.Context) {
	h := HistoryItem{}

	id := c.Param("id")

	row := s.db.QueryRow(
		`SELECT history.id, user_id, plan_id, action_id, name, cost * (input_size / length) as cost, status, history.created_at FROM history 
			JOIN action_costs ON history.cost_id = action_costs.id
			JOIN actions ON action_costs.action_id = actions.id
			WHERE history.id = ?;`, id)

	if err := row.Scan(&h.Id, &h.UserId, &h.PlanId, &h.ActionId, &h.ActionName, &h.Cost, &h.Status, &h.CreatedAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error with the server"})
		return
	}

	c.JSON(http.StatusOK, h)
}

func (s *Service) GetGeneratedFromHistory(c *gin.Context) {
	generatedItems := []GeneratedItem{}

	id := c.Param("id")

	rows, err := s.db.Query(`SELECT * FROM gen_items WHERE history_id = ?;`, id)

	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		g := GeneratedItem{}
		if err := rows.Scan(&g.Id, &g.HistoryId, &g.Url, &g.CreatedAt); err != nil {
			log.Println(err)
			continue
		}
		generatedItems = append(generatedItems, g)
	}

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, generatedItems)

}
