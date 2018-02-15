FROM alpine:latest

RUN \
  apk --no-cache add --virtual .rundeps \
    ca-certificates \
    curl \
    docker \
    git

ADD . /

CMD ["./docker-proxy", "-host", "localhost"]