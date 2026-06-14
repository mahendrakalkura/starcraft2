# syntax=docker/dockerfile:1

# Build the Rust sidecar.
FROM rust:1-bookworm AS sidecar
WORKDIR /build
COPY sidecar/Cargo.toml sidecar/Cargo.lock ./
RUN mkdir src && echo 'fn main() {}' > src/main.rs && cargo build --release && rm -rf src
COPY sidecar/src ./src
RUN touch src/main.rs && cargo build --release

# Go toolchain image. The source is mounted at /sources and entrypoint.sh builds
# and runs it: air drives hot reload in development, a one-off build in production.
FROM golang:1.26-bookworm
RUN go install github.com/air-verse/air@latest \
    && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY --from=sidecar /build/target/release/sc2json /usr/local/bin/sc2json
ENV SC2JSON=/usr/local/bin/sc2json
WORKDIR /sources
ENTRYPOINT ["./entrypoint.sh"]
