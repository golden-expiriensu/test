package controllers

import (
	"arkis_test/stream"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Message": "pong"})
}

func getStreams(c *gin.Context) {

	var inputs []string
	for k := range stream.InputStreams {
		inputs = append(inputs, k)
	}

	var outputs []string
	for k := range stream.OutputStreams {
		outputs = append(outputs, k)
	}

	c.JSON(http.StatusOK, gin.H{"input": inputs, "output": outputs})
}

func StartGinServer() {
	router := gin.Default()
	router.GET("/", getStreams)
	router.GET("/ping", getHealth)
	router.POST("/create", endpointCreateStream)
	router.PUT("/publish/:input", endpointPublishToStream)
	router.GET("/consume/:output", endpointConsumeFromStream)
	router.Run()
}
