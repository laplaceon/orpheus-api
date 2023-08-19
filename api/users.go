package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"

	"github.com/alexedwards/argon2id"
)

func createUser(createUserPayload UserAuthPayload, db *sql.DB, httpClient *http.Client) (newId int64, err ClientError) {
	// map[error-codes:[timeout-or-duplicate] messages:[] success:false]
	// map[action: cdata: challenge_ts:2023-06-29T16:39:46.455Z error-codes:[] hostname:localhost metadata:map[interactive:false] success:true]

	validationErrors := []string{}

	if !govalidator.IsEmail(createUserPayload.Email) {
		validationErrors = append(validationErrors, "Email is invalid.")
	}

	if !govalidator.IsAlphanumeric(createUserPayload.Password) || len(createUserPayload.Password) < 6 {
		validationErrors = append(validationErrors, "Password must be at least 6 characters and alphanumeric.")
	}

	if len(validationErrors) > 0 {
		return 0, NewHttpError(nil, http.StatusBadRequest, strings.Join(validationErrors, " "))
	}

	cfSuccess, err := checkTurnstile(createUserPayload.CfToken, httpClient)

	if err != nil {
		return 0, NewHttpError(err, http.StatusInternalServerError, "Error checking cloudflare.")
	} else if !cfSuccess {
		return 0, NewHttpError(nil, http.StatusBadRequest, "Failed cloudflare check.")
	}

	row := db.QueryRow("SELECT EXISTS(SELECT id FROM users WHERE email = ?);", createUserPayload.Email)

	var exists bool
	if err = row.Scan(&exists); err != nil {
		return 0, NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	if exists {
		return 0, NewHttpError(err, http.StatusUnauthorized, "A user with this email already exists")
	}

	hash, err := argon2id.CreateHash(createUserPayload.Password, argon2id.DefaultParams)
	if err != nil {
		return 0, NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	fmt.Println(hash)

	insertStmt, err := db.Prepare("INSERT into users (email, password) VALUES (?, ?);")
	if err != nil {
		return 0, NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}
	defer insertStmt.Close()

	insertResult, err := insertStmt.Exec(createUserPayload.Email, hash)
	if err != nil {
		return 0, NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	newId, err = insertResult.LastInsertId()
	if err != nil {
		return 0, NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	return
}

func (s *Service) CreateUser(c *gin.Context) {
	var createUserPayload UserAuthPayload
	err := c.BindJSON(&createUserPayload)

	var newId int64
	if err == nil {
		newId, err = createUser(createUserPayload, s.db, s.httpClient)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": newId})
}

func (s *Service) GetUser(c *gin.Context) {
	var getUserPayload UserAuthPayload

	err := c.BindJSON(&getUserPayload)

	var token string
	if err == nil {
		token, err = getUser(getUserPayload, s.db, s.httpClient)
	}

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func getUser(getUserPayload UserAuthPayload, db *sql.DB, httpClient *http.Client) (tokenString string, err ClientError) {
	validationErrors := []string{}

	if !govalidator.IsEmail(getUserPayload.Email) {
		validationErrors = append(validationErrors, "Email is invalid.")
	}

	if !govalidator.IsAlphanumeric(getUserPayload.Password) || len(getUserPayload.Password) < 6 {
		validationErrors = append(validationErrors, "Password must be at least 6 characters and alphanumeric.")
	}

	if len(validationErrors) > 0 {
		return "", NewHttpError(nil, http.StatusBadRequest, strings.Join(validationErrors, " "))
	}

	cfSuccess, err := checkTurnstile(getUserPayload.CfToken, httpClient)

	if err != nil {
		return "", NewHttpError(err, http.StatusInternalServerError, "Error checking cloudflare.")
	} else if cfSuccess {
		return "", NewHttpError(nil, http.StatusBadRequest, "Failed cloudflare check.")
	}

	row := db.QueryRow("SELECT id, password FROM users WHERE email = ?;", getUserPayload.Email)
	if err = row.Err(); err != nil {
		return "", NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	var id int64
	var hash string
	if err = row.Scan(&id, &hash); err != nil {
		return "", NewHttpError(err, http.StatusBadRequest, "Either the username or password is incorrect.")
	}

	match, err := argon2id.ComparePasswordAndHash(getUserPayload.Password, hash)
	if err != nil {
		return "", NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	if !match {
		return "", NewHttpError(err, http.StatusUnauthorized, "Either the username or password is incorrect.")
	}

	return idToJwt(id)
}
