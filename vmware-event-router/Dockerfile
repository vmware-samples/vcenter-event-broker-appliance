# golang:1.15.2-buster linux/amd64
FROM golang@sha256:0978cc067eb3f53901c00b70a024f182baa371bdfe7f35f3d64e56cab2471c4d AS builder
ARG VERSION
ARG COMMIT

WORKDIR /build

# modules
COPY go.mod .
COPY go.sum .
RUN go mod download

# code
COPY cmd cmd
COPY internal internal

# build
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -a -installsuffix nocgo -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT}" -o vmware-event-router cmd/router/main.go

# ubuntu:bionic-20200921 linux/amd64
FROM ubuntu@sha256:45c6f8f1b2fe15adaa72305616d69a6cd641169bc8b16886756919e7c01fa48b
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
