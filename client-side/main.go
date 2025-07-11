package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	ServerURL   string `json:"serverUrl"`
	Secret      string `json:"secret"`
	WorkerCount int    `json:"workerCount"`
}

var config Config

func (c *Config) initConfig() error {

	exePath, _ := os.Executable()
	configPath := filepath.Join(filepath.Dir(exePath), "config.json")
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(file, c); err != nil {
		log.Fatal("Unable to parse config")
	}

	return nil
}

func watchFolder(folderPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	watcher.Add(folderPath)

	uploadQueue := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < config.WorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filename := range uploadQueue {
				uploadFile(filename)
			}
		}()
	}

	entries, _ := os.ReadDir(folderPath)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".NEF") ||
			strings.HasSuffix(e.Name(), ".jpg") ||
			strings.HasSuffix(e.Name(), ".png") ||
			strings.HasSuffix(e.Name(), ".img") {
			uploadQueue <- filepath.Join(folderPath, e.Name())
		}
	}

	close(uploadQueue)
	wg.Wait()
	log.Println("Done! Did I mention you look beautiful today?")

	// for {
	// 	select {
	// 	case event := <-watcher.Events:
	// 		if event.Op&fsnotify.Create == fsnotify.Create {
	// 			if strings.HasSuffix(event.Name, ".jpg") ||
	// 				strings.HasSuffix(event.Name, ".JPG") ||
	// 				strings.HasSuffix(event.Name, ".NEF") ||
	// 				strings.HasSuffix(event.Name, ".img") {
	// 				go func(filename string) {
	// 					uploadQueue <- event.Name
	// 				}(event.Name)
	// 			}
	// 		}
	// 	case err := <-watcher.Errors:
	// 		log.Println("Error", err)
	// 	}
	// }
}

func uploadFile(path string) {
	time.Sleep(time.Second * 1)
	fmt.Printf("Uploading %s...\n", filepath.Base(path))
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("photo", path)
	if err != nil {
		log.Println(err)
		return
	}

	io.Copy(part, file)
	writer.Close()

	req, err := http.NewRequest("POST", config.ServerURL+"/upload", &buf)
	if err != nil {
		log.Printf("❌ Failed to upload %s: %v\n", filepath.Base(path), err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", config.Secret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.Status != "200 OK" {
		log.Printf("❌ Failed to upload %s: [%v]\n", filepath.Base(path), resp.Status)
		return
	}

	fmt.Printf("✅ Uploaded %s\n", filepath.Base(path))
}

func main() {
	log.Println("     =====     ")
	log.Println("Avery is the fairest of all the land. I'll be damned if she moves a photo by hand.")
	log.Println("     =====     ")

	config.initConfig()

	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("Can't find exe path:", err)
	}

	folderPath := filepath.Join(filepath.Dir(exePath), "..")

	log.Printf("Watching folder: %s", folderPath)

	watchFolder(folderPath)
}
