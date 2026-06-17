#!/bin/bash
#try to connect to google to determine whether user need to use proxy
curl www.google.com -o /dev/null --connect-timeout 5 2> /dev/null
if [ $? == 0 ]
then
    echo "Successfully connected to Google, no need to use Go proxy"
else
    echo "Google is blocked, Go proxy is enabled: GOPROXY=https://goproxy.cn,direct"
    export GOPROXY="https://goproxy.cn,direct"
fi

TARGETOS=${TARGETOS:-linux}
TARGETARCH=${TARGETARCH:-amd64}
GO_BUILD_P=${GO_BUILD_P:-1}
export GOCACHE=${GOCACHE:-/tmp/go-cache}
export GOTMPDIR=${GOTMPDIR:-/tmp/go-tmp}
mkdir -p "$GOCACHE" "$GOTMPDIR"

CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -p "$GO_BUILD_P" -ldflags="-w -s" -o server_${TARGETOS}_${TARGETARCH} .
rm -rf "$GOCACHE" "$GOTMPDIR"

# old develop
# CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o server .
