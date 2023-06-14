package api

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (s *Service) CreateUser(c *gin.Context) {
	newUser := User{}
	if err := c.BindJSON(&newUser); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.db.Ping(); err != nil {
		log.Println(err)
		return
	}

	rowStmt, err := s.db.Prepare("SELECT MAX(id) AS id FROM orders")
	if err != nil {
		log.Println(err)
		return
	}
	defer rowStmt.Close()

	// get the last order id

	var id sql.NullInt32
	if err = rowStmt.QueryRow().Scan(&id); err != nil {
		log.Println(err)
		return
	}

	var newID int

	if id.Valid {
		newID = int(id.Int32) + 1
	} else {
		newID = 1
	}

	// write each order line as a row

	insertStmt, err := s.db.Prepare("INSERT INTO orders (id, product_id, quantity) values (?, ?, ?)")
	if err != nil {
		log.Println(err)
		return
	}
	defer insertStmt.Close()

	// var itemCount int
	// for _, line := range newUser.Lines {
	// 	itemCount += line.Quantity
	// 	if _, err = insertStmt.Exec(newID, line.ProductID, line.Quantity); err != nil {
	// 		log.Println(err)
	// 	}
	// }

	// log.Printf("Order #%d (%d items) added\n", newID, itemCount)

	if err != nil || newID == 0 {
		if newID == 0 {
			err = errors.New("unable to get new user id")
		}
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// return the new order id
	c.JSON(http.StatusCreated, gin.H{"id": newID})
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
