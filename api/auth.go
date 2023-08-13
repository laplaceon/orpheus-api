package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type CreateUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	CfToken  string `json:"cf_token"`
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

type authHeader struct {
	IDToken string `header:"Authorization"`
}

func checkAuthHeader(c *gin.Context) {
	h := authHeader{}

	if err := c.ShouldBindHeader(&h); err != nil {
		c.Abort()
		return
	}

	idTokenHeader := strings.Split(h.IDToken, "Bearer ")
	if len(idTokenHeader) < 2 {
		// err := apperrors.NewAuthorization("Must provide Authorization header with format `Bearer {token}`")

		// c.JSON(err.Status(), gin.H{
		// 	"error": err,
		// })
		c.Abort()
		return
	}

	fmt.Println(idTokenHeader[1])
}

func AuthRequired(c *gin.Context) {
	checkAuthHeader(c)
	c.Next()
}
