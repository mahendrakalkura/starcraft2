# vim: set filetype=dockerfile:

FROM rust:1-bookworm AS sidecar
WORKDIR /build

COPY sidecar/Cargo.toml sidecar/Cargo.lock ./
RUN mkdir src && echo 'fn main() {}' > src/main.rs && cargo build --release && rm -rf src

COPY sidecar/src ./src
RUN touch src/main.rs && cargo build --release

FROM golang:1.26-bookworm
WORKDIR /build

RUN go install github.com/air-verse/air@latest && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

RUN curl -fsSL -o /usr/local/bin/biome "https://github.com/biomejs/biome/releases/download/%40biomejs%2Fbiome%402.5.0/biome-linux-x64" && chmod +x /usr/local/bin/biome

COPY --from=sidecar /build/target/release/sc2json /usr/local/bin/sc2json
ENV SC2JSON=/usr/local/bin/sc2json

COPY go.mod go.sum ./
RUN go mod download

WORKDIR /sources
ENTRYPOINT ["sh", "entrypoint.sh"]
