package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type CreateUserPayload struct {
	Email   string `json:"email"`
	CfToken string `json:"cf_token"`
}

func idToJwt(id int64) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": id,
	})

	tokenString, err = token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return
}

func checkTurnstile(cfResponse string, httpClient *http.Client) (isSuccess bool, err error) {
	body := gin.H{
		"response": cfResponse,
		"secret":   os.Getenv("TURNSTILE_SECRET"),
	}

	bodyStr, err := json.Marshal(body)

	res, err := httpClient.Post("https://challenges.cloudflare.com/turnstile/v0/siteverify", "application/json", bytes.NewBuffer(bodyStr))
	if err != nil {
		return
	}

	defer res.Body.Close()

	var outcome map[string]interface{}
	json.NewDecoder(res.Body).Decode(&outcome)

	return outcome["success"].(bool), err
}

func AuthRequired(c *gin.Context) {
	c.Next()
}
