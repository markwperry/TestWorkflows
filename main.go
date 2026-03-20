package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	version    = "MISSING"
	gitHash    = "MISSING"
	buildStamp = "MISSING"
	branch     = "MISSING"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := setupRouter()
	r.Run(":" + port)
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/helloworld", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    version,
			"gitHash":    gitHash,
			"buildStamp": buildStamp,
			"branch":     branch,
		})
	})

	return r
}
