
# KranTorbox

Watches a local directory for new .torrent, .magnet, .usenet files and uploads them to TorBox.

This serverse as a focussed, lightweight alternative to [West's Blackhole Script](https://github.com/westsurname/scripts).

Forked from [Paul Irish's Krantor](https://gitlab.com/paulirish/krantor), itself a fork of [Krantor by klippz](https://gitlab.com/klippz/krantor) - look to those for put.io support.

## Differences from Krantor


**TorBox** ✅ 

**put.io** ❌



**Added**: Usenet | Option to delete .torrent, .magnet, .usenet files after upload.


**Removed**: Synology build script (I'm unable to test).

## Table of Contents

* [Requirements](#requirements)
* [Installation](#installation)
* [Configuration](#configuration)
* [Advanced Usage](#advanced-usage)
  * [Docker](#docker)
  * [Docker-compose](#docker-compose)
* [How to use with Sonarr / Radarr](#how-to-use-with-sonarr/radarr)
* [Example](#example)

## Requirements

Docker, Go

## Installation

Build the image with the given Dockerfile:

    docker build -t krantorbox:local .

## Configuration

2 ENV variables are required - **copy your API Key from [torbox.app/settings](https://torbox.app/settings)**:
```
TORBOX_API_KEY             [Key for TorBox's API]
TORBOX_WATCH_FOLDER        [Folder to watch for new files]
```


1 ENV is optional (and defaults to `false`):
```
DELETE_AFTER_UPLOAD        [Delete original .torrent, .magnet, .usenet file after upload]
```

### Use with Sonarr / Radarr

 * In Radarr / Sonarr go to `Settings` -> `Download Clients` - *you need to do this for both Radarr and Sonarr separately*


 * Add `Torrent Blackhole` or `Usenet Blackhole` - *setting up both will work, so repeat the steps if you want to watch for torrents, magnets **and** usenet files*

 * Chose a suitable name e.g. `Torrent Blackhole`

 * Set `Torrent/Usenet Folder` to your chosen directory e.g. `/blackhole` - *must be the same as your `TORBOX_WATCH_FOLDER`*


## Advanced Usage

### Docker

```
docker create \
  --name=krantorbox \
  -e TORBOX_API_KEY=xxx \
  -e TORBOX_WATCH_FOLDER=/blackhole \
  -e DELETE_AFTER_UPLOAD=true \
  -v ./blackhole:/blackhole \
  --restart unless-stopped \
  krantorbox
```

### Docker-compose

```
services:

  krantorbox:
    container_name: krantorbox
    image: krantorbox:local
    environment:
      - TORBOX_API_KEY=xxx
      - TORBOX_WATCH_FOLDER=/blackhole
      - DELETE_AFTER_UPLOAD=true
    volumes:
      - ./blackhole:/blackhole
    restart: unless-stopped
```
