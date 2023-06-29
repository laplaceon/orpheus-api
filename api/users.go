package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func CreateUser(createUser CreateUserPayload, db *sql.DB, httpClient *http.Client) (newId int, err error) {
	// map[error-codes:[timeout-or-duplicate] messages:[] success:false]
	// map[action: cdata: challenge_ts:2023-06-29T16:39:46.455Z error-codes:[] hostname:localhost metadata:map[interactive:false] success:true]

	isSuccess, err := CheckTurnstile(createUser.CfToken, httpClient)

	if err != nil {
		log.Println(err)
		return
	} else if !isSuccess {
		return
	}

	row := db.QueryRow("SELECT id FROM users WHERE email = ?", createUser.Email)

	var id int
	if err = row.Scan(&id); err != nil {
		log.Println(err)
		return
	}

	insertStmt, err := db.Prepare("INSERT into users (email) (?)")
	if err != nil {
		log.Println(err)
		return
	}
	defer insertStmt.Close()

	if _, err = insertStmt.Exec(createUser.Email); err != nil {
		log.Println(err)
		return
	}

	return newId, err
}

func (s *Service) CreateUser(c *gin.Context) {
	var createUser CreateUserPayload
	err := c.BindJSON(&createUser)

	var newId int
	if err == nil {
		newId, err = CreateUser(createUser, s.db, s.httpClient)
	}

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": newId})
}

func (s *Service) GetUser(c *gin.Context) {
	email := c.Param("email")

	token, err := GetUser(email, s.db)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func GetUser(email string, db *sql.DB) (tokenString string, err error) {
	row := db.QueryRow("SELECT id FROM users WHERE email = ?", email)
	if err = row.Err(); err != nil {
		log.Println(err)
		return
	}

	var id int
	if err = row.Scan(&id); err != nil {
		log.Println(err)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": id,
	})

	tokenString, err = token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return
}
