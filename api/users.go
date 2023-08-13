package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alexedwards/argon2id"
)

func createUser(createUserPayload CreateUserPayload, db *sql.DB, httpClient *http.Client) (tokenString string, err error) {
	// map[error-codes:[timeout-or-duplicate] messages:[] success:false]
	// map[action: cdata: challenge_ts:2023-06-29T16:39:46.455Z error-codes:[] hostname:localhost metadata:map[interactive:false] success:true]

	isSuccess, err := checkTurnstile(createUserPayload.CfToken, httpClient)

	if err != nil {
		log.Println(err)
		return
	} else if !isSuccess {
		return
	}

	row := db.QueryRow("SELECT EXISTS(SELECT id FROM users WHERE email = ?);", createUserPayload.Email)

	var exists bool
	if err = row.Scan(&exists); err != nil {
		log.Println(err)
		return
	}

	if exists {
		err = errors.New("User already exists")
		log.Println(err)
		return
	}

	hash, err := argon2id.CreateHash(createUserPayload.Password, argon2id.DefaultParams)
	if err != nil {
		log.Println(err)
	}

	insertStmt, err := db.Prepare("INSERT into users (email, password) VALUES (?, ?);")
	if err != nil {
		log.Println(err)
		return
	}
	defer insertStmt.Close()

	insertResult, err := insertStmt.Exec(createUserPayload.Email, hash)
	if err != nil {
		log.Println(err)
		return
	}

	newId, err := insertResult.LastInsertId()

	return idToJwt(newId)
}

func (s *Service) CreateUser(c *gin.Context) {
	var createUserPayload CreateUserPayload
	err := c.BindJSON(&createUserPayload)

	var tokenString string
	if err == nil {
		tokenString, err = createUser(createUserPayload, s.db, s.httpClient)
	}

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": tokenString})
}

func (s *Service) GetUser(c *gin.Context) {
	email := c.Param("email")

	token, err := getUser(email, "", s.db)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func getUser(email string, password string, db *sql.DB) (tokenString string, err error) {
	row := db.QueryRow("SELECT id FROM users WHERE email = ?;", email)
	if err = row.Err(); err != nil {
		log.Println(err)
		return
	}

	var id int64
	if err = row.Scan(&id); err != nil {
		log.Println(err)
		return
	}

	match, err := argon2id.ComparePasswordAndHash(password, "asdsad")
	if err != nil {
		log.Println(err)
	}

	fmt.Println(match)

	return idToJwt(id)
}
