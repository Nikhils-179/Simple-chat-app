package db

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoadTestHandler(c *gin.Context) {
	// Simulate chat message storage
	err := StoreCachedMessage(ctx, "room1", "Test message")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store message"})
		return
	}

	// Simulate chat message retrieval
	_, err = GetCachedHistory(ctx, "room1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chat history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Load test successful"})
}

//command
//hey -n 10 -c 10 http://localhost:8080/load-test
