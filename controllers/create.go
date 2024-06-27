package controllers

import (
	"arkis_test/stream"
	"net/http"

	"context"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func endpointCreateStream(c *gin.Context) {
	ctx := context.Background()

	var create struct {
		Name string
	}

	if err := c.ShouldBindJSON(&create); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if create.Name == "" {
		c.JSON(400, gin.H{"error": "Name is required"})
		log.WithField("name", string(create.Name)).Error("Name is required")
		return
	}

	success, inputName, outputName := stream.New(create.Name, ctx)
	if !success {
		c.JSON(400, gin.H{"error": "Stream already exists"})
		return
	}

	c.JSON(200, gin.H{"input": inputName, "output": outputName})
}
