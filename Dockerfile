# syntax=docker/dockerfile:1

# Stage 1: build the Rust sidecar (sc2json).
FROM rust:1-bookworm AS sidecar
WORKDIR /build
COPY sidecar/Cargo.toml sidecar/Cargo.lock ./
RUN mkdir src && echo 'fn main() {}' > src/main.rs && cargo build --release && rm -rf src
COPY sidecar/src ./src
RUN touch src/main.rs && cargo build --release

# Stage 2: build the Go binary (index.* and sqlc/schema.sql are go:embed'd into it).
FROM golang:1.26-bookworm AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/main .

# Stage 3: runtime with both binaries side by side.
FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /out/main /app/main
COPY --from=sidecar /build/target/release/sc2json /app/sc2json
ENTRYPOINT ["/app/main"]
