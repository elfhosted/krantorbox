FROM	golang:1.15-alpine

RUN		mkdir /app \
		&& apk add git \
		&& go get github.com/igungor/go-putio/putio \
		&& go get github.com/radovskyb/watcher \
		&& go get golang.org/x/oauth2
ADD		. /app
WORKDIR		/app
ENV		GOPATH=/app
RUN		go build -o krantor
CMD		["/app/krantor"]
