package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Service) GetActions(c *gin.Context) {
	actions := []Action{}

	if err := s.db.Ping(); err != nil {
		log.Println(err)
		return
	}

	rows, err := s.db.Query(
		`SELECT action_id as id, name, cost, length FROM actions JOIN (
			SELECT cost, length, a.action_id FROM action_costs a
			INNER JOIN (
				SELECT action_id, MAX(created_at) created_at
				FROM action_costs
				GROUP BY action_id
			) b ON a.action_id = b.action_id AND a.created_at = b.created_at
		) costs ON actions.id = costs.action_id`)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		action := Action{}
		if err := rows.Scan(&action.Id, &action.Name, &action.Cost, &action.Length); err != nil {
			log.Println(err)
			continue
		}
		actions = append(actions, action)
	}

	log.Printf("Queried %d action items", len(actions))

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// return JSON
	c.JSON(http.StatusOK, actions)
}

// LoadProductsFromDatabase load product list from DB
func (s *Service) GetHistory(c *gin.Context) {
	history := []HistoryItem{}

	userId := c.Param("id")

	if err := s.db.Ping(); err != nil {
		log.Println(err)
		return
	}

	rows, err := s.db.Query(
		`SELECT history.id, user_id, action_id, cost * (input_size / length) as cost, history.created_at FROM history 
			JOIN action_costs ON history.cost_id = action_costs.id
			JOIN actions ON action_costs.action_id = actions.id
			WHERE user_id = ?
			ORDER BY created_at DESC;`, userId)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		h := HistoryItem{}
		if err := rows.Scan(&h.Id, &h.UserId, &h.ActionId, &h.Cost, &h.CreatedAt); err != nil {
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
