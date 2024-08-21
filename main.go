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

var (
	folderPath        = os.Getenv("TORBOX_WATCH_FOLDER")
	torboxAPIKey      = os.Getenv("TORBOX_API_KEY")
	deleteAfterUpload bool
)

const (
	maxRetries        = 3
	retryDelay        = 5 * time.Second
	torboxAPIBase     = "https://api.torbox.app"
	torboxAPIVersion  = "v1"
)

type TorBoxResponse struct {
	Success bool   `json:"success"`
	Detail  string `json:"detail"`
}

func init() {
	deleteAfterUpload = strings.ToLower(os.Getenv("DELETE_AFTER_UPLOAD")) == "true"
}

func uploadToTorBox(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	for attempt := 0; attempt < maxRetries; attempt++ {
		var err error
		if strings.HasSuffix(filename, ".nzb") {
			err = tryUploadUsenet(file, filename)
		} else {
			err = tryUploadTorrent(file, filename)
		}

		if err == nil {
			log.Printf("Successfully uploaded %s to TorBox", filename)
			if deleteAfterUpload {
				if err := os.Remove(filename); err != nil {
					log.Printf("Warning: Failed to delete file %s: %v", filename, err)
				} else {
					log.Printf("Deleted file: %s", filename)
				}
			}
			return nil
		}

		log.Printf("Attempt %d failed: %v. Retrying in %v...", attempt+1, err, retryDelay)
		time.Sleep(retryDelay)

		// Rewind the file for the next attempt
		_, err = file.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("error rewinding file: %v", err)
		}
	}

	return fmt.Errorf("failed to upload after %d attempts", maxRetries)
}

func tryUploadTorrent(file *os.File, filename string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return fmt.Errorf("error creating form file: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("error copying file to form: %v", err)
	}

	writer.WriteField("seed", "1")
	writer.WriteField("allow_zip", "true")

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error closing multipart writer: %v", err)
	}

	url := fmt.Sprintf("%s/%s/api/torrents/createtorrent", torboxAPIBase, torboxAPIVersion)
	return sendRequest(url, writer.FormDataContentType(), body)
}

func tryUploadUsenet(file *os.File, filename string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return fmt.Errorf("error creating form file: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("error copying file to form: %v", err)
	}

	// Use the filename without the path and .nzb extension
	name := strings.TrimSuffix(filepath.Base(filename), ".nzb")
	err = writer.WriteField("name", name)
	if err != nil {
		return fmt.Errorf("error writing name field: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error closing multipart writer: %v", err)
	}

	url := fmt.Sprintf("%s/%s/api/usenet/createusenetdownload", torboxAPIBase, torboxAPIVersion)
	return sendRequest(url, writer.FormDataContentType(), body)
}

func sendRequest(url, contentType string, body *bytes.Buffer) error {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+torboxAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var torboxResp TorBoxResponse
	err = json.NewDecoder(resp.Body).Decode(&torboxResp)
	if err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	if !torboxResp.Success {
		return fmt.Errorf("TorBox API error: %s", torboxResp.Detail)
	}

	return nil
}

func watchFolder() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					if strings.HasSuffix(event.Name, ".torrent") || strings.HasSuffix(event.Name, ".magnet") || strings.HasSuffix(event.Name, ".nzb") {
						log.Println("New file detected:", event.Name)
						err := uploadToTorBox(event.Name)
						if err != nil {
							log.Printf("Error uploading %s: %v", event.Name, err)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	err = watcher.Add(folderPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Watching folder: %s", folderPath)
	<-done
}

func main() {
	log.Println("Krantorbox Auto-Upload Started")
	log.Printf("Watching folder: %s", folderPath)
	log.Printf("TorBox API Base: %s", torboxAPIBase)
	log.Printf("TorBox API Version: %s", torboxAPIVersion)
	if deleteAfterUpload {
		log.Println("File deletion after upload is enabled")
	} else {
		log.Println("File deletion after upload is disabled")
	}

	if folderPath == "" || torboxAPIKey == "" {
		log.Fatal("Please set all required environment variables")
	}

	watchFolder()
}