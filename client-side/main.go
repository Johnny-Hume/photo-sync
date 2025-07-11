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
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	ServerURL string `json:"serverUrl"`
	Secret string `json:"secret"`
}

func config () (Config, error) {

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

func watchFolder(folderPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer watcher.Close()

	watcher.Add(folderPath)

	config, err := config()
	if err != nil {
		log.Fatal(err)
		return
	}

	for {
	    select {
		case event := <- watcher.Events:
		    if event.Op&fsnotify.Create == fsnotify.Create {
			time.Sleep(2 * time.Second)
			if strings.HasSuffix(event.Name, ".jpg") ||
				strings.HasSuffix(event.Name, ".png") ||
				strings.HasSuffix(event.Name, ".img") {
					uploadFile(event.Name, config)
				}
		}
		case err := <- watcher.Errors:
			log.Println("Error", err)
		}
	}
}

func uploadFile(path string, config Config) error {
	fmt.Printf("Uploading %s...\n", filepath.Base(path))
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    var buf bytes.Buffer
    writer := multipart.NewWriter(&buf)
    
    part, err := writer.CreateFormFile("photo", path)
    if err != nil {
        return err
    }
    
    io.Copy(part, file)
    writer.Close()

    req, err := http.NewRequest("POST", config.ServerURL+"/upload", &buf)
    if err != nil {
		fmt.Printf("❌ Failed to upload %s: %v\n", filepath.Base(path), err)
        return err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("Authorization", config.Secret)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
	fmt.Printf("✅ Uploaded %s\n", filepath.Base(path))
    return nil
}

func main() {
    exePath, err := os.Executable()
    if err != nil {
        log.Fatal("Can't find exe path:", err)
    }
    
    folderPath := filepath.Dir(exePath)
    log.Printf("Watching folder: %s", folderPath)
    
    watchFolder(folderPath)
}
