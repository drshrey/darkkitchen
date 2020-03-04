FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
RUN apk add --no-cache curl
ENV GOPATH /go/
WORKDIR /go/src/github.com/drshrey/darkkitchen/client/
COPY . .
COPY src/testdata /tmp
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure
CMD ["go", "run", "src/main.go"]