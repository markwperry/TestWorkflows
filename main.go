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

	r.GET("/dadjoke", func(c *gin.Context) {
		joke := getRandomDadJoke()
		c.JSON(http.StatusOK, gin.H{
			"setup":     joke.Setup,
			"punchline": joke.Punchline,
		})
	})
  
	r.GET("/8ball", func(c *gin.Context) {
		question := c.Query("q")
		if question == "" {
			question = "Will this release go smoothly?"
		}
		c.JSON(http.StatusOK, gin.H{
			"question": question,
			"answer":   shakeMagic8Ball(),
		})
	})
  
	r.GET("/hillbilly", func(c *gin.Context) {
		text := c.Query("text")
		if text == "" {
			text = "Hello everyone, I am going to go fishing over there"
		}
		c.JSON(http.StatusOK, gin.H{
			"original":   text,
			"translated": translateToHillbilly(text),
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
