FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci --no-audit --no-fund
COPY web/ ./
RUN npm run build

FROM rust:1-alpine AS rust
RUN apk add --no-cache musl-dev
WORKDIR /src
COPY search-engine/ search-engine/
COPY fraud-engine/ fraud-engine/
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/src/search-engine/target \
    --mount=type=cache,target=/src/fraud-engine/target \
    cd search-engine && cargo build --release --locked && cp target/release/search-engine /search-engine && \
    cd ../fraud-engine && cargo build --release --locked && cp target/release/fraud-engine /fraud-engine

FROM golang:1.26-alpine AS go
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
COPY --from=web /web/dist ./web/dist
ARG VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /mercatus ./cmd/mercatus

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go /mercatus /usr/local/bin/mercatus
COPY --from=rust /search-engine /usr/local/bin/search-engine
COPY --from=rust /fraud-engine /usr/local/bin/fraud-engine
USER 65532:65532
EXPOSE 8080 9090 8092 8093
ENTRYPOINT ["/usr/local/bin/mercatus"]
CMD ["serve"]
