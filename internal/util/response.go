package util

import "github.com/gin-gonic/gin"

func RespondWithJSON(c *gin.Context, status int, payload interface{}) {
	c.JSON(status, payload)
}

func RespondWithError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
