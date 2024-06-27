package controllers

import (
	"arkis_test/stream"
	"net/http"

	"context"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func endpointPublishToStream(c *gin.Context) {
	ctx := context.Background()
	input := c.Param("input")

	var publish struct {
		Message string
	}

	if err := c.ShouldBindJSON(&publish); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if publish.Message == "" {
		c.JSON(400, gin.H{"error": "Message is required"})
		log.WithField("message", string(publish.Message)).Error("Message is required")
		return
	}

	if stream.InputStreams[input] == nil {
		c.JSON(400, gin.H{"error": "Stream does not exist"})
		return
	}

	err := stream.InputStreams[input].Publish(ctx, []byte(publish.Message))
	if err != nil {
		c.JSON(500, gin.H{"error": "Cannot publish the message"})
		log.WithError(err).Error("Cannot publish the message")
		return
	}

	c.JSON(200, gin.H{"message": "Message published"})
}
