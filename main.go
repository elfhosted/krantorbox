package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/igungor/go-putio/putio"
	"golang.org/x/oauth2"

	"github.com/radovskyb/watcher"
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

func folderIDConvert(fodlerID string) (int64, error) {
	folderID, err := strconv.ParseInt(downloadFolderID, 10, 32)
	if err != nil {
		str := fmt.Sprintf("strconv err: %v", err)
		err := errors.New(str)
		return 0, err
	}
	return folderID, nil
}

func uploadToPutio(filename string, filepath string, client *putio.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*5))
	defer cancel()

	// Convert FolderID from string to int to use with Files.Upload
	folderID, err := folderIDConvert(downloadFolderID)
	if err != nil {
		return err
	}

	// Using open since Upload need an *os.File variable
	file, err := os.Open(folderPath + "/" + filename)
	if err != nil {
		str := fmt.Sprintf("Openfile err: %v", err)
		err := errors.New(str)
		return err
	}

	// Uploading file to Putio
	result, err := client.Files.Upload(ctx, file, filename, folderID)
	if err != nil {
		str := fmt.Sprintf("Upload to Putio err: %v", err)
		err := errors.New(str)
		return err
	}

	// fmt.Println("-------------------")
	fmt.Printf("File: %v has been transfered to Putio at %v\n-------------------\n", filename, result.Transfer.CreatedAt)
	return nil
}

func transferToPutio(filename string, filepath string, client *putio.Client) error {
	// Creating a context with 5 second timout in case Transfer is too long
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*5))
	defer cancel()

	// Convert FolderID from string to int to use with Files.Upload
	folderID, err := folderIDConvert(downloadFolderID)
	if err != nil {
		return err
	}

	// Reading the link inside the magnet file to give to Putio
	magnetData, err := ioutil.ReadFile(folderPath + "/" + filename)
	if err != nil {
		str := fmt.Sprintf("Couldn't read file %v: %v", filename, err)
		err := errors.New(str)
		return err
	}

	// Using Transfer to DL file via magnet file
	result, err := client.Transfers.Add(ctx, string(magnetData), folderID, "")
	if err != nil {
		str := fmt.Sprintf("Transfer to putio err: %v", err)
		err := errors.New(str)
		return err
	}

	// fmt.Println("-------------------")
	fmt.Printf("File: %v has been transfered to Putio at %v\n-------------------\n", filename, result.CreatedAt)
	return nil
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
		str := fmt.Sprintf("File: %v doesn't seems to be either a torrent or a magnet file", filename)
		err := errors.New(str)
		return "", err
	}
}

func runScript() ([]byte, error) {
	// Running the bash script that will rename files by replacing space by dot
	// Argument is where to rename file, here is the watch folder where files will be DL
	cmd := exec.Command("/bin/sh", "/app/chFileName.sh", folderPath)
	// cmd.Env = []string{"A=B"}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func cleaningFilename(event watcher.Event) string {
	// Retrieve the events and convert it into string to eb able to work with it
	wordsWithSpace := fmt.Sprintf("%v", event)

	// Convert the string by replacing the space by dot
	wordsWithSpace = strings.Replace(wordsWithSpace, " ", ".", -1)
	// Regex to only take what's inside the double quote
	re2 := regexp.MustCompile(`"(.*?)"`)
	cleanFileNameRegex := re2.FindStringSubmatch(wordsWithSpace)
	// We only take the second result as it's the good one
	// fmt.Println("Not totally finish filename: ", cleanFileNameRegex)
	cleanFileName := cleanFileNameRegex[1]

	return cleanFileName
}

func prepareFile(event watcher.Event, client *putio.Client) {
	var filepath string
	var err error
	var fileType string

	// We run the script to **rename the file** by repalcing space by dot
	_, err = runScript()
	if err != nil {
		fmt.Println("Couldn't run the script: ", err)
	}

	// From events with lots of informations
	// To a clean filename ready to be used
	cleanFilename := cleaningFilename(event)

	// Checking if the file is a torrent of a magnet file
	torrentOrMagnet, err := checkFileType(cleanFilename)
	if err != nil {
		log.Println(err)
	} else {
		fileType = torrentOrMagnet
	}

	fmt.Printf("File: %v has been added to the folder\n\n", cleanFilename)
	if fileType == "torrent" {
		err = uploadToPutio(cleanFilename, filepath, client)
		if err != nil {
			log.Println("err: ", err)
		}
	} else if fileType == "magnet" {
		err = transferToPutio(cleanFilename, filepath, client)
		if err != nil {
			log.Println("err: ", err)
		}
	}
}

func watchFolder(client *putio.Client) {
	w := watcher.New()

	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	//
	// If SetMaxEvents is not set, the default is to send all events.
	w.SetMaxEvents(1)

	// Only notify rename and move events.
	w.FilterOps(watcher.Move, watcher.Create)

	// Only files that match the regular expression during file listings
	// will be watched.
	// r := regexp.MustCompile("^abc$")
	// w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		for {
			select {
			case event := <-w.Event:
				prepareFile(event, client) // Print the event's info.
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch this folder for changes.
	if err := w.Add(folderPath); err != nil {
		log.Fatalln(err)
	}

	fmt.Println()

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
		fmt.Println()
	}
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
	log.Println("putioUploadr Started")

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
