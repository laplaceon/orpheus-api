package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type CreateUserPayload struct {
	Email   string `json:"email"`
	CfToken string `json:"cf_token"`
}

func CheckTurnstile(cfResponse string, httpClient *http.Client) (isSuccess bool, err error) {
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
