package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		wsName := os.Getenv("onedevx-workspec")
		compName := os.Getenv("onedevx-component")
		c.JSON(200, gin.H{
			"message":   "pong",
			"component": compName,
			"workspec":  wsName,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
