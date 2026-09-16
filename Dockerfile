# Фронт — новая shadcn-консоль upstream'а из web/ (с нашими правками: magic link и т.д.); web-old/ в образ не попадает
FROM --platform=$BUILDPLATFORM node:20-alpine AS front
WORKDIR /web

# Copy only dependency files first for better caching
COPY ./web/package.json ./web/yarn.lock ./
RUN yarn install --frozen-lockfile --network-timeout 1000000

# Copy source files
COPY ./web .

# Brand assets (logos, favicon) live outside the repository: the tree carries no
# brand, the image gets one at build time. The directory is copied to
# web/public/brand, so its files are served at /brand/<file> and can be pointed at
# with CASDOOR_BRAND_LOGO_URL=/brand/logo.svg etc. See docs/branding.md.
ARG BRAND_ASSETS_DIR=web/brand-default
COPY ./${BRAND_ASSETS_DIR}/ ./public/brand/

RUN NODE_OPTIONS="--max-old-space-size=4096" yarn run build


FROM --platform=$BUILDPLATFORM golang:latest AS back
WORKDIR /go/src/casdoor
ARG TARGETOS
ARG TARGETARCH

# Copy only go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY . .

RUN rm -rf web web-old && ./build.sh

FROM alpine:latest AS standard
LABEL MAINTAINER="https://casdoor.org/"
ARG USER=casdoor
ARG TARGETOS
ARG TARGETARCH
ENV BUILDX_ARCH="${TARGETOS:-linux}_${TARGETARCH:-amd64}"

RUN sed -i 's/https/http/' /etc/apk/repositories && apk add --update sudo tzdata curl ca-certificates && update-ca-certificates && \
    adduser -D $USER -u 1000 && echo "$USER ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER && chmod 0440 /etc/sudoers.d/$USER && \
    mkdir logs && chown -R $USER:$USER logs

USER 1000
WORKDIR /
COPY --from=back --chown=$USER:$USER /go/src/casdoor/server_${BUILDX_ARCH} ./server
COPY --from=back --chown=$USER:$USER /go/src/casdoor/swagger ./swagger
COPY --from=back --chown=$USER:$USER /go/src/casdoor/conf/app.conf ./conf/app.conf
COPY --from=front --chown=$USER:$USER /web/build ./web/build

ENTRYPOINT ["/server"]