package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"

	"net/url"
	"log"

	"github.com/vitalii-q/selena-users-service/internal/services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

type OAuthHandler struct {
	UserService *services.UserServiceImpl
}

func GetAuthorize(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	state := c.Query("state")

	// Для MVP — логируем clientID
	log.Printf("Authorize request for client_id: %s", clientID)

	// TODO: Validate client_id, redirect_uri

	// Сгенерировать временный код (в будущем — сохранить в БД)
	authCode := "sample_auth_code_abc123"

	// Редирект на redirect_uri с кодом
	redirect, _ := url.Parse(redirectURI)
	q := redirect.Query()
	q.Set("code", authCode)
	q.Set("state", state)
	redirect.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, redirect.String())
}

func (h *OAuthHandler) PostToken(c *gin.Context) {
	log.Println("Received request to /oauth2/token")
	
	grantType := c.PostForm("grant_type")
	email := c.PostForm("email")
	password := c.PostForm("password")

	if grantType != "password" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
		return
	}

	user, err := h.UserService.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}

	token, err := utils.GenerateJWT(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_generation_failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "bearer",
		"expires_in":   3600,
	})
}
