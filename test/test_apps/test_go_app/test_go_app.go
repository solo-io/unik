package main

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	r := gin.Default()
	r.GET("/ping_test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/env_test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": os.Getenv("KEY"),
		})
	})
	r.GET("/mount_test", func(c *gin.Context) {
		data, err := ioutil.ReadFile(filepath.Join("/data", "data.txt"))
		if err != nil {
			log.Printf("mounts_test error: %v", err)
			c.Error(err)
		}
		c.JSON(200, gin.H{
			"message": string(data),
		})
	})
	r.Run(":8080") // listen and server on 0.0.0.0:8080
}
