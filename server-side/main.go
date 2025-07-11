package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Secret string `json:"secret"`
}

func main() {

	r := gin.Default()

	r.POST("/upload", upload)
	r.Run()
}

func config() (Config, error) {

	exePath, _ := os.Executable()
	configPath := filepath.Join(filepath.Dir(exePath), "config.json")
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		log.Fatal("Unable to parse config")
		return Config{}, err
	}

	return config, nil
}

func upload(c *gin.Context) {
	config, err := config()
	if err != nil {
		log.Fatal(err)
		return
	}

	if c.GetHeader("Authorization") != config.Secret {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(400, gin.H{"message": "No File"})
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
