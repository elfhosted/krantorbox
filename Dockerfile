FROM	golang:1.15-alpine

ADD		. /app
WORKDIR	/app
RUN		apk add git \
		&& go get github.com/igungor/go-putio/putio \
		&& go get github.com/radovskyb/watcher \
		&& go get golang.org/x/oauth2

RUN		go build -o krantor
CMD		["/app/krantor"]
