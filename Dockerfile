FROM golang:alpine AS build

MAINTAINER yhm

WORKDIR /go/src/app

COPY go.mod .
COPY go.sum .
RUN go env -w GOPROXY=https://goproxy.io,direct && go mod download

COPY . .
RUN go build ./cmd/chatroom 

FROM alpine 
WORKDIR /app
COPY --from=build /go/src/app/chatroom ./
ENTRYPOINT [ "./chatroom" ]
