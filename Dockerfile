# Build rpc server
FROM rust:1.96 AS rpc-builder

RUN cargo install \
    --git https://github.com/chatmail/core/ \
    deltachat-rpc-server

# Build Go app
FROM golang:1.26 AS go-builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o /out/bridge .

# Runtime
FROM ubuntu:resolute

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=rpc-builder /usr/local/cargo/bin/deltachat-rpc-server /usr/local/bin/
COPY --from=go-builder /out/bridge /usr/local/bin/

COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh", "dcaccount:nine.testrun.org"]
