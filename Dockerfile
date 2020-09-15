FROM golang:1.15-alpine as builder

ENV GOPROXY=https://goproxy.io

ARG VERSION

ARG BUILD

ADD . /usr/local/go/src/base

WORKDIR /usr/local/go/src/base

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix cgo -ldflags "-s -X main.Version=${VERSION} -X main.Build=${BUILD}" -gcflags "all=-trimpath=${GOPATH}/src" -o taobaoke cmd/main.go

FROM alpine:3.12

ENV GIN_MODE="release"

RUN echo "http://mirrors.aliyun.com/alpine/v3.12/main/" > /etc/apk/repositories && \
        apk update && \
        apk add ca-certificates

WORKDIR /app

COPY --from=builder /usr/local/go/src/base/taobaoke ./taobaoke
# docker运行打开注释
#ADD ./configs ./configs


RUN chmod +x ./taobaoke


ENTRYPOINT ["./taobaoke","-conf","configs"]

#docker build --build-arg VERSION=$(echo "$(git symbolic-ref --short -q HEAD)-$(git rev-parse HEAD)"),BUILD=$(date +%FT%T%z) -t naist4869/taobaoke --network=host .

