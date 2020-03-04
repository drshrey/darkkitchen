FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
RUN apk add --no-cache curl
ENV GOPATH /go/
WORKDIR /go/src/github.com/drshrey/darkkitchen/backend/
COPY . .
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure
# Build the binary.
RUN cd /go/src/github.com/drshrey/darkkitchen/backend/src/ && CGO_ENABLED=0 go build -o /go/bin/ckse
FROM scratch
COPY --from=builder /go/bin/ckse /go/bin/ckse
ENTRYPOINT ["/go/bin/ckse"]