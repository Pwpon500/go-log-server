FROM golang:alpine

ADD go-log-server_linux_amd64 /go-log-server

ENTRYPOINT ["/go-log-server"]
