FROM golang:alpine

ARG pkg=github.com/tony-yang/google-cloud-stack/bookshelf

RUN apk add --no-cache ca-certificates

RUN set -ex \
 && apk add --no-cache --virtual .build-deps \
    git \
 && go get -v $pkg/... \
 && apk del .build-deps

RUN go install $pkg/...

WORKDIR $GOPATH/src/$pkg/app

CMD echo "Use the app or pubsub_worker commands."; exit 1
