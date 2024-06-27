package controllers

import (
	"arkis_test/stream"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type streamResponse struct {
	Result string `json:"result"`
}

func endpointConsumeFromStream(c *gin.Context) {
	output := c.Param("output")

	if stream.OutputStreams[output] == nil {
		c.JSON(400, gin.H{"error": "Stream does not exist"})
		return
	}

	log.Info("Consume from stream ", output)

	delivery := stream.OutputStreams[output]
	if delivery == nil {
		c.JSON(500, gin.H{"error": "Stream is empty"})
		log.Error("Stream is empty")
		return
	}

	log.Info("Waiting message from stream ", output)

	var resp streamResponse
	var outMsg = make([]streamResponse, 0)
	timer := time.NewTimer(500 * time.Millisecond)

OuterLoop:
	for {
		select {
		case message := <-delivery:
			if err := json.Unmarshal(message.Body, &resp); err == nil {
				outMsg = append(outMsg, resp)
			}

		case <-timer.C:
			log.Info("Breaking Outerloop ")
			break OuterLoop
		}
	}

	c.JSON(200, outMsg)
}
