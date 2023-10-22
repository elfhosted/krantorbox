package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/putdotio/go-putio"
	"golang.org/x/oauth2"
)

var (
	folderPath       string = os.Getenv("PUTIO_WATCH_FOLDER")
	putioToken       string = os.Getenv("PUTIO_TOKEN")
	downloadFolderID string = os.Getenv("PUTIO_DOWNLOAD_FOLDER_ID")
)

func connectToPutio() (*putio.Client, error) {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: putioToken})
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)

	client := putio.NewClient(oauthClient)

	return client, nil
}

func folderIDConvert() (int64, error) {
	folderID, err := strconv.ParseInt(downloadFolderID, 10, 32)
	if err != nil {
		str := fmt.Sprintf("strconv err: %v", err)
		err := errors.New(str)
		return 0, err
	}
	return folderID, nil
}

func uploadTorrentToPutio(filename string, client *putio.Client) error {
	// putio client's default timeout is 30sec. We'll allow a tad more.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*32))
	defer cancel()

	// Convert FolderID from string to int to use with Files.Upload
	folderID, err := folderIDConvert()
	if err != nil {
		return err
	}

	// Using open since Upload need an *os.File variable
	file, err := os.Open(filename)
	if err != nil {
		str := fmt.Sprintf("Openfile err: %v", err)
		err := errors.New(str)
		return err
	}

	// Uploading file to Putio
	log.Println("Read torrent file. Uploading...")
	result, err := client.Files.Upload(ctx, file, filename, folderID)
	if err != nil {
		str := fmt.Sprintf("Upload to Putio err: %v", err)
		err := errors.New(str)
		return err
	}

	fmt.Printf("Transferred to putio:              %v at %v\n-------------------\n", filename, result.Transfer.CreatedAt)
	return nil
}

func transferMagnetToPutio(filename string, client *putio.Client) error {
	// putio client's default timeout is 30sec. We'll allow a tad more.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*32))
	defer cancel()

	// Convert FolderID from string to int to use with Files.Upload
	folderID, err := folderIDConvert()
	if err != nil {
		return err
	}

	// Reading the link inside the magnet file to give to Putio
	log.Println("Reading...")
	magnetData, err := ioutil.ReadFile(filename)
	if err != nil {
		str := fmt.Sprintf("Couldn't read file %v: %v", filename, err)
		err := errors.New(str)
		return err
	}
	log.Println("magnetData: ", string(magnetData))

	// Using Transfer to DL file via magnet file
	result, err := client.Transfers.Add(ctx, string(magnetData), folderID, "")
	if err != nil {
		str := fmt.Sprintf("Transfer to putio err: %v", err)
		err := errors.New(str)
		return err
	}

	fmt.Printf("Transferred to putio:              %v at %v\n-------------------\n", filename, result.CreatedAt)
	// TODO: should we delete (or move) the file after successful uploading?
	//       To prevent accidental reuploads if someone moves files around?
	return nil
}

// https://stackoverflow.com/questions/67069723/keep-retrying-a-function-in-golang
func retryIfNeeded(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Println("retrying after error:", err)
			time.Sleep(sleep)
			sleep *= 2
		}
		err = f()
		if err == nil {
			return nil
		} else if !strings.Contains(err.Error(), "context deadline") {
			// Don't retry if its not a context deadline error
			return err
		}
	}
	return fmt.Errorf("retries failed. After %d attempts, last error: %s", attempts, err)
}

func checkFileType(filename string) (string, error) {
	// Checking what's at the end of the string
	isMagnet := strings.HasSuffix(filename, ".magnet")
	isTorrent := strings.HasSuffix(filename, ".torrent")

	if isMagnet {
		return "magnet", nil
	} else if isTorrent {
		return "torrent", nil
	} else {
		str := fmt.Sprintf("File isn't a torrent or magnet file: %v", filename)
		err := errors.New(str)
		return "", err
	}
}

func prepareFile(event fsnotify.Event, client *putio.Client) {
	time.Sleep(100 * time.Millisecond) // wait for WRITE event(s) to finish

	var err error
	var fileType string

	filename := event.Name

	// Checking if the file is a torrent of a magnet file
	torrentOrMagnet, err := checkFileType(filename)
	if err != nil {
		log.Println(err)
	} else {
		fileType = torrentOrMagnet
	}

	fmt.Printf("Detected new file in watch folder: %v\n", filename)
	// Retry the upload up to 3 times, in case of "context deadline exceeded" aka Timeout on the http POST
	sleepBetweenRetry := time.Duration(60) * time.Second

	if fileType == "torrent" {
		err = retryIfNeeded(3, sleepBetweenRetry, func() (err error) {
			return uploadTorrentToPutio(filename, client)
		})
		if err != nil {
			log.Println("ERROR: ", err)
		}
	} else if fileType == "magnet" {
		err = retryIfNeeded(3, sleepBetweenRetry, func() (err error) {
			return transferMagnetToPutio(filename, client)
		})
		if err != nil {
			log.Println("ERROR: ", err)
		}
	}
}

func watchFolder(client *putio.Client) {
	// https://pkg.go.dev/github.com/fsnotify/fsnotify
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				log.Println("event:", event)    // verbose logging of fsnotify events…
				if event.Has(fsnotify.Create) { // However CREATE is the only one we take action on
					// run in separate thread
					go prepareFile(event, client)
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.Fatalln(err)
			}
		}
	}()

	// Watch this folder for changes.
	if err := w.Add(folderPath); err != nil {
		log.Fatalln(err)
	}
	log.Println("Watching", folderPath)

	<-make(chan struct{})
}

func checkEnvVariables() error {
	var envToSet string

	if folderPath == "" {
		envToSet = "PUTIO_WATCH_FOLDER is not set / "
	}
	if downloadFolderID == "" {
		envToSet = envToSet + "PUTIO_DOWNLOAD_FOLDER_ID is not set / "
	}
	if putioToken == "" {
		envToSet = envToSet + "PUTIO_TOKEN is not set / "
	}
	if envToSet != "" {
		return errors.New(envToSet)
	}
	return nil
}

func main() {
	log.Println("Krantor Started")

	client, err := connectToPutio()
	if err != nil {
		log.Fatalln("connection to Putio err: ", err)
	}

	// We check that the env variable are set to avoid issues
	err = checkEnvVariables()
	if err != nil {
		log.Fatal(err)
	}

	// We start watching the folders
	watchFolder(client)
}
