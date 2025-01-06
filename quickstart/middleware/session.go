package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"net/http"
)

func SessionMiddleware(hankoUrl string) echo.MiddlewareFunc {
	client := http.Client{}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("hanko")
			if err == http.ErrNoCookie {
				return c.Redirect(http.StatusTemporaryRedirect, "/unauthorized")
			}
			if err != nil {
				return err
			}

			requestBody := CheckSessionRequest{SessionToken: cookie.Value}

			bodyJson, err := json.Marshal(requestBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}
			httpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/sessions/validate", hankoUrl), bytes.NewReader(bodyJson))
			if err != nil {
				return err
			}
			httpReq.Header.Set("Content-Type", "application/json")

			response, err := client.Do(httpReq)
			if err != nil {
				return err
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to get session response: %d", response.StatusCode)
			}

			responseBytes, err := io.ReadAll(response.Body)
			if err != nil {
				return err
			}

			var sessionResponse CheckSessionResponse
			err = json.Unmarshal(responseBytes, &sessionResponse)
			if err != nil {
				return err
			}

			if !sessionResponse.IsValid {
				return c.Redirect(http.StatusTemporaryRedirect, "/unauthorized")
			}
			log.Printf("session for user '%s' verified successfully", sessionResponse.UserID)
			c.Set("token", cookie.Value)
			c.Set("user", sessionResponse.UserID)

			return next(c)
		}
	}
}

type CheckSessionRequest struct {
	SessionToken string `json:"session_token"`
}

type CheckSessionResponse struct {
	IsValid        bool   `json:"is_valid"`
	ExpirationTime string `json:"expiration_time"`
	UserID         string `json:"user_id"`
}
