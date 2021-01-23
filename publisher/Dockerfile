FROM golang:alpine as build

LABEL maintainer "github.com/jxsl13"

WORKDIR /build
COPY * ./

ENV CGO_ENABLED=0
ENV GOOS=linux 

RUN go get -d && go build -a -ldflags '-w -extldflags "-static"' -o publisher .


FROM alpine:latest as minimal
WORKDIR /app
COPY --from=build /build/publisher .
VOLUME ["/data"]
ENTRYPOINT ["/app/publisher"]