# golang:1.16.3-buster linux/amd64
FROM golang@sha256:3fc96f3fc8a5566a07ac45759bad6381397f2f629bd9260ab0994ef0dc3b68ca AS builder
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

# ubuntu:bionic-20210325 linux/amd64
FROM ubuntu@sha256:141d4a94a045f5b42bf6a6c74d9d868beab0ab5c5352de132f2a6068e1bd8d16
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
