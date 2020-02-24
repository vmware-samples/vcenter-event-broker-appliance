# golang:1.13.7-stretch
FROM golang@sha256:d16d1e0b4021e15dfc17dc58e794ad1e794155abf74c579a4c7b8a2c83ff8682 AS builder
ARG VERSION
ARG COMMIT

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd cmd
COPY internal internal
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT}" -o vmware-event-router cmd/main.go

# debian:stable-slim
FROM debian@sha256:55f6837fd25a8e2ab94790bc73e762ffeb0d5d2e510177aef76101d2822d5937
ARG VERSION
ARG COMMIT
LABEL maintainer="mgasch@vmware.com" \
      commit="${COMMIT}"

RUN apt-get update && apt-get install -y \
    ca-certificates=20190110 \
    procps=2:3.3.15-2 \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -g 61000 events
RUN useradd -g 61000 -l -m -s /bin/false -u 61000 events

WORKDIR /home/events
COPY --from=builder /build/vmware-event-router .

RUN chown -R events:events ./
USER events
ENTRYPOINT ["./vmware-event-router"]
