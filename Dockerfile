FROM golang as golang

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /build
ADD / /build

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -s' -o wechatbot

FROM alpine:latest

WORKDIR /dist

RUN apk add --no-cache bash
COPY  config.dev.json /dist/config.json
COPY templates/* /dist/templates/
COPY --from=golang /build/wechatbot /dist


EXPOSE 5000
ENTRYPOINT ["/dist/wechatbot"]