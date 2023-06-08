package api

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// OrderService is the POST service to add new order
func (s *Service) CreateUser(c *gin.Context) {
	newUser := User{}
	if err := c.BindJSON(&newUser); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	db := s.Database

	if err := db.Ping(); err != nil {
		log.Println(err)
		return
	}

	rowStmt, err := db.Prepare("SELECT MAX(id) AS id FROM orders")
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

	insertStmt, err := db.Prepare("INSERT INTO orders (id, product_id, quantity) values (?, ?, ?)")
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
