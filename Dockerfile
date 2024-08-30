FROM --platform=$BUILDPLATFORM node:20-alpine AS FRONT
WORKDIR /web
COPY ./web .
RUN yarn install --frozen-lockfile --network-timeout 1000000 && yarn run build && rm -rf node_modules

FROM --platform=$BUILDPLATFORM golang:latest AS BACK
WORKDIR /go/src/casdoor
COPY . .
RUN ./build.sh && go test -v -run TestGetVersionInfo ./util/system_test.go ./util/system.go > version_info.txt

FROM alpine:latest AS STANDARD
LABEL MAINTAINER="https://maxs.pro/"
ARG USER=vitalik
ARG TARGETOS
ARG TARGETARCH
ENV BUILDX_ARCH="${TARGETOS:-linux}_${TARGETARCH:-amd64}"

RUN sed -i 's/https/http/' /etc/apk/repositories && apk add --update sudo tzdata curl ca-certificates && update-ca-certificates && \
    adduser -D $USER -u 1000 && echo "$USER ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER && chmod 0440 /etc/sudoers.d/$USER && \
    mkdir logs && chown -R $USER:$USER logs

USER 1000
WORKDIR /
COPY --from=BACK --chown=$USER:$USER /go/src/casdoor/server_${BUILDX_ARCH} ./server
COPY --from=BACK --chown=$USER:$USER /go/src/casdoor/swagger ./swagger
COPY --from=BACK --chown=$USER:$USER /go/src/casdoor/conf/app.conf ./conf/app.conf
COPY --from=BACK --chown=$USER:$USER /go/src/casdoor/version_info.txt ./go/src/casdoor/version_info.txt
COPY --from=FRONT --chown=$USER:$USER /web/build ./web/build

ENTRYPOINT ["/server"]