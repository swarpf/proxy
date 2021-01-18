ARG BUILDPLATFORM

FROM --platform=${BUILDPLATFORM:-linux/amd64} tonistiigi/xx:golang AS xgo
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:alpine AS build

ENV CGO_ENABLED 0
ENV GO111MODULE on
ENV GOPROXY https://proxy.golang.org,direct
COPY --from=xgo / /

ARG TARGETPLATFORM
RUN go env

RUN echo "> running on $BUILDPLATFORM, building for $TARGETPLATFORM"

RUN apk --update --no-cache add \
    build-base \
    gcc \
    git \
    ca-certificates \
  && rm -rf /tmp/* /var/cache/apk/*

# Compile the cmd to a standalone binary
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-s -w -extldflags '-static'" ./cmd/proxy/

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:latest

COPY --from=build /etc/ssl/certs/ca-certificates.crt \
     /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/proxy /proxy
ENTRYPOINT ["/proxy"]
