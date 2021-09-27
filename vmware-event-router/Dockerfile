# golang:1.17.1-buster linux/amd64
FROM golang@sha256:2a0bd53951c1b1746c5adab10e1db1495cc6f2da0a3d21818466265e34a78597 AS builder
ARG VERSION
ARG COMMIT

WORKDIR /build

# modules
COPY go.mod .
COPY go.sum .

# dependencies
COPY vendor vendor

# code
COPY cmd cmd
COPY internal internal

# build
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix nocgo -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT}" -o vmware-event-router cmd/router/main.go

# ubuntu:bionic-20210827 linux/amd64
FROM ubuntu@sha256:b9caadbf898c50ce67da0ab5bafc4680997b010c3e17d2bb73d2ae5fe056e52b
ARG VERSION
ARG COMMIT
LABEL maintainer="mgasch@vmware.com" \
      commit="${COMMIT}"

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -g 61000 events
RUN useradd -g 61000 -l -m -s /bin/false -u 61000 events

WORKDIR /home/events
COPY --from=builder /build/vmware-event-router .

RUN chown -R events:events ./
USER events
ENTRYPOINT ["./vmware-event-router"]
