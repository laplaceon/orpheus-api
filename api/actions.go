package api

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-audio/wav"
	"github.com/vincent-petithory/dataurl"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/wagslane/go-rabbitmq"
)

type ActionRequest struct {
	UserId   int    `json:"user_id"`
	ActionId int    `json:"action_id"`
	Data     string `json:"data"`
}

type ActionRequestProcessable struct {
	HistoryId int    `msgpack:"history_id"`
	ActionId  int    `msgpack:"action_id"`
	Data      string `msgpack:"data"`
}

func (s *Service) GetActions(c *gin.Context) {
	actions := []ActionCost{}

	if err := s.db.Ping(); err != nil {
		log.Println(err)
		return
	}

	rows, err := s.db.Query(
		`SELECT costId as id, action_id, name, cost, length FROM actions JOIN (
			SELECT a.id as costId, cost, length, a.action_id FROM action_costs a
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
		action := ActionCost{}
		if err := rows.Scan(&action.Id, &action.ActionId, &action.Name, &action.Cost, &action.Length); err != nil {
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

func getAction(actionId int, db *sql.DB) (ActionCost, error) {
	row := db.QueryRow(
		`SELECT costId as id, action_id, name, cost, length FROM actions JOIN (
			SELECT a.id as costId, cost, length, a.action_id FROM action_costs a
			INNER JOIN (
				SELECT action_id, MAX(created_at) created_at
				FROM action_costs
				GROUP BY action_id
			) b ON a.action_id = b.action_id AND a.created_at = b.created_at
		) costs ON actions.id = costs.action_id 
        WHERE action_id = ?`, actionId)

	action := ActionCost{}
	err := row.Scan(&action.Id, &action.ActionId, &action.Name, &action.Cost, &action.Length)

	return action, err
}

func createActionRequest(actionRequest ActionRequest, db *sql.DB, pub *rabbitmq.Publisher) (historyId int64, err ClientError) {
	row := db.QueryRow("SELECT verified FROM users WHERE id = ?", actionRequest.UserId)

	var verified bool
	if err = row.Scan(&verified); err != nil {
		log.Println(err)
		return
	}

	if !verified {
		fmt.Println("Not verified")
		return
	}

	dUrl, err := dataurl.DecodeString(actionRequest.Data)

	if err != nil {
		log.Println(err)
		return
	}

	if dUrl.MediaType.ContentType() != "audio/wav" {
		fmt.Println("Incorrect content type")
		return
	}

	decoder := wav.NewDecoder(bytes.NewReader(dUrl.Data))
	dur, err := decoder.Duration()
	if err != nil {
		log.Println(err)
		return
	}

	userData, err := getUserWithId(actionRequest.UserId, db)
	if err != nil {
		log.Println(err)
		return
	}

	actionCost, err := getAction(actionRequest.ActionId, db)
	if err != nil {
		log.Println(err)
		return
	}

	estimatedCost := actionCost.Cost * dur.Seconds() / actionCost.Length

	if estimatedCost > userData.UsableCredits {
		fmt.Println("Not enough credits")
		return
	}

	insertStmt, err := db.Prepare("INSERT into history (user_id, plan_id, cost_id, input_size, status) VALUES (?, ?, ?, ?, 0);")
	if err != nil {
		return 0, NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}
	defer insertStmt.Close()

	insertResult, err := insertStmt.Exec(actionRequest.UserId, userData.PlanId, actionCost.Id, dur.Seconds())
	if err != nil {
		return 0, NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	historyId, err = insertResult.LastInsertId()
	if err != nil {
		return 0, NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	arB, err := msgpack.Marshal(ActionRequestProcessable{
		HistoryId: int(historyId),
		ActionId:  actionRequest.ActionId,
		Data:      actionRequest.Data,
	})

	err = pub.Publish(
		arB,
		[]string{"actions"},
	)

	if err != nil {
		log.Println(err)
		return
	}

	return historyId, err
}

func (s *Service) CreateActionRequest(c *gin.Context) {
	var actionRequest ActionRequest
	err := c.BindJSON(&actionRequest)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "There was an error processing the request"})
		return
	}

	historyId, err := createActionRequest(actionRequest, s.db, s.amqpPub)

	if err != nil {
		err := err.(HttpError)
		c.JSON(err.Status, gin.H{"error": err.Message})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"history_id": historyId})
}
