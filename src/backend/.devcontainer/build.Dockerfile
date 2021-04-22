# Production environment (alias: base)
FROM golang:1.16-alpine as base
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh
WORKDIR /workspace

# Development environment
# Unfortunately, linux alpine doesn't have fswatch package by default, so we will need to download source code and make it by outselves.
FROM base as dev
RUN apk add --no-cache autoconf automake libtool gettext gettext-dev make g++ texinfo curl
WORKDIR /root
RUN wget https://github.com/emcrisostomo/fswatch/releases/download/1.14.0/fswatch-1.14.0.tar.gz
RUN tar -xvzf fswatch-1.14.0.tar.gz
WORKDIR /root/fswatch-1.14.0
RUN ./configure
RUN make
RUN make install
WORKDIR /workspace

RUN export PATH="$PATH:$(go env GOROOT)/misc/wasm"
RUN export GOOS=js
RUN export GOARCH=wasm