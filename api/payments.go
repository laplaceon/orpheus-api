package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v75"
)

func (s *Service) ProcessPaymentFromStripe(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read webhook body"})
		return
	}

	event := stripe.Event{}

	if err := json.Unmarshal(payload, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse webhook payload"})
		return
	}

	switch event.Type {
	// case "charge.succeeded":
	case "checkout.session.completed":
		var checkoutSession stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &checkoutSession)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse checkout session"})
			return
		}

		if checkoutSession.PaymentStatus == "paid" && checkoutSession.Status == "complete" {
			userId, err := strconv.Atoi(checkoutSession.ClientReferenceID)

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client reference id"})
				return
			}

			if product_key, ok := checkoutSession.Metadata["product_key"]; ok {
				switch product_key {
				case "credits":
					err = updatePurchasedCredits(userId, checkoutSession.AmountSubtotal, checkoutSession.ID, s.db)
				case "plan_basic":
					err = updatePlan(userId, 2, checkoutSession.ID, s.db)
				case "plan_artist":
					err = updatePlan(userId, 3, checkoutSession.ID, s.db)
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to access product key"})
				return
			}
		}
	}

	if err != nil {
		err := err.(HttpError)
		c.JSON(err.Status, gin.H{"status": err.Message})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
