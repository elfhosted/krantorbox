# Putio-go

After searching for something that could help facilitate transfering files from local/apps/services to Putio, nothing corresponded to what I was looking for.
Almost all projects were at least 2years old.

So I decided to do something simple in Go, Putio-Go

## Table of Contents

* [Installation](#installation)
* [Configuration](#configuration)
* [Advanced Usage](#advanced-usage)
  * [Docker](#docker)
  * [Docker-compose](#docker-compose)
* [Example](#example)

## Installation

Just build the image with the given Dockerfile:

    docker build --no-cache -t putio-go .

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
  --name=putio \
  -e PUID=1000 \
  -e PGID=1000 \
  -e PUTIO_TOKEN=xxx \
  -e PUTIO_WATCH_FOLDER=/torrents \
  -e PUTIO_DOWNLOAD_FOLDER_ID=0 \
  -v /path/to/torrent:/torrents \
  --restart unless-stopped \
  putio-go
```

### Docker-compose

```
---
version: "3.7"
services:
  putio:
    image: putio-go
    container_name: putio
    environment:
      - PUID=1000
      - PGID=1000
      - PUTIO_TOKEN=xxx
      - PUTIO_WATCH_FOLDER=/torrents
      - PUTIO_DOWNLOAD_FOLDER_ID=0
    volumes:
      - /path/to/torrent:/torrents
    restart: unless-stopped
```

### Example
![alt text](https://i.imgur.com/xzM3Srr.png "Example of logs given by PutioUploadr")