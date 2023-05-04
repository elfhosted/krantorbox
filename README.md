# Krantor

After searching for something that could help facilitate transfering files from local/apps/services to Putio, nothing corresponded to what I was looking for.
Almost all projects were at least 2years old.

So I decided to do something simple in Go, Krantor

## Table of Contents

* [Installation](#installation)
* [Configuration](#configuration)
* [Advanced Usage](#advanced-usage)
  * [Docker](#docker)
  * [Docker-compose](#docker-compose)
* [How to use with Sonarr/Radarr](#how-to-use-with-sonarr/radarr)
* [Example](#example)

## Installation

Just build the image with the given Dockerfile:

    docker build --no-cache -t krantor .

## Configuration

To make it run, you need to set 3 ENV variables:
```
PUTIO_TOKEN               [Putio Token to communication with their APIs]
PUTIO_WATCH_FOLDER        [Folder to watch for new files]
PUTIO_DOWNLOAD_FOLDER_ID  [ID of the folder in PUTIO where you want to uplaod the file, in general it's 0 but could be something else]
```
To know the DOWNLOAD_FOLDER_ID, just go to your Putio account a chose the folder where you want your file to bbe uploaded
In the URL, you should see something like: `https://app.put.io/files/your_folder_id`

## Advanced Usage

### Docker

```
docker create \
  --name=krantor \
  -e PUTIO_TOKEN=xxx \
  -e PUTIO_WATCH_FOLDER=/torrents \
  -e PUTIO_DOWNLOAD_FOLDER_ID=0 \
  -v /path/to/torrent:/torrents \
  --restart unless-stopped \
  krantor
```

### Docker-compose

```
---
version: "3.7"
services:
  putio:
    image: krantor
    container_name: krantor
    environment:
      - PUTIO_TOKEN=xxx
      - PUTIO_WATCH_FOLDER=/torrents
      - PUTIO_DOWNLOAD_FOLDER_ID=0
    volumes:
      - /path/to/torrent:/torrents
    restart: unless-stopped
```

### How to use with Sonarr/Radarr
What you have to do is:
 * Go to your Radarr/Sonarr configuration
 * `Download Client` tab
 * Add a new `torrent blackhole` client
 * Chose a name
 * In torrent & watch folder, put the same folder you set as `PUTIO_WATCH_FOLDER`
   * If for `PUTIO_WATCH_FOLDER` you set `/torrent`, you should put the same in torrent & watch folder
 * Save magnet file !!
 * Done !

### Example
![alt text](https://i.imgur.com/1jUU1xn.png "Example of logs given by Krantor")


## TODO

* make dev mention of `go get -d ./...` and `go build` and the syn bash script
* setup instructions a la https://www.reddit.com/r/putdotio/comments/136u8r2/comment/jisszuf/?utm_source=reddit&utm_medium=web2x&context=3