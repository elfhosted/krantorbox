
# Krantor

Watches on local directory for new .torrent or .magnet files. When any are added, it uploads to a single put.io folder.


## Changelog

* 2023 Oct (paulirish): Add retries to fix 'context deadline exceeded' timeout on uploads. Update deps
* 2023 July ([agunal](https://github.com/agunal/krantor)) Fix dockerfile, add threading, timeout
* 2023 May ([paulirish](https://gitlab.com/paulirish/krantor)): Switch to inotify, using less CPU/energy (from polling every 100ms to waiting on events). Handle multiple files being added at once.
* 2020 Sept ([klippz](https://gitlab.com/klippz/krantor)): Original commits

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

Or build the binary for your target platform.

## Configuration

To make it run, you need to set 3 ENV variables:
```
PUTIO_TOKEN               [Putio Token to communication with their APIs]
PUTIO_WATCH_FOLDER        [Folder to watch for new files]
PUTIO_DOWNLOAD_FOLDER_ID  [Go into your put.io folder in your browser and copy the number in the URL:
https://app.put.io/files/<folder-id> ]
```

For oauth token (API Key): https://help.put.io/en/articles/5972538-how-to-get-an-oauth-token-from-put-io

If you need to watch multiple folders (TV, Movies, etc), you'll have to run multiple times.


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


### Integration 

[From reddit thread](https://www.reddit.com/r/putdotio/comments/136u8r2/comment/jisszuf/)...

>Use Krantor https://gitlab.com/klippz/krantor This gets torrents from sonarr into put.io Set up as a Download Client > Torrent blackhole. Follow readme instructions, however I do separate local folders for Torrent and Watch. My putio download folder is named `/dropzone/TV``. (I also follow the TRaSH guide for hardlinks)
>
>You need something to automatically download from put.io into your local "downloads" folder. (Probably.) I use `rclone`. Set up rclone and add a putio remote. Test it with rclone ls and stuff. Here's the rclone command that'll move (copy and delete) files from putio to your machine: `rclone -v --config="pathto/rclone.conf" --log-file="pathto/rclone.log" move putio:dropzone/TV /data/Downloads/TV/ --delete-empty-src-dirs` I run this every 30 minutes. You probably want to ensure a second invocation doesn't overlap, so.. handle that with your task scheduler mechanism or manually with `flock``.
>
>I personally never understood how people use Sonarr when all indexers are paid/private except for rarbg (RIP). I found a solution with **Jackett**. In there, I added EZTV, 1337, TPB.. and then hooked Sonarr up to those. Finally both search and rss both work effectively.
>
>If using radarr, repeat all the above with it for movies. I personally get a lot of value from Sonarr, but for movies the chill.institute + download (manually, ftp, rclone, etc) seems fine and radarr seems kinda overkill. But to each their own. :)

### Example
![alt text](https://i.imgur.com/1jUU1xn.png "Example of logs given by Krantor")

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

## Hacking

* Initial setup: `go mod init gitlab.com/paulirish/krantor`
* Dependency update: for direct deps in go.mod, replace versions with `latest` then run `go mod tidy`.
* Running: `PUTIO_TOKEN=<token> PUTIO_WATCH_FOLDER=<localpath> PUTIO_DOWNLOAD_FOLDER_ID=<folderid> go run main.go`
