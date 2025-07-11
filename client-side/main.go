package main

import (
	"bytes"
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

func watchFolder(folderPath string) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()
    
    watcher.Add(folderPath)
    
    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Create == fsnotify.Create {
                // Wait a bit for file to finish copying
                time.Sleep(2 * time.Second)
                
                if strings.HasSuffix(event.Name, ".jpg") || 
                   strings.HasSuffix(event.Name, ".png") {
                    uploadFile(event.Name, "http://localhost:8080")
                }
            }
        case err := <-watcher.Errors:
            log.Println("Error:", err)
        }
    }
}

func uploadFile(path string, serverURL string) error {
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

    req, err := http.NewRequest("POST", serverURL+"/upload", &buf)
    if err != nil {
		fmt.Printf("❌ Failed to upload %s: %v\n", filepath.Base(path), err)
        return err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("Authorization", "It's 2AM I'm not setting up configs")
    
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