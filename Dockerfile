# syntax=docker/dockerfile:1

ARG BUILDPLATFORM

FROM --platform=$BUILDPLATFORM node:26.7.0-alpine AS web-build
WORKDIR /src/web

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN npm install --global pnpm@10.23.0 && pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm run build

FROM --platform=$BUILDPLATFORM golang:1.26.6-alpine AS server-build
WORKDIR /src/server

ARG TARGETOS
ARG TARGETARCH

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /out/jobmatcha ./cmd/api

FROM alpine:3.22 AS runtime
RUN addgroup -S jobmatcha && adduser -S -G jobmatcha -u 10001 jobmatcha \
    && mkdir -p /app/web/dist/client /data \
    && chown -R jobmatcha:jobmatcha /app /data

WORKDIR /app
COPY --from=server-build --chown=jobmatcha:jobmatcha /out/jobmatcha ./jobmatcha
COPY --from=web-build --chown=jobmatcha:jobmatcha /src/web/dist/client ./web/dist/client

ENV SERVER_PORT=8181 \
    DB_PATH=/data/app.db \
    STATIC_DIR=/app/web/dist/client

EXPOSE 8181
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget -q -O /dev/null http://127.0.0.1:8181/api/health || exit 1

USER jobmatcha
ENTRYPOINT ["/app/jobmatcha"]
