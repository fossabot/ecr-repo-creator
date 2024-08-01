ARG GOLANG_VERSION="1.22.5"
ARG KANIKO_VERSION="v1.23.2"

FROM golang:${GOLANG_VERSION}-bookworm AS build-stage

WORKDIR /app

COPY . .

RUN set -xe \
    && CGO_ENABLED=0 go build .

FROM gcr.io/kaniko-project/executor:${KANIKO_VERSION}-debug AS final-stage

COPY --from=ghcr.io/jqlang/jq /jq /usr/local/bin/
COPY --from=0 /app/ecr-repo-creator /usr/local/bin/
