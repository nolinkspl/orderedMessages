FROM golang:1.21-bullseye

ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

ENV APP_HOME /message_loader
RUN mkdir -p "$APP_HOME"

WORKDIR "$APP_HOME"

ADD src/. .

RUN go mod download && go mod vendor && go mod verify

RUN go build -o /go/bin/messageLoaderApp

CMD ["/go/bin/messageLoaderApp"]
