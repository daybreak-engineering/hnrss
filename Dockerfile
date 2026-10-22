# syntax=docker/dockerfile:1

# Builder runs natively and cross-compiles for the target platform.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY *.go ./

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS="$TARGETOS" GOARCH="$TARGETARCH" \
    go build -trimpath \
      -ldflags "-s -w -X main.buildString=${VERSION}" \
      -o /out/hnrss .

# Static base with CA certs; runs as uid 65532.
FROM gcr.io/distroless/static-debian12:nonroot

ARG VERSION=dev
LABEL org.opencontainers.image.title="hnrss" \
      org.opencontainers.image.description="Custom, realtime Hacker News RSS feeds" \
      org.opencontainers.image.version="${VERSION}"

COPY --from=build /out/hnrss /hnrss

ENV GIN_MODE=release \
    HNRSS_BIND=0.0.0.0:9000 \
    HNRSS_SITE_URL=https://hnrss.org

EXPOSE 9000
USER nonroot:nonroot
ENTRYPOINT ["/hnrss"]
