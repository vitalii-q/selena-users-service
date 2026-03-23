package router

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func protected(c *gin.Context) {
	logrus.Info("Protected check request")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// getPublicIPv4 — IMDSv2 support for EC2 public IPv4
func getPublicIPv4() string {
	client := &http.Client{}

	// 1. IMDSv2 token
	tokenReq, err := http.NewRequest("PUT", "http://169.254.169.254/latest/api/token", nil)
	if err != nil {
		logrus.Errorf("IMDSv2 token request build failed: %v", err)
		return ""
	}
	tokenReq.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")

	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		logrus.Errorf("IMDSv2 token request failed: %v", err)
		return ""
	}
	defer tokenResp.Body.Close()

	token, err := json.Marshal(tokenResp.Body)
	if err != nil {
		logrus.Errorf("IMDSv2 token read failed: %v", err)
		return ""
	}

	// 2. Public IPv4 request
	metaReq, err := http.NewRequest("GET", "http://169.254.169.254/latest/meta-data/public-ipv4", nil)
	if err != nil {
		logrus.Errorf("Public IPv4 request build failed: %v", err)
		return ""
	}
	metaReq.Header.Set("X-aws-ec2-metadata-token", string(token))

	metaResp, err := client.Do(metaReq)
	if err != nil {
		logrus.Errorf("Public IPv4 request failed: %v", err)
		return ""
	}
	defer metaResp.Body.Close()

	body, err := json.Marshal(metaResp.Body)
	if err != nil {
		logrus.Errorf("Public IPv4 read failed: %v", err)
		return ""
	}

	return string(body)
}