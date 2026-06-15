# vim: set filetype=dockerfile:

FROM rust:1-bookworm AS sidecar
WORKDIR /sidecar
COPY sidecar/ .
RUN cargo build --release

FROM golang:1.26-bookworm
WORKDIR /sources

COPY --from=sidecar /sidecar/target/release/sc2json /usr/local/bin/sc2json
ENV SC2JSON=/usr/local/bin/sc2json

RUN go install github.com/air-verse/air@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
RUN curl -fsSL -o /usr/local/bin/biome "https://github.com/biomejs/biome/releases/download/%40biomejs%2Fbiome%402.5.0/biome-linux-x64" && chmod +x /usr/local/bin/biome

COPY go.mod go.sum ./
RUN go mod download

ENTRYPOINT ["sh", "entrypoint.sh"]
