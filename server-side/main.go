package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main () {

	r := gin.Default()

	r.POST("/upload", upload)
	r.Run()
}

func upload (c *gin.Context) {
	if c.GetHeader("Authorization") != "It's 2AM I'm not setting up configs" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(400, gin.H{"message":"No File"})
		return
	}
	defer file.Close()

	
	dst := filepath.Join("../sharedPhotos", header.Filename)

	out, err := os.Create(dst)
	if err != nil {
		c.JSON(500, gin.H{"message": "Cant save file", "err": err})
		return
	}
	defer out.Close()

	io.Copy(out, file)
	c.JSON(200, gin.H{"status": "uploaded"})
}