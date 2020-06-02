# golang:1.14.4-stretch amd64
FROM golang@sha256:8b6b86e45c7847fea9cf401da52e41747cfb18bbfb4aebe924b4ef13a3366521 AS builder
ARG VERSION
ARG COMMIT

WORKDIR /build

# install linter into ./bin/
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.27.0

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd cmd
COPY internal internal

RUN ./bin/golangci-lint run ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT}" -o vmware-event-router cmd/main.go

# debian:stable-slim amd64
FROM debian@sha256:aa1db351593d2849034c0395bc604dd65aa80ae4471da4a37a0c38e30aed3ab8
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
