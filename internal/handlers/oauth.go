package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"net/url"
	"log"
)

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

func PostToken(c *gin.Context) {
	grantType := c.PostForm("grant_type")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")
	code := c.PostForm("code")
	//redirectURI := c.PostForm("redirect_uri")

	// Поддерживаем только authorization_code
	if grantType != "authorization_code" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
		return
	}

	// TODO: Валидация client_id/client_secret
	if clientID != "my_client_id" || clientSecret != "my_client_secret" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
		return
	}

	// TODO: Проверка кода (в реальности: проверка в Redis/DB)
	if code != "sample_auth_code_abc123" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	// Генерация токена (в реальности: JWT или UUID)
	accessToken := "access_token_xyz987"
	refreshToken := "refresh_token_uvw456" // опционально

	// TODO: Сохранять access_token и привязку к user

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"token_type":    "bearer",
		"expires_in":    3600,
		"refresh_token": refreshToken,
	})
}
