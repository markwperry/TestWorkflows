package main

import (
	"net/http"
	"os"
	"strconv"

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

	r.GET("/fortune", func(c *gin.Context) {
		fortune, luckyNums := getRandomFortune()
		c.JSON(http.StatusOK, gin.H{
			"fortune":      fortune,
			"luckyNumbers": luckyNums,
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

	r.GET("/coinflip", func(c *gin.Context) {
		n := 1
		if q := c.Query("n"); q != "" {
			if parsed, err := strconv.Atoi(q); err == nil && parsed > 0 && parsed <= 1000 {
				n = parsed
			}
		}
		if n == 1 {
			c.JSON(http.StatusOK, gin.H{"result": flipCoin()})
		} else {
			c.JSON(http.StatusOK, gin.H{"flips": n, "results": flipMultiple(n)})
		}
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
