FROM golang:alpine 
WORKDIR /go/src/main
RUN go install github.com/hanyue2020/evadoc/cmd/evadoc@latest

ENTRYPOINT ["evadoc"]
